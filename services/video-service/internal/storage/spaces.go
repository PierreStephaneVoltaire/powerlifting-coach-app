package storage

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/powerlifting-coach-app/video-service/internal/config"
)

type SpacesClient struct {
	client *s3.S3
	bucket string
	cdnURL string
}

func NewSpacesClient(cfg *config.Config) (*SpacesClient, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.SpacesRegion),
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

	return &SpacesClient{
		client: client,
		bucket: cfg.SpacesBucket,
		cdnURL: cfg.CDNUrl,
	}, nil
}

func (s *SpacesClient) GeneratePresignedUploadURL(key string, contentType string, duration time.Duration) (string, error) {
	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         aws.String("private"),
	})

	url, err := req.Presign(duration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

func (s *SpacesClient) UploadFile(key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        aws.ReadSeekCloser(body),
		ContentType: aws.String(contentType),
		ACL:         aws.String("private"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *SpacesClient) UploadPublicFile(key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        aws.ReadSeekCloser(body),
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload public file: %w", err)
	}

	return nil
}

func (s *SpacesClient) GetFileURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cdnURL, "/"), key)
	}

	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	url, _ := req.Presign(time.Hour * 24) // 24 hour expiry for private files
	return url
}

func (s *SpacesClient) GetPublicFileURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cdnURL, "/"), key)
	}

	return fmt.Sprintf("https://%s.%s/%s", s.bucket,
		strings.TrimPrefix(*s.client.Config.Endpoint, "https://"), key)
}

func (s *SpacesClient) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *SpacesClient) FileExists(key string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "Not Found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}

	return true, nil
}

func (s *SpacesClient) GetFileSize(key string) (int64, error) {
	result, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return *result.ContentLength, nil
}