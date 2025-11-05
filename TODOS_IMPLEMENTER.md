# TODOS_IMPLEMENTER.md

## Status Legend
- `[ ]` todo
- `[>]` in-progress
- `[x]` done
- `[!]` blocked

---

## Recent Completion Summary

### Frontend Implementation (All Complete)
- **Onboarding Form**: User settings form with all required fields (Story 2)
- **Offline Queue System**: IndexedDB-based queue with exponential backoff and auto-retry (Story 3)
- **Sync Indicator**: Real-time offline status and pending queue count display (Story 3)
- **Feed UI**: Feed list with cursor pagination, like/comment functionality, optimistic updates (Story 4, 6)
- **Comment Section**: Threaded comment display and input (Story 6)
- **Program Planner**: Create training plans with comp date calculation (Story 7)
- **Workout Session**: Start/complete workout with exercise tracking (Story 7, 8)
- **DM UI**: Conversation list and 1:1 chat view with pin attempts (Story 9)
- **Plate Calculator**: Client-side plate calculation with telemetry (Story 11)
- **Machine Notes**: Save equipment settings and preferences (Story 11)
- **Video Upload**: File selection, metadata form, progress indicator (Story 5)
- **Stale Data Banner**: Warning for outdated cache with refresh option (Story 3)
- **Passcode Prompt**: Modal for feed access control with storage options (Story 12)

### Backend Services Implementation

#### Video Service (Enhanced)
- Added EventPublisher interface and PublishEvent method to RabbitMQ client
- Updated comment_handlers.go to emit comment.persisted event after DB commit
- Supports comment.created and interaction.liked event consumers
- HTTP endpoints for feed, comments, likes (Story 4, 6)
- Media upload processing with media_handlers.go (Story 5):
  - HandleMediaUploadRequested: Creates upload records, emits media.uploaded
  - HandleMediaUploaded: Simulates processing, emits media.processed and feed.post.created
  - Migration 004 adds media metadata fields and upload tracking tables

#### Program Service (Enhanced)
- Created event consumer infrastructure (event_consumer.go)
- Implemented program_event_handlers.go with four handlers:
  - HandleProgramPlanCreated: Persists programs with comp_date, training_days_per_week
  - HandleProgramPlanUpdated: Updates existing programs, emits program.plan.persisted
  - HandleWorkoutStarted: Records workout start_time
  - HandleWorkoutCompleted: Stores duration, exercises_summary JSONB, notes
- Created migration 002_add_event_tables.up.sql with programs, workout_sessions, idempotency_keys (Story 7)

#### DM Service (New - Complete)
- Full microservice created from scratch
- Structure: cmd/main.go, config, database, handlers, queue, migrations
- Event consumers for dm.message.sent and dm.pin.attempts
- Auto-creates conversations with normalized participant IDs (participant_1_id < participant_2_id)
- Persists messages and pin attempts to JSONB metadata
- Implements idempotency pattern (Story 9)

#### Machine Service (New - Complete)
- Full microservice created from scratch
- Structure: cmd/main.go, config, database, handlers, queue, migrations
- Event consumer for machine.notes.submitted
- Database schema with machine_type_enum (barbell, hack_squat, leg_press, hex_bar, cable, other)
- Visibility control (public/private)
- Persists brand, model, machine_type, settings
- Implements idempotency pattern
- Emits machine.notes.persisted event (Story 11)

#### Reminder Service (New - Complete)
- Full microservice created from scratch
- Consumes program.plan.persisted events
- Creates reminders at 1 week, 2 weeks, and 4 weeks before competition
- Background scheduler processes pending reminders every minute
- Emits reminder.sent events
- Database schema with reminder_status_enum (pending, sent, cancelled)
- Port 8086 (Story 7)

### Key Technical Patterns Implemented
- **Event-Driven Architecture**: All service communication via RabbitMQ with topic exchange
- **Idempotency**: client_generated_id in all events, idempotency_keys table in each service
- **Event Emissions**: All services emit persisted events after successful DB commits
- **Offline-First Frontend**: IndexedDB queue persistence, exponential backoff (max 5 retries, capped at 60s)
- **Manual Ack with Requeue**: RabbitMQ consumers use manual ack, requeue on transient failure
- **Structured JSON Logging**: zerolog with consistent log patterns across all services
- **Database Migrations**: golang-migrate with up/down migrations for all schema changes
- **JSONB for Flexibility**: exercises_summary, pinned_attempts, metadata stored as JSONB
- **Normalized Relationships**: Conversation participants normalized to ensure uniqueness
- **Time-Based Scheduling**: Background ticker for processing pending reminders
- **Data Freshness Monitoring**: Frontend hook to detect stale cached data (5-minute threshold)

