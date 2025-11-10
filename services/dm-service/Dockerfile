FROM public.ecr.aws/docker/library/golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o dm-service ./cmd/main.go

FROM public.ecr.aws/docker/library/alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/dm-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8084

CMD ["./dm-service"]
