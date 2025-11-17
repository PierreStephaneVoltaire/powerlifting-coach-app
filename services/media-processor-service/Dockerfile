FROM public.ecr.aws/docker/library/golang:1.23-alpine AS builder

# Install FFmpeg and build dependencies
RUN apk add --no-cache git ffmpeg ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o media-processor ./cmd/main.go

FROM public.ecr.aws/docker/library/alpine:latest

# Install FFmpeg and ca-certificates
RUN apk --no-cache add ffmpeg ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/media-processor .

# Create temp directory for video processing
RUN mkdir -p /tmp/videos

# Run as non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app /tmp/videos

USER appuser

CMD ["./media-processor"]
