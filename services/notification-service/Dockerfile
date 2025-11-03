FROM public.ecr.aws/docker/library/golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY ../../shared ./shared
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

FROM public.ecr.aws/docker/library/alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8086

CMD ["./main"]