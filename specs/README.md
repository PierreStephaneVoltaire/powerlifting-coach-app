# OpenAPI Specifications

This directory contains OpenAPI 3.0 specifications for all service REST endpoints.

## Services

### notification-service
- `notification-service.yaml` - Event ingress and notification management

### settings-service
- `settings-service.yaml` - User settings and preferences

### feed-service
- `feed-service.yaml` - Social feed and posts

### media-service
- `media-service.yaml` - Media upload and presigned URLs

### comments-service
- `comments-service.yaml` - Comments and interactions

### program-service
- `program-service.yaml` - Training programs and workouts

### dm-service
- `dm-service.yaml` - Direct messaging

### machine-service
- `machine-service.yaml` - Machine notes and equipment

### reminder-service
- `reminder-service.yaml` - Workout reminders

### ai-agent-service
- `ai-agent-service.yaml` - AI coach interactions

## Conventions

All specs should include:
- OpenAPI version 3.0
- Service info (title, version, description)
- Server URLs (development, production)
- Security schemes (Bearer JWT)
- Request/response schemas
- Error responses (400, 401, 403, 404, 500)
- Health check endpoint (/health)
