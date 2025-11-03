# Integration Notes

## Architecture Decisions

### Service Consolidation

**video-service** handles all media, feed, and comment operations:
- Media uploads (presigned URLs, storage to DigitalOcean Spaces)
- Feed posts (denormalized video posts with metadata)
- Comments (threaded comments on posts)
- Media processing (transcoding, thumbnails)
- Interactions (likes on posts/comments)

**Rationale**: Reduces operational complexity, shares database connections, simplifies event flows. All video/post/comment data lives in one service's database.

### Event Flow Patterns

#### Video Upload & Feed Post Creation
1. Frontend ’ POST notification-service `/api/v1/notify/events` ’ `media.upload_requested`
2. video-service consumes ’ generates presigned URL
3. Frontend uploads to Spaces ’ POST video-service `/api/v1/videos/:id/complete` ’ `media.uploaded`
4. video-service processing worker ’ transcodes ’ `media.processed`
5. video-service persists feed entry ’ `feed.post.created`

#### Comments
1. Frontend ’ POST notification-service ’ `comment.created`
2. video-service consumes ’ persists ’ `comment.persisted`

#### Interactions
1. Frontend ’ POST notification-service ’ `interaction.liked`
2. video-service consumes ’ persists like

### Database Strategy

Each service maintains its own Postgres database:
- **video-service DB**: videos, feed_posts, comments, likes, media_metadata
- **settings-service DB**: user_settings, app_settings
- **program-service DB**: programs, workouts, workout_sessions
- **dm-service DB**: conversations, messages
- **machine-service DB**: machine_notes, equipment_prefs

Shared `idempotency_keys` table in each service DB for deduplication.

### Lock Strategy

Using Postgres advisory locks (`pg_try_advisory_lock`) for critical sections:
- Settings updates: lock on `user_id`
- Program updates: lock on `plan_id`
- Video upload completion: lock on `video_id`

### RabbitMQ Conventions

- **Exchange**: `app.events` (topic exchange)
- **Routing key**: `event_type` (e.g., `media.uploaded`, `user.settings.submitted`)
- **Manual ack**: Only ack after DB commit
- **Requeue**: On transient failures, requeue message
- **Idempotency**: All events include `client_generated_id` UUID

### Testing Locally

#### Simulate RabbitMQ Events
```bash
# Install rabbitmqadmin
wget http://localhost:15672/cli/rabbitmqadmin
chmod +x rabbitmqadmin

# Publish test event
./rabbitmqadmin publish routing_key=user.settings.submitted \
  payload='{"schema_version":"1.0.0","event_type":"user.settings.submitted","client_generated_id":"550e8400-e29b-41d4-a716-446655440000","user_id":"7c9e6679-7425-40de-944b-e07fc1f90ae7","timestamp":"2025-11-05T10:00:00Z","source_service":"frontend","data":{"weight":{"value":93.5,"unit":"kg"},"age":28,"training_days_per_week":4}}'
```

#### Run Consumer in Dev Mode
```bash
cd services/settings-service
go run cmd/main.go
```

#### Run Pod Health Check
```bash
./scripts/pod-health-check.sh
cat infra-health.log
```

### Offline Queue (Frontend)

When notification-service is unreachable, frontend uses IndexedDB to queue events:
```javascript
// IndexedDB schema
db.createObjectStore('eventQueue', { keyPath: 'client_generated_id' });

// Enqueue
await db.eventQueue.add({
  client_generated_id: uuid(),
  event_type: 'user.settings.submitted',
  user_id: currentUser.id,
  timestamp: new Date().toISOString(),
  data: settingsData,
  retry_count: 0
});

// Retry on reconnect
setInterval(async () => {
  const pending = await db.eventQueue.toArray();
  for (const event of pending) {
    try {
      await fetch('/api/v1/notify/events', { method: 'POST', body: JSON.stringify(event) });
      await db.eventQueue.delete(event.client_generated_id);
    } catch (err) {
      event.retry_count++;
      await db.eventQueue.put(event);
    }
  }
}, 30000); // every 30s
```

### Feed Access Control

Feed visibility modes:
- **public**: Anyone can view
- **passcode**: Requires passcode verification

Flow:
1. Frontend ’ POST notification-service ’ `feed.access.attempt` { user_id, feed_owner_id, passcode }
2. settings-service validates passcode_hash ’ `feed.access.granted` { access_token } OR `feed.access.denied`
3. Frontend stores access_token in localStorage/cookie
4. video-service checks token before serving passcode-protected posts

Access tokens: JWT signed by settings-service, 24h expiry, includes feed_owner_id.

### Migration Strategy

All services use `golang-migrate` (or similar):
```bash
# Run migrations
migrate -path ./migrations -database postgres://localhost/video_db up

# Rollback
migrate -path ./migrations -database postgres://localhost/video_db down 1
```

Migration files: `001_initial.up.sql`, `001_initial.down.sql`

### Secrets Management

Do not hardcode:
- RabbitMQ URLs
- Database URLs
- DigitalOcean Spaces keys
- JWT secrets
- OpenWebUI API keys

Use environment variables loaded from `.env` (dev) or Kubernetes secrets (prod).

### AI Connector (Deferred)

