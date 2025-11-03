module github.com/powerlifting-coach-app/video-service

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/lib/pq v1.10.9
	github.com/golang-migrate/migrate/v4 v4.16.2
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.31.0
	github.com/stretchr/testify v1.8.4
	github.com/google/uuid v1.3.0
	github.com/aws/aws-sdk-go v1.45.0
	github.com/streadway/amqp v1.1.0
	github.com/powerlifting-coach-app/shared v0.0.0
	github.com/h2non/filetype v1.1.3
	github.com/disintegration/imaging v1.6.2
)

replace github.com/powerlifting-coach-app/shared => ../../shared