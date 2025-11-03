FROM public.ecr.aws/docker/library/golang:alpine AS builder

WORKDIR /app
RUN apk add --no-cache git 

COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOPRIVATE=github.com/PierreStephaneVoltaire/powerlifting-coach-app/*

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

FROM public.ecr.aws/docker/library/alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8086

CMD ["./main"]