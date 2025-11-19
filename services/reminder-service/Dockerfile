FROM public.ecr.aws/docker/library/golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build arguments for multi-arch support
ARG TARGETARCH
ARG TARGETOS=linux

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o reminder-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/reminder-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8086

CMD ["./reminder-service"]
