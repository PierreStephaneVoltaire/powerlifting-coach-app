package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/powerlifting-coach-app/video-service/internal/config"
)

type SpacesClient struct {
	client         *azblob.Client
	containerName  string
	cdnURL         string
	accountName    string
	credential     *azblob.SharedKeyCredential
}

func NewSpacesClient(cfg *config.Config) (*SpacesClient, error) {
	// Create credential
	credential, err := azblob.NewSharedKeyCredential(cfg.SpacesAccessKey, cfg.SpacesSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Create blob service client
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.SpacesAccessKey)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure blob client: %w", err)
	}

	return &SpacesClient{
		client:        client,
		containerName: cfg.SpacesBucket,
		cdnURL:        cfg.CDNUrl,
		accountName:   cfg.SpacesAccessKey,
		credential:    credential,
	}, nil
}

func (s *SpacesClient) GeneratePresignedUploadURL(key string, contentType string, duration time.Duration) (string, error) {
	ctx := context.Background()

	// Create SAS URL for upload
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.containerName, key)

	// Set SAS permissions for upload
	permissions := sas.BlobPermissions{
		Read:   false,
		Add:    false,
		Create: true,
		Write:  true,
		Delete: false,
	}

	startTime := time.Now().UTC().Add(-10 * time.Minute)
	expiryTime := time.Now().UTC().Add(duration)

	sasQueryParams, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     startTime,
		ExpiryTime:    expiryTime,
		Permissions:   permissions.String(),
		ContainerName: s.containerName,
		BlobName:      key,
	}.SignWithSharedKey(s.credential)

	if err != nil {
		return "", fmt.Errorf("failed to generate SAS token: %w", err)
	}

	sasURL := fmt.Sprintf("%s?%s", blobURL, sasQueryParams.Encode())
	return sasURL, nil
}

func (s *SpacesClient) UploadFile(key string, body io.Reader, contentType string) error {
	ctx := context.Background()

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	// Upload with private access (no public read)
	_, err := blobClient.UploadStream(ctx, body, &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
		AccessTier: to.Ptr(blob.AccessTierHot),
	})

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *SpacesClient) UploadPublicFile(key string, body io.Reader, contentType string) error {
	ctx := context.Background()

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	// Upload blob (container already has public read access)
	_, err := blobClient.UploadStream(ctx, body, &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
		AccessTier: to.Ptr(blob.AccessTierHot),
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

	// Generate SAS URL for private access
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.containerName, key)

	permissions := sas.BlobPermissions{Read: true}
	startTime := time.Now().UTC().Add(-10 * time.Minute)
	expiryTime := time.Now().UTC().Add(24 * time.Hour)

	sasQueryParams, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     startTime,
		ExpiryTime:    expiryTime,
		Permissions:   permissions.String(),
		ContainerName: s.containerName,
		BlobName:      key,
	}.SignWithSharedKey(s.credential)

	if err != nil {
		// Fallback to public URL if SAS generation fails
		return blobURL
	}

	return fmt.Sprintf("%s?%s", blobURL, sasQueryParams.Encode())
}

func (s *SpacesClient) GetPublicFileURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cdnURL, "/"), key)
	}

	// Public URL (container has public read access)
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.containerName, key)
}

func (s *SpacesClient) DeleteFile(key string) error {
	ctx := context.Background()

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	_, err := blobClient.Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *SpacesClient) FileExists(key string) (bool, error) {
	ctx := context.Background()

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	_, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		// Check if error is "not found"
		if strings.Contains(err.Error(), "BlobNotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}

	return true, nil
}

func (s *SpacesClient) GetFileSize(key string) (int64, error) {
	ctx := context.Background()

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return *props.ContentLength, nil
}
