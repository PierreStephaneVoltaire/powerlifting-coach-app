package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/powerlifting-coach-app/video-service/internal/config"
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

func (s *SpacesClient) GeneratePresignedUploadURL(key string, contentType string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(duration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return request.URL, nil
}

func (s *SpacesClient) UploadFile(key string, body io.Reader, contentType string) error {
	ctx := context.Background()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
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

func (s *SpacesClient) UploadPublicFile(key string, body io.Reader, contentType string) error {
	// S3 bucket has public read policy, so this is the same as UploadFile
	return s.UploadFile(key, body, contentType)
}

func (s *SpacesClient) GetFileURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cdnURL, "/"), key)
	}

	// Generate presigned URL for private access
	presignClient := s3.NewPresignClient(s.client)
	request, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(24*time.Hour))

	if err != nil {
		// Fallback to public URL if presign fails
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)
	}

	return request.URL
}

func (s *SpacesClient) GetPublicFileURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cdnURL, "/"), key)
	}

	// Public URL (bucket has public read access)
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)
}

func (s *SpacesClient) DeleteFile(key string) error {
	ctx := context.Background()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *SpacesClient) FileExists(key string) (bool, error) {
	ctx := context.Background()

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if error is "not found"
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}

	return true, nil
}

func (s *SpacesClient) GetFileSize(key string) (int64, error) {
	ctx := context.Background()

	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	if result.ContentLength == nil {
		return 0, fmt.Errorf("content length is nil")
	}

	return *result.ContentLength, nil
}
