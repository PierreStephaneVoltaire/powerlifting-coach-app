package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/powerlifting-coach-app/media-processor-service/internal/config"
	"github.com/powerlifting-coach-app/media-processor-service/internal/processing"
	"github.com/powerlifting-coach-app/media-processor-service/internal/queue"
	"github.com/powerlifting-coach-app/media-processor-service/internal/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/utils"
)

type ProcessVideoMessage struct {
	VideoID  string `json:"video_id"`
	UserID   string `json:"user_id"`
	Filename string `json:"filename"`
}

type VideoMetadataMessage struct {
	VideoID      string                 `json:"video_id"`
	ProcessedURL string                 `json:"processed_url"`
	ThumbnailURL string                 `json:"thumbnail_url"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func main() {
	godotenv.Load()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENVIRONMENT") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cfg := config.Load()
	log.Info().Msg("Starting media processor service")

	// Initialize storage client
	spacesClient, err := storage.NewSpacesClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Spaces client")
	}

	// Initialize video processor
	processor := processing.NewVideoProcessor(cfg)

	// Initialize queue client
	queueClient, err := queue.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer queueClient.Close()

	// Start consuming from processing queue
	log.Info().Msg("Starting to consume from video.processing queue")

	msgs, err := queueClient.ConsumeVideoProcessing()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start consuming")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Process messages
	go func() {
		for msg := range msgs {
			if err := processMessage(ctx, msg, processor, spacesClient, queueClient); err != nil {
				retryCount := utils.GetRetryCount(msg)
				log.Error().Err(err).Int("retry_count", retryCount).Msg("Failed to process message")

				// Handle failure with retry logic
				if handleErr := utils.HandleMessageFailure(queueClient.GetChannel(), msg, queue.VideoProcessingExchange, "process"); handleErr != nil {
					log.Error().Err(handleErr).Msg("Failed to handle message failure")
					// Fallback to simple nack without requeue to avoid infinite loops
					msg.Nack(false, false)
				}
			} else {
				msg.Ack(false)
			}
		}
	}()

	log.Info().Msg("Media processor service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down media processor service...")
	cancel()
	log.Info().Msg("Service exited")
}

func processMessage(
	ctx context.Context,
	msg amqp.Delivery,
	processor *processing.VideoProcessor,
	spacesClient *storage.SpacesClient,
	queueClient *queue.RabbitMQClient,
) error {
	var processMsg ProcessVideoMessage
	if err := json.Unmarshal(msg.Body, &processMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Info().
		Str("video_id", processMsg.VideoID).
		Str("user_id", processMsg.UserID).
		Str("filename", processMsg.Filename).
		Msg("Processing video")

	// Download original video from S3
	originalKey := fmt.Sprintf("originals/%s/%s", processMsg.UserID, processMsg.Filename)
	localOriginalPath := fmt.Sprintf("/tmp/videos/%s_%s", processMsg.VideoID, processMsg.Filename)

	if err := spacesClient.DownloadFile(originalKey, localOriginalPath); err != nil {
		return fmt.Errorf("failed to download original: %w", err)
	}
	defer os.Remove(localOriginalPath)

	// Process video
	processedPath := fmt.Sprintf("/tmp/videos/%s_processed.mp4", processMsg.VideoID)
	metadata, err := processor.ProcessVideo(localOriginalPath, processedPath)
	if err != nil {
		return fmt.Errorf("failed to process video: %w", err)
	}
	defer os.Remove(processedPath)

	// Generate thumbnail
	thumbnailPath := fmt.Sprintf("/tmp/videos/%s_thumb.jpg", processMsg.VideoID)
	if err := processor.GenerateThumbnail(localOriginalPath, thumbnailPath); err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}
	defer os.Remove(thumbnailPath)

	// Upload processed video to /feed/
	processedKey := fmt.Sprintf("feed/%s/%s", processMsg.UserID, processMsg.Filename)
	if err := spacesClient.UploadFileFromPath(processedKey, processedPath, "video/mp4"); err != nil {
		return fmt.Errorf("failed to upload processed video: %w", err)
	}

	// Upload thumbnail to /thumbnails/
	thumbnailKey := fmt.Sprintf("thumbnails/%s/%s.jpg", processMsg.UserID, processMsg.VideoID)
	if err := spacesClient.UploadFileFromPath(thumbnailKey, thumbnailPath, "image/jpeg"); err != nil {
		return fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// Get public URLs
	processedURL := spacesClient.GetPublicFileURL(processedKey)
	thumbnailURL := spacesClient.GetPublicFileURL(thumbnailKey)

	// Send metadata back to video service
	metadataMsg := VideoMetadataMessage{
		VideoID:      processMsg.VideoID,
		ProcessedURL: processedURL,
		ThumbnailURL: thumbnailURL,
		Metadata: map[string]interface{}{
			"width":      metadata.Width,
			"height":     metadata.Height,
			"resolution": metadata.Resolution,
			"codec":      metadata.Codec,
			"bitrate":    metadata.Bitrate,
			"framerate":  metadata.Framerate,
		},
	}

	if err := queueClient.PublishVideoMetadata(metadataMsg); err != nil {
		return fmt.Errorf("failed to publish metadata: %w", err)
	}

	log.Info().
		Str("video_id", processMsg.VideoID).
		Str("processed_url", processedURL).
		Str("thumbnail_url", thumbnailURL).
		Msg("Video processing completed")

	return nil
}
