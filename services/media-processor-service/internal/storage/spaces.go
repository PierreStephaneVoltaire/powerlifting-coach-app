package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/powerlifting-coach-app/media-processor-service/internal/config"
)

type SpacesClient struct {
	client     *s3.Client
	bucketName string
	cdnURL     string
	region     string
}

func NewSpacesClient(appCfg *cfg.Config) (*SpacesClient, error) {
	// Create AWS credentials
	creds := credentials.NewStaticCredentialsProvider(
		appCfg.SpacesAccessKey,
		appCfg.SpacesSecretKey,
		"",
	)

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(appCfg.SpacesRegion),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	return &SpacesClient{
		client:     client,
		bucketName: appCfg.SpacesBucket,
		cdnURL:     appCfg.CDNUrl,
		region:     appCfg.SpacesRegion,
	}, nil
}

func (s *SpacesClient) DownloadFile(key string, localPath string) error {
	ctx := context.Background()

	// Create directory if it doesn't exist
	dir := strings.TrimSuffix(localPath, "/"+key)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	downloader := manager.NewDownloader(s.client)
	_, err = downloader.Download(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func (s *SpacesClient) UploadFileFromPath(key string, localPath string, contentType string) error {
	ctx := context.Background()

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	uploader := manager.NewUploader(s.client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *SpacesClient) UploadFile(key string, body io.Reader, contentType string) error {
	ctx := context.Background()

	uploader := manager.NewUploader(s.client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *SpacesClient) GetPublicFileURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cdnURL, "/"), key)
	}

	// Public URL (bucket has public read access)
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)
}
