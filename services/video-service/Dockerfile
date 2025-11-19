FROM public.ecr.aws/docker/library/golang:alpine AS builder

WORKDIR /app
RUN apk add --no-cache git 

# Install FFmpeg
RUN apk add --no-cache ffmpeg

COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOPRIVATE=github.com/PierreStephaneVoltaire/powerlifting-coach-app/*

RUN go mod download

COPY . .

ARG TARGETARCH
ARG TARGETOS=linux

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o main cmd/main.go

FROM public.ecr.aws/docker/library/alpine:latest

# Install FFmpeg in final image
RUN apk --no-cache add ca-certificates ffmpeg
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8082

CMD ["./main"]