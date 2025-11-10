package processing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/powerlifting-coach-app/media-processor-service/internal/config"
	"github.com/rs/zerolog/log"
)

type VideoProcessor struct {
	config *config.Config
}

type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	CodecType        string `json:"codec_type"`
	CodecName        string `json:"codec_name"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	RFrameRate      string `json:"r_frame_rate"`
	BitRate         string `json:"bit_rate"`
	Duration        string `json:"duration"`
}

type Format struct {
	Duration string `json:"duration"`
	Size     string `json:"size"`
	BitRate  string `json:"bit_rate"`
}

func NewVideoProcessor(cfg *config.Config) *VideoProcessor {
	return &VideoProcessor{
		config: cfg,
	}
}

func (vp *VideoProcessor) ProcessVideo(inputPath, outputPath string) (*VideoMetadata, error) {
	log.Info().Str("input", inputPath).Str("output", outputPath).Msg("Starting video processing")

	metadata, err := vp.extractMetadata(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	if err := vp.compressVideo(inputPath, outputPath); err != nil {
		return nil, fmt.Errorf("failed to compress video: %w", err)
	}

	log.Info().Str("output", outputPath).Msg("Video processing completed")
	return metadata, nil
}

func (vp *VideoProcessor) GenerateThumbnail(inputPath, outputPath string) error {
	log.Info().Str("input", inputPath).Str("output", outputPath).Msg("Generating thumbnail")

	cmd := exec.Command(vp.config.FFmpegPath,
		"-i", inputPath,
		"-ss", "00:00:01", // Extract frame at 1 second
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:-1", vp.config.ThumbnailSize),
		"-y", // Overwrite output files
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg thumbnail generation failed: %w, stderr: %s", err, stderr.String())
	}

	if err := vp.optimizeThumbnail(outputPath); err != nil {
		log.Warn().Err(err).Msg("Failed to optimize thumbnail, using original")
	}

	log.Info().Str("output", outputPath).Msg("Thumbnail generation completed")
	return nil
}

func (vp *VideoProcessor) extractMetadata(inputPath string) (*VideoMetadata, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	metadata := &VideoMetadata{}

	// Find video stream
	for _, stream := range probe.Streams {
		if stream.CodecType == "video" {
			metadata.Width = stream.Width
			metadata.Height = stream.Height
			metadata.Codec = stream.CodecName
			metadata.Resolution = fmt.Sprintf("%dx%d", stream.Width, stream.Height)

			if stream.BitRate != "" {
				if bitrate, err := strconv.Atoi(stream.BitRate); err == nil {
					metadata.Bitrate = bitrate
				}
			}

			if stream.RFrameRate != "" {
				if parts := strings.Split(stream.RFrameRate, "/"); len(parts) == 2 {
					if num, err := strconv.ParseFloat(parts[0], 64); err == nil {
						if den, err := strconv.ParseFloat(parts[1], 64); err == nil && den != 0 {
							metadata.Framerate = num / den
						}
					}
				}
			}
			break
		}
	}

	return metadata, nil
}

func (vp *VideoProcessor) compressVideo(inputPath, outputPath string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// FFmpeg command for video compression
	cmd := exec.Command(vp.config.FFmpegPath,
		"-i", inputPath,
		"-c:v", "libx264",           // Video codec
		"-preset", "medium",         // Encoding speed vs compression efficiency
		"-crf", "23",               // Constant Rate Factor (quality)
		"-c:a", "aac",              // Audio codec
		"-b:a", "128k",             // Audio bitrate
		"-movflags", "+faststart",   // Web optimization
		"-vf", "scale='min(1920,iw)':-2", // Scale down if larger than 1920px width
		"-y", // Overwrite output files
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Info().Str("cmd", cmd.String()).Msg("Running FFmpeg compression")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg compression failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

func (vp *VideoProcessor) optimizeThumbnail(imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := imaging.Decode(file)
	if err != nil {
		return err
	}

	// Resize to thumbnail size maintaining aspect ratio
	img = imaging.Resize(img, vp.config.ThumbnailSize, 0, imaging.Lanczos)

	// Create optimized JPEG
	outputFile, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 85})
}

func (vp *VideoProcessor) GetVideoDuration(inputPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration failed: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

func (vp *VideoProcessor) ValidateVideo(inputPath string) error {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-count_packets",
		"-show_entries", "stream=nb_read_packets",
		"-of", "csv=p=0",
		inputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("video validation failed: %w", err)
	}

	return nil
}