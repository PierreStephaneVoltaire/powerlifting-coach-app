# TODOS_IMPLEMENTER.md

## Status Legend
- `[ ]` todo
- `[>]` in-progress
- `[x]` done
- `[!]` blocked

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
- [ ] Add "data may be out of date" banner for stale cache
- [x] Add queue persistence across reloads

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
- [ ] Create migration: video-service add media metadata fields (movement_label, weight, rpe, visibility)

### Events
- [x] Create event schema: media.upload_requested.json
- [x] Create event schema: media.uploaded.json
- [x] Create event schema: media.processed.json

### Backend - Video Service (Media & Processing)
- [ ] Add RabbitMQ consumer for media.upload_requested event
- [ ] Add idempotency using client_generated_id
- [ ] Emit media.uploaded event after presigned upload complete
- [ ] Add processing worker consumer for media.uploaded event
- [ ] Add transcoding logic (stub initially)
- [ ] Add thumbnail generation (stub initially)
- [ ] Emit media.processed event with media_url, thumbnail_url
- [ ] Emit feed.post.created event after processing

### Frontend
- [ ] Add video upload UI with progress indicator
- [ ] Add metadata form: movement_label, weight, rpe, comment_text, visibility
- [ ] Add background upload support
- [ ] Add upload resume on app restart
- [ ] Enqueue upload locally if notification-service unreachable

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

### Frontend
- [x] Add comment input UI on posts
- [x] Add threaded comment display
- [x] Add like/upvote button
- [ ] Add cached comments fallback
- [x] Emit comment.created via Notification service
- [x] Emit interaction.liked via Notification service

### OpenAPI
- [ ] Update video-service OpenAPI spec with comment/like endpoints

---

## Story 7: Program Planner

### Database & Migrations
- [ ] Create migration: program-service create programs table
- [ ] Create migration: program-service create workouts table

### Events
- [x] Create event schema: program.plan.created.json
- [x] Create event schema: program.plan.updated.json
- [x] Create event schema: program.plan.persisted.json
- [x] Create event schema: workout.started.json
- [x] Create event schema: workout.completed.json

### Backend - Program Service
- [ ] Add RabbitMQ consumer for program.plan.created event
- [ ] Add RabbitMQ consumer for program.plan.updated event
- [ ] Add idempotency using client_generated_id (Postgres advisory lock)
- [ ] Persist programs to Postgres
- [ ] Emit program.plan.persisted event
- [ ] Add consumer for workout.started event
- [ ] Add consumer for workout.completed event
- [ ] Persist workout sessions to Postgres

### Backend - Reminder Service (NEW)
- [ ] Create new service: reminder-service with Dockerfile and structure
- [ ] Create migration: reminder-service create reminders table
- [ ] Add consumer for program.plan.created to schedule reminders
- [ ] Add consumer for program.plan.updated to update reminders
- [ ] Add scheduler to publish reminder events when due

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
- [ ] Create new service: dm-service with Dockerfile and structure
- [ ] Create migration: dm-service create conversations table
- [ ] Create migration: dm-service create messages table

### Events
- [x] Create event schema: dm.message.sent.json
- [x] Create event schema: dm.message.persisted.json
- [x] Create event schema: dm.pin.attempts.json

### Backend - DM Service
- [ ] Create dm-service cmd/main.go with RabbitMQ consumer
- [ ] Add consumer for dm.message.sent event
- [ ] Add idempotency using client_generated_id
- [ ] Persist messages to Postgres
- [ ] Emit dm.message.persisted event
- [ ] Add consumer for dm.pin.attempts event
- [ ] Persist pinned attempts to conversation metadata

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
- [ ] Create new service: machine-service with Dockerfile and structure
- [ ] Create migration: machine-service create machine_notes table

### Events
- [x] Create event schema: tools.platecalc.query.json (optional telemetry)
- [x] Create event schema: machine.notes.submitted.json
- [x] Create event schema: machine.notes.persisted.json

### Backend - Machine Service
- [ ] Create machine-service cmd/main.go with RabbitMQ consumer
- [ ] Add consumer for machine.notes.submitted event
- [ ] Add idempotency using client_generated_id
- [ ] Persist notes to Postgres
- [ ] Emit machine.notes.persisted event

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
- [ ] Add passcode prompt UI for protected feeds
- [ ] Submit passcode via Notification service (feed.access.attempt)
- [ ] Store access token in cookie/localStorage
- [ ] Add UI option to choose cookie or localStorage
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

## Deployment & K8s

### Kubernetes Manifests
- [ ] Update k8s deployment for video-service (add consumer workers)
- [ ] Add k8s deployment for dm-service
- [ ] Add k8s deployment for reminder-service
- [ ] Add k8s deployment for machine-service
- [ ] Add k8s deployment for ai-agent-service (stub)
- [ ] Update k8s/base/kustomization.yaml with new services
- [ ] Update infrastructure/helm.tf with new services

---

## Total Tasks: 170+
