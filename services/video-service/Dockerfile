FROM public.ecr.aws/docker/library/golang:1.21-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git 

# Install FFmpeg
RUN apk add --no-cache ffmpeg

COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

FROM public.ecr.aws/docker/library/alpine:latest

# Install FFmpeg in final image
RUN apk --no-cache add ca-certificates ffmpeg
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8082

CMD ["./main"]