---

## Operational Setup (Story 14, 16, 17)

### Event Schemas & Directories
- [x] Create /events/schemas/ directory structure
- [x] Create /specs/ directory for OpenAPI specs
- [x] Create /ai/templates/ directory for AI prompt templates

### Infrastructure Logging
- [x] Create infra-health.log at repo root
- [x] Add pod health check script in /scripts/pod-health-check.sh

---

## Story 1: Login & Auth Gate

### Events
- [x] Create event schema: auth.user.logged_in.json

### Frontend
- [ ] Add Keycloak OIDC flow integration
- [ ] Add offline/auth-unavailable UI with retry
- [ ] Add JWT storage (secure cookie or localStorage)
- [ ] Emit auth.user.logged_in event on successful login

---

## Story 2: Onboarding Settings Form

### Database & Migrations
- [x] Create migration: settings-service add onboarding fields to user_settings table
- [x] Add fields: weight, age, target_weight_class, weeks_until_comp, squat_goal, bench_goal, dead_goal, most_important_lift, least_important_lift, recovery_rating_squat, recovery_rating_bench, recovery_rating_dead, training_days_per_week, session_length_minutes, weight_plan, form_issues, injuries, evaluate_feasibility, federation, knee_sleeve, deadlift_style, squat_stance, add_per_month, volume_preference, recovers_from_heavy_deads, height, past_competitions, feed_visibility, passcode_hash

### Events
- [x] Create event schema: user.settings.submitted.json
- [x] Create event schema: user.settings.persisted.json
- [x] Create event schema: user.settings.failed.json

### Backend - Notification Service
- [x] Add POST /api/v1/notify/events endpoint to accept user.settings.submitted
- [x] Add RabbitMQ publisher for user.settings.submitted event

### Backend - Settings Service
- [x] Add RabbitMQ consumer for user.settings.submitted
- [x] Add idempotency check using client_generated_id (Postgres advisory lock)
- [x] Add validation logic for settings
- [x] Persist settings to Postgres
- [x] Emit user.settings.persisted on success
- [x] Emit user.settings.failed on validation error

### Frontend
- [x] Create onboarding settings form UI with all required fields
- [x] Add offline queue support (IndexedDB) for settings submission
- [x] Submit settings to Notification service /api/v1/notify/events

### OpenAPI
- [ ] Generate OpenAPI spec for notification-service /api/v1/notify/events endpoint
- [ ] Generate OpenAPI spec for settings-service read endpoints

---

## Story 3: Frontend Resilience

### Frontend
- [x] Add network timeout handling for all API calls
- [x] Add IndexedDB queue for offline writes
- [x] Add exponential backoff retry logic
- [x] Add "sync pending" UI indicator
- [x] Add "data may be out of date" banner for stale cache
- [x] Add queue persistence across reloads
- [x] Create useDataFreshness hook (5-minute threshold)
- [x] Create StaleDataBanner component with refresh option

---

## Story 4: Main Feed (Consolidated into video-service)

### Database & Migrations
- [x] Create migration: video-service add feed_posts table

### Events
- [x] Create event schema: feed.post.created.json
- [x] Create event schema: feed.post.updated.json
- [x] Create event schema: feed.post.deleted.json

### Backend - Video Service (Feed functionality)
- [x] Add RabbitMQ consumer for feed.post.created event
- [x] Add RabbitMQ consumer for feed.post.updated event
- [x] Add RabbitMQ consumer for feed.post.deleted event
- [x] Persist feed entries to Postgres (denormalized)
- [x] Add GET /api/v1/feed endpoint with cursor pagination
- [x] Add GET /api/v1/feed/:post_id endpoint

### Frontend
- [x] Add feed list UI with cursor-based pagination
- [x] Add cached feed fallback when video-service unreachable
- [x] Add feed refresh logic

### OpenAPI
- [ ] Update video-service OpenAPI spec with feed endpoints

---

## Story 5: Video Upload & Metadata (Consolidated into video-service)

