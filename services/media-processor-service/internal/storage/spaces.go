package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/lazylifts/media-processor-service/internal/config"
)

type SpacesClient struct {
	client        *azblob.Client
	containerName string
	cdnURL        string
	accountName   string
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

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	// Download blob to file
	_, err = blobClient.DownloadFile(ctx, file, nil)
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

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	// Upload blob with public access (container already has public read)
	_, err = blobClient.UploadFile(ctx, file, &azblob.UploadFileOptions{
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

func (s *SpacesClient) UploadFile(key string, body io.Reader, contentType string) error {
	ctx := context.Background()

	blobClient := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlockBlobClient(key)

	// Upload blob with public access (container already has public read)
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

func (s *SpacesClient) GetPublicFileURL(key string) string {
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.cdnURL, "/"), key)
	}

	// Public URL (container has public read access)
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.containerName, key)
}
