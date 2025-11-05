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