### Database & Migrations
- [x] Create migration: video-service add media metadata fields (movement_label, weight, rpe, visibility)
- [x] Create migration 004_add_media_metadata_fields with movement_label_enum and visibility_enum
- [x] Add media_uploads table for tracking upload lifecycle
- [x] Add media_idempotency_keys table for event deduplication

### Events
- [x] Create event schema: media.upload_requested.json
- [x] Create event schema: media.uploaded.json
- [x] Create event schema: media.processed.json

### Backend - Video Service (Media & Processing)
- [x] Add RabbitMQ consumer for media.upload_requested event
- [x] Add idempotency using client_generated_id
- [x] Emit media.uploaded event after presigned upload complete
- [x] Add processing worker consumer for media.uploaded event
- [x] Add transcoding logic (stub initially)
- [x] Add thumbnail generation (stub initially)
- [x] Emit media.processed event with media_url, thumbnail_url
- [x] Emit feed.post.created event after processing
- [x] Created media_handlers.go with HandleMediaUploadRequested and HandleMediaUploaded

### Frontend
- [x] Add video upload UI with progress indicator
- [x] Add metadata form: movement_label, weight, rpe, comment_text, visibility
- [x] Enqueue upload locally if notification-service unreachable
- [ ] Add background upload support
- [ ] Add upload resume on app restart

### OpenAPI
- [ ] Update video-service OpenAPI spec with media endpoints

---

## Story 6: Comments & Interactions (Consolidated into video-service)

### Database & Migrations
- [x] Create migration: video-service add comments and likes tables

### Events
- [x] Create event schema: comment.created.json
- [x] Create event schema: comment.persisted.json
- [x] Create event schema: interaction.liked.json

### Backend - Video Service (Comments & Likes)
- [x] Add RabbitMQ consumer for comment.created event
- [x] Add idempotency using client_generated_id
- [x] Persist comments to Postgres (threaded with parent_comment_id)
- [x] Emit comment.persisted event
- [x] Add consumer for interaction.liked event
- [x] Persist likes to Postgres (deduped by user_id + target_id)
- [x] Add GET /api/v1/posts/:post_id/comments endpoint
- [x] Add GET /api/v1/posts/:post_id/likes endpoint
- [x] Updated comment_handlers.go to emit comment.persisted after successful DB commit

### Frontend
- [x] Add comment input UI on posts
- [x] Add threaded comment display
- [x] Add like/upvote button
- [x] Emit comment.created via Notification service
- [x] Emit interaction.liked via Notification service
- [x] Add cached comments fallback with localStorage

### OpenAPI
- [ ] Update video-service OpenAPI spec with comment/like endpoints

---

## Story 7: Program Planner

### Database & Migrations
- [x] Create migration: program-service create programs table
- [x] Create migration: program-service create workout_sessions table
- [x] Create migration: program-service create idempotency_keys table

### Events
- [x] Create event schema: program.plan.created.json
- [x] Create event schema: program.plan.updated.json
- [x] Create event schema: program.plan.persisted.json
- [x] Create event schema: workout.started.json
- [x] Create event schema: workout.completed.json

### Backend - Program Service
- [x] Add RabbitMQ consumer for program.plan.created event
- [x] Add idempotency using client_generated_id
- [x] Persist programs to Postgres with comp_date and training_days_per_week
- [x] Emit program.plan.persisted event
- [x] Add consumer for workout.started event
- [x] Persist workout start_time to Postgres
- [x] Add consumer for workout.completed event
- [x] Persist workout duration, exercises_summary JSONB, notes to Postgres
- [x] Add RabbitMQ consumer for program.plan.updated event
- [x] HandleProgramPlanUpdated updates existing programs and emits program.plan.persisted

### Backend - Reminder Service (NEW)
- [x] Create new service: reminder-service with Dockerfile and structure
- [x] Create migration: reminder-service create reminders table
- [x] Add consumer for program.plan.persisted to schedule reminders
- [x] Add scheduler to publish reminder.sent events when due
- [x] Create reminders at 1 week, 2 weeks, and 4 weeks before competition
- [ ] Add consumer for program.plan.updated to update reminders

### Frontend
- [x] Add program planner UI (create/edit plan)
- [x] Add workout list UI showing weeks until comp
- [x] Add start workout button
- [x] Add complete workout form with duration and summary
- [x] Emit program.plan.created/updated via Notification service
- [x] Emit workout.started via Notification service
- [x] Emit workout.completed via Notification service

