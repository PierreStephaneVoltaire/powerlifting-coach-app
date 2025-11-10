package storage

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/powerlifting-coach-app/media-processor-service/internal/config"
)

type SpacesClient struct {
	client     *s3.S3
	downloader *s3manager.Downloader
	uploader   *s3manager.Uploader
	bucket     string
	cdnURL     string
}

func NewSpacesClient(cfg *config.Config) (*SpacesClient, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(cfg.SpacesRegion),
		Endpoint: aws.String(cfg.SpacesEndpoint),
		Credentials: credentials.NewStaticCredentials(
			cfg.SpacesAccessKey,
			cfg.SpacesSecretKey,
			"",
		),
		S3ForcePathStyle: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)
	downloader := s3manager.NewDownloader(sess)
	uploader := s3manager.NewUploader(sess)

	return &SpacesClient{
		client:     client,
		downloader: downloader,
		uploader:   uploader,
		bucket:     cfg.SpacesBucket,
		cdnURL:     cfg.CDNUrl,
	}, nil
}

func (s *SpacesClient) DownloadFile(key string, localPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(strings.TrimSuffix(localPath, "/"+key), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = s.downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func (s *SpacesClient) UploadFileFromPath(key string, localPath string, contentType string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *SpacesClient) UploadFile(key string, body io.Reader, contentType string) error {
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
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

	return fmt.Sprintf("https://%s.%s/%s", s.bucket,
		strings.TrimPrefix(*s.client.Config.Endpoint, "https://"), key)
}