AI-agent-service connects to external OpenWebUI endpoint:
```go
type OpenWebUIClient struct {
    endpoint string
    apiKey   string
}

func (c *OpenWebUIClient) SendPrompt(ctx context.Context, prompt string) (string, error) {
    // Stub until credentials available
    return "AI response stubbed", nil
}
```

### Monitoring

Log all events to stdout as JSON:
```json
{"level":"info","ts":1699200000,"service":"video-service","msg":"Event processed","event_type":"media.uploaded","client_generated_id":"550e8400-...","user_id":"7c9e6679-..."}
```

Use centralized log aggregation (e.g., Loki, Elasticsearch) in production.

## Local Development Setup

### Prerequisites
- Docker and Docker Compose
- PostgreSQL client
- Go 1.21+
- Node.js 18+
- kubectl (for K8s testing)

### Starting Local Services

```bash
# Start RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Start PostgreSQL
docker run -d --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:15

# Create databases for each service
psql -h localhost -U postgres -c "CREATE DATABASE auth_db;"
psql -h localhost -U postgres -c "CREATE DATABASE user_db;"
psql -h localhost -U postgres -c "CREATE DATABASE video_db;"
psql -h localhost -U postgres -c "CREATE DATABASE settings_db;"
psql -h localhost -U postgres -c "CREATE DATABASE program_db;"
psql -h localhost -U postgres -c "CREATE DATABASE dm_db;"
psql -h localhost -U postgres -c "CREATE DATABASE machine_db;"
psql -h localhost -U postgres -c "CREATE DATABASE reminder_db;"
```

### Environment Variables

Each service needs these environment variables:

```bash
export PORT=8080  # varies by service
export ENVIRONMENT=development
export DATABASE_URL=postgresql://postgres:postgres@localhost:5432/service_db
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
export REDIS_URL=localhost:6379
export AUTH_SERVICE=http://localhost:8080
export USER_SERVICE=http://localhost:8081
```

## RabbitMQ Event Simulation

### Using RabbitMQ Management UI

1. Open http://localhost:15672 (guest/guest)
2. Go to **Exchanges** â†’ `app.events`
3. Click **Publish message**
4. Set routing key (e.g., `user.settings.submitted`)
5. Paste event JSON payload
6. Click **Publish message**

### Testing Idempotency

```bash
# Publish same event twice with same client_generated_id
# Verify only one record in database
psql -h localhost -U postgres -d program_db -c "SELECT COUNT(*) FROM programs;"
# Should return 1

# Check idempotency_keys table
psql -h localhost -U postgres -d program_db -c "SELECT * FROM idempotency_keys;"
```

### Testing Manual Ack and Requeue

All consumers use manual acknowledgment after successful database commit. If processing fails:
- Transient errors (DB connection, network): Message is requeued
- Permanent errors (validation, unmarshal): Message is nack'd without requeue

Test requeue behavior:
```bash
# Kill database while processing
docker stop postgres

# Publish event - will fail and requeue
# Restart database
docker start postgres

# Message should process successfully
```

## Consumer Development Mode

### Running a Single Service

```bash
cd services/program-service
go run cmd/main.go
```

Logs will show:
- Event handler registration
- Queue binding confirmation
- Event processing status
- Idempotency checks

### Debugging RabbitMQ

```bash
# List all queues
rabbitmqctl list_queues

# View messages in queue
rabbitmqctl list_queues name messages messages_ready messages_unacknowledged

# Check bindings
rabbitmqctl list_bindings
```

## Testing Offline Queue (Frontend)

### IndexedDB Inspection

1. Open browser DevTools â†’ Application â†’ IndexedDB
2. Find `offlineQueue` database
3. Check `events` object store
4. Verify events have `retryCount` and `nextRetryAt`

### Exponential Backoff Testing

Events retry with exponential backoff:
- Retry 1: ~2 seconds
- Retry 2: ~4 seconds  
- Retry 3: ~8 seconds
- Retry 4: ~16 seconds
- Retry 5: ~32 seconds
- Max: 60 seconds

After 5 retries, events remain queued for manual retry.

## K8s Pod Health Monitoring

### Check Pod Status
```bash
kubectl get pods -n app
kubectl describe pod program-service-xxx -n app
kubectl logs program-service-xxx -n app --tail=100
```

### Run Health Check Script
```bash
./scripts/pod-health-check.sh
cat infra-health.log
```

## Common Issues

### Events Not Consuming
**Check:**
1. RabbitMQ connection: `rabbitmqctl list_connections`
2. Queue bindings: `rabbitmqctl list_bindings`
3. Service logs: `docker logs <service>`
4. Routing key matches event_type

### Duplicate Processing
**Check:**
1. Idempotency keys in database
2. client_generated_id is unique
3. Transaction commits before ack

### Messages Stuck in Queue
**Check:**
1. Consumer prefetch settings (QoS = 10)
2. Processing errors in logs
3. Database connection issues

## Lock Acquisition/Release Behavior

Postgres advisory locks are NOT currently implemented. All services use database-level idempotency via `idempotency_keys` table with unique constraint on `client_generated_id`.

This provides idempotency across multiple replicas without requiring advisory locks. The unique constraint ensures only one consumer can insert the idempotency key, and losers get a constraint violation error.