### OpenAPI
- [ ] Generate OpenAPI spec for program-service endpoints
- [ ] Generate OpenAPI spec for reminder-service endpoints

---

## Story 8: Start/Complete Workout UX

(Covered in Story 7 - no additional tasks)

---

## Story 9: DMs - 1:1 Private Chats

### Database & Migrations
- [x] Create new service: dm-service with Dockerfile and structure
- [x] Create migration: dm-service create conversations table
- [x] Create migration: dm-service create messages table
- [x] Create migration: dm-service create idempotency_keys table

### Events
- [x] Create event schema: dm.message.sent.json
- [x] Create event schema: dm.message.persisted.json
- [x] Create event schema: dm.pin.attempts.json

### Backend - DM Service
- [x] Create dm-service cmd/main.go with RabbitMQ consumer
- [x] Add consumer for dm.message.sent event
- [x] Add idempotency using client_generated_id
- [x] Persist messages to Postgres
- [x] Auto-create conversations with normalized participant IDs
- [x] Emit dm.message.persisted event
- [x] Add consumer for dm.pin.attempts event
- [x] Persist pinned attempts to conversation metadata JSONB

### Frontend
- [x] Add DM list UI (conversations)
- [x] Add DM chat UI (1:1 messages)
- [x] Add pin attempts button
- [ ] Add media attachment support (reuse media upload)
- [x] Emit dm.message.sent via Notification service
- [x] Emit dm.pin.attempts via Notification service
- [ ] Add cached messages fallback

### OpenAPI
- [ ] Generate OpenAPI spec for dm-service endpoints

---

## Story 11: Tools - Plate Calculator & Machine Notes

### Database & Migrations
- [x] Create new service: machine-service with Dockerfile and structure
- [x] Create migration: machine-service create machine_notes table with machine_type_enum and visibility_enum
- [x] Create migration: machine-service create idempotency_keys table

### Events
- [x] Create event schema: tools.platecalc.query.json (optional telemetry)
- [x] Create event schema: machine.notes.submitted.json
- [x] Create event schema: machine.notes.persisted.json

### Backend - Machine Service
- [x] Create machine-service cmd/main.go with RabbitMQ consumer
- [x] Add consumer for machine.notes.submitted event
- [x] Add idempotency using client_generated_id
- [x] Persist notes to Postgres with brand, model, machine_type, settings, visibility
- [x] Emit machine.notes.persisted event

### Frontend
- [x] Add plate calculator UI (client-side logic)
- [x] Add machine notes UI (create/edit notes)
- [x] Add plate inventory settings
- [x] Emit machine.notes.submitted via Notification service
- [x] Optionally emit tools.platecalc.query for telemetry

### OpenAPI
- [ ] Generate OpenAPI spec for machine-service endpoints

---

## Story 12: Feed Access Control - Public vs Passcode

### Database & Migrations
(Uses existing settings-service user_settings table with passcode_hash)

### Events
- [x] Create event schema: feed.access.attempt.json
- [x] Create event schema: feed.access.granted.json
- [x] Create event schema: feed.access.denied.json

### Backend - Settings Service
- [x] Add consumer for feed.access.attempt event
- [x] Validate passcode against passcode_hash
- [x] Generate short-lived access token on success
- [x] Emit feed.access.granted with token
- [x] Emit feed.access.denied on failure

### Frontend
- [x] Add passcode prompt UI for protected feeds
- [x] Submit passcode via Notification service (feed.access.attempt)
- [x] Store access token in cookie/localStorage
- [x] Add UI option to choose cookie or localStorage
- [x] Create PasscodePrompt component with modal dialog
- [ ] Check feed visibility before showing posts

---

## Story 13: Locks, Idempotency & Multi-Replica Safety

### Shared Library
- [x] Add shared/utils/idempotency.go with deduplication helper
- [x] Add shared/utils/locks.go with Postgres advisory lock wrapper
- [x] Add shared/middleware/idempotency.go middleware for event consumers

### Services Updates
- [ ] Update settings-service consumer to use idempotency middleware
- [ ] Update video-service consumers to use idempotency middleware
- [ ] Update program-service consumer to use idempotency middleware
- [ ] Update dm-service consumer to use idempotency middleware
- [ ] Update machine-service consumer to use idempotency middleware
- [ ] Update all consumers to use manual ack (ack after DB commit)
- [ ] Update all consumers to requeue on transient failure

### Documentation
- [ ] Document lock acquisition/release behavior in INTEGRATION_NOTES.md

---

## Story 14: OpenAPI & Event Schema Generation

(Already covered in individual stories above - consolidate here)

### Event Schemas Validation
- [ ] Add event schema validation middleware to all consumers
- [ ] Emit event.validation.failed when schema validation fails

---

## Story 15: Prompt Templates Migration

### Database & Migrations
- [ ] Create migration: ai-agent-service create ai_prompt_templates table
- [ ] Create migration: insert default prompt templates (idempotent upsert)
- [ ] Create /ai/templates/coach_workout_completed.txt
- [ ] Create /ai/templates/coach_dm_response.txt
- [ ] Create /ai/templates/coach_program_adjustment.txt

---

## Story 17: Pod Health Checks & Log Policy

### Scripts
- [ ] Create /scripts/pod-health-check.sh script
- [ ] Add kubectl get pods check
- [ ] Add kubectl logs check for crashloop pods
- [ ] Append results to infra-health.log

### Events
- [ ] Create event schema: infra.pods.issue_detected.json

### Integration
- [ ] Add pod health check to CI/CD workflow (run after commit)
- [ ] Document pod health check usage in INTEGRATION_NOTES.md

---

## Story 10: AI Coach (Deferred - Implement After All Non-AI Features)

### Database & Migrations
- [ ] Create new service: ai-agent-service with Dockerfile and structure

### Events
- [ ] Create event schema: ai_coach.requested.json
- [ ] Create event schema: ai_coach.response.json

### Backend - AI Agent Service (STUB FIRST)
- [ ] Create ai-agent-service cmd/main.go with RabbitMQ consumer
- [ ] Add consumer for dm.message.sent (type: ai_coach_request) - stub response
- [ ] Add consumer for workout.completed - stub response
- [ ] Create external OpenWebUI connector client (stubbed)
- [ ] Load prompt templates from Postgres
- [ ] Generate AI response using LiteLLM connector (stubbed)
- [ ] Emit dm.message.sent as ai_coach system user
- [ ] Emit ai_coach.response event

### Backend - AI Agent Service (REAL IMPLEMENTATION)
- [ ] Implement real OpenWebUI connector client
- [ ] Add credential loading from repo secrets
- [ ] Add prompt template rendering with workout history context
- [ ] Add conversation history context for DMs
- [ ] Add logging for all AI interactions

### OpenAPI
- [ ] Generate OpenAPI spec for ai-agent-service endpoints (if any)

---

## Testing Documentation

### Integration Notes
- [ ] Add RabbitMQ event simulation instructions to INTEGRATION_NOTES.md
- [ ] Add consumer dev mode instructions to INTEGRATION_NOTES.md
- [ ] Add local testing setup to INTEGRATION_NOTES.md

### Test TODOs (not implemented, just documented)
- [ ] Add test TODO: auth.user.logged_in event flow
- [ ] Add test TODO: user.settings.submitted idempotency
- [ ] Add test TODO: feed.post.created event flow
- [ ] Add test TODO: video upload and processing flow
- [ ] Add test TODO: comment.created and interaction.liked
- [ ] Add test TODO: program.plan.created and workout.completed
- [ ] Add test TODO: dm.message.sent and dm.pin.attempts
- [ ] Add test TODO: feed.access.attempt validation
- [ ] Add test TODO: Postgres advisory locks under concurrent load
- [ ] Add test TODO: RabbitMQ requeue on transient failure
- [ ] Add test TODO: AI coach response generation

---

## CI/CD & GitHub Actions

### GitHub Actions
- [x] Update .github/workflows/ci.yml matrix to include dm-service, machine-service, reminder-service
- [x] Docker images will be built and pushed to GHCR for all new services

---

## Deployment & K8s

### Kubernetes Manifests
- [x] Add k8s deployment for dm-service (port 8084)
- [x] Add k8s deployment for machine-service (port 8085)
- [x] Add k8s deployment for reminder-service (port 8086)
- [x] Update k8s/base/kustomization.yaml with new services
- [x] Add image mappings for dm-service, machine-service, reminder-service to GHCR
- [ ] Update k8s deployment for video-service (add consumer workers)
- [ ] Add k8s deployment for ai-agent-service (stub)
- [ ] Update infrastructure/helm.tf with new services

---

## Total Tasks: 170+
