USER_STORIES.md
Implementation ordering / priority

Auth + resilient frontend behavior (no backend → frontend must not crash)

Onboarding / Settings form + event flow into Notification → RabbitMQ → settings-service → Postgres

Main feed: read-only resilience, video upload enqueue, metadata (movement, weight, RPE, comments)

Program planner: plan creation, edit, select workout, start workout, complete workout reporting

DMs: 1:1 private chats (user-user), pin attempts, private media share

Tools: plate calculator, machine notes, equipment prefs

Feed access control: public vs passkey (cookie/localStorage)

AI features: AI coach DM, AI adjustments, prompt templates migration (do AI features after the above)

Operational: OpenAPI + event schemas generated, TODO tracking & resume, pod checks & logs, locks & idempotency

Story 1 — Login & initial auth gate

As a returning user
I want to authenticate via Keycloak so I can use the app securely
Acceptance criteria / engineering notes

Frontend uses Keycloak OIDC flow. On failure or Keycloak unreachability, show a friendly offline/auth-unavailable UI that lets the user retry and continues in limited mode (read-only public feed if available cached).

Keycloak groups: athlete, coach, admin. Also video_public permission group used for feed token generation. Agent should not invent roles beyond these unless repo already contains others.

Frontend stores JWT in secure cookie (httpOnly where possible) or localStorage for passkey only. Use same storage conventions that exist in repo.

Emit event on successful login: auth.user.logged_in { user_id, session_id, timestamp } to Notification service.

Events produced

auth.user.logged_in

Services

Frontend, Keycloak, Notification (event emitter)

Story 2 — Onboarding: settings form (first-run)

As a newly authenticated user
I want to fill a single settings form so the app can personalize plans
Fields (must include):

weight (value + units kg/lb), age, target weight class, weeks until comp (int), goals for squat/bench/dead (kg/lb), most important lift, least important lift, recovery rating per lift (1–5), training days/week (int), session length (minutes), weight plan (gain/lose/maintain), form issues list (free text tags), injuries/imbalances, whether coach should evaluate feasibility (bool), federation, knee sleeve (text), deadlift style (sumo/conv), squat stance (wide/narrow), add-per-month ability (2.5/5/none), volume preference (low/high), recovers-from-heavy-deads (bool), height, past competitions (list of {date-range, attempts, total}), video feed visibility (public|passcode).

Acceptance criteria / engineering notes

Frontend validates minimally; still allow saving offline queue if Notification service unreachable.

The frontend must not call settings-service directly. Instead POST to Notification service API endpoint (e.g., /notify/events) which enqueues user.settings.submitted. Notification service publishes to RabbitMQ. Settings service subscribes and persists.

Ensure idempotency: include client_generated_id in event payload. If same id reprocessed, state should not duplicate.

Emit user.settings.submitted event with full payload: { client_generated_id, user_id, settings, timestamp }.

Settings-service subscribes, validates, persists to Postgres, then emits user.settings.persisted { user_id, settings_id, timestamp }.

If validation fails in settings-service, emits user.settings.failed with error codes.

Events produced

user.settings.submitted

user.settings.persisted (from settings service)

user.settings.failed (error path)

Services

Frontend, Notification (producer), RabbitMQ, settings-service (consumer + Postgres), Notification (could publish success to a user-notification queue)

Story 3 — Frontend resilience: backend unavailable

As a user
I want the app to keep working (non-crashy) when backend services are down
Acceptance criteria / engineering notes

All frontend network calls must have timeouts and handle 5xx, 429, and network errors gracefully. Show local cached data or a clear offline state.

For writes (settings, video uploads, comments): queue locally (IndexedDB or repo’s existing client queue) and send to Notification service when available. If Notification is down, persist the queue across reloads.

For reads: use stale cache, show “data may be out of date” banners.

Exponential backoff on retry attempts; surface a “sync pending” indicator in UI.

No direct REST calls between services; frontend only talks to Notification service or to service-side public endpoints strictly for static content. Notification service is the universal event ingress.

Story 4 — Main feed: view posts, resilience, caching

As a user
I want to view a feed of posts (videos + metadata) even when parts of backend are flaky
Acceptance criteria / engineering notes

Feed data served from feed-service which subscribes to events and writes denormalized feed entries to its Postgres/Read store. Frontend queries feed-service read endpoints (or cached snapshots) — but if feed-service unreachable, show cached feed.

Feed entry schema: { post_id, user_id, visibility(public|passcode), pass_hash?, media_url, movement_label, weight, rpe, created_at, comments_count, top_comments[] }.

Comments are accessible via comments-service events. If comments-service down, show previously cached comment threads and disable adding new comments (but allow queueing).

Follow pagination with cursor-based tokens.

Events produced

feed.post.created (from media processing workflow)

feed.post.updated

feed.post.deleted

Services

Frontend, Notification (ingress for new post events), media-processing-service, feed-service, comments-service

Story 5 — Video upload & metadata submission

As a user
I want to upload a lift video and attach movement, weight, RPE and comments
Acceptance criteria / engineering notes

Frontend should upload media to a pre-signed storage URL from media-service if available; if media-service is unreachable, enqueue the media upload operation and capture metadata locally.

UX: allow backgrounding the upload; show progress. If user closes app, resume later.

On successful upload, frontend emits media.uploaded via Notification with metadata including media_id, uploader_id, movement_label, weight, rpe, comment_text, visibility. Include client_generated_id for idempotency.

Media processing service (transcoding, thumbnails) subscribes to media.uploaded and on finish emits media.processed containing media_url, thumbnail_url. That triggers feed.post.created.

Support tagging (movement type enum), optional free text comments. Comments themselves are separate events.

Events produced

media.upload_requested (queued if presigned not available)

media.uploaded

media.processed

feed.post.created

Services

Frontend, Notification, media-service, media-processing-worker, feed-service

Story 6 — Comments & interaction model

As a user
I want to comment on posts and see threaded replies
Acceptance criteria / engineering notes

Comments are emitted as comment.created to Notification. Comments-service persists and emits comment.persisted.

Comment display should degrade gracefully — show cached comments or "comments unavailable".

For like/upvote actions: event interaction.liked (idempotent).

Comments should include moderation flags; lightweight validation client-side.

Events produced

comment.created, comment.persisted, interaction.liked

Services

Frontend, Notification, comments-service

Story 7 — Program planner (create/update till comp)

As a user
I want to create and update my training plan up to comp day and set reminders
Acceptance criteria / engineering notes

Plan object: { plan_id, user_id, start_date, comp_date, workouts: [{date, exercises[], notes}], reminders: [{cron/offset, channel}] }

Frontend emits program.plan.created / program.plan.updated to Notification. Program-service persists and emits program.plan.persisted.

Reminders are scheduled by reminder-service (subscribe to plan events). Reminder-service publishes reminder events to Notification queue when due.

UI must allow editing up until comp day and show weeks_until_comp derived from comp_date.

When frontend marks a workout as started: emit workout.started with plan_id, workout_id, start_timestamp. On completion: workout.completed with duration and summary array of sets/reps/weights. These are used by AI coach later.

Events produced

program.plan.created, program.plan.updated, program.plan.persisted

workout.started, workout.completed

Services

Frontend, Notification, program-service, reminder-service

Story 8 — Start workout / complete workout UX

As a user
I want to start a workout from a program day, record duration and summary at the end
Acceptance criteria / engineering notes

Starting a workout emits workout.started (idempotent with client_generated_id).

On completion, frontend sends workout.completed with duration, exercises_summary[], rpe_summary, notes. Persisted by program-service or workout-service.

AI coach reads full history of completed + future workouts (subscribe or query read-store) to make recommendations.

Events produced

workout.started, workout.completed

Story 9 — DMs: 1:1 private chats (user-user)

As a user
I want private, 1:1 chats with other users to plan and share media
Acceptance criteria / engineering notes

Chats are private between two user ids. Messages emitted as dm.message.sent with conversation_id, sender_id, recipient_id, message_body, attachments[]. Store messages in dm-service (subscribe) and emit dm.message.persisted.

Pin target attempts: a special message type dm.pin.attempts referencing the pinned attempts in a conversation metadata object.

Media in DM may reuse the media upload flow (enqueue/upload then reference by media_id).

Chats must be end-to-end privacy within app scope (server-side privacy controls). No e2e encryption required unless repo already has it.

Events produced

dm.message.sent, dm.message.persisted, dm.pin.attempts

Services

Frontend, Notification, dm-service, media-service

Story 10 — DMs: AI coach (deferred — implement after non-AI features)

As a user
I want to chat with an AI coach which can suggest program adjustments based on conversation & full workout history
Acceptance criteria / engineering notes

AI-agent service subscribes to dm.message.sent (type: ai_coach_request) or workout.completed and processes using LiteLLM (external OpenWebUI connector). AI outputs are posted back into DM as dm.message.sent from ai_coach system user.

All AI interactions should be logged. Prompt templates are stored in Postgres via migration at service startup. Provide schema and migration script to insert templates.

Support optional human co-coach inclusion: conversation.participants may include coach:<user_id>; AI should weight human coach messages higher where indicated. (Behavioral note for implementer — do not change policy.)

The AI coach is a later pass; initial implementation should stub the service to accept events and forward to the external AI connector.

Events produced

ai_coach.requested, ai_coach.response, dm.message.sent (by ai service)

Services

ai-agent-service, Notification, dm-service, external OpenWebUI connector

Story 11 — Tools: plate calculator & machine notes

As a user
I want tools to calculate plate loads and store machine notes for equipment setups
Acceptance criteria / engineering notes

Plate calc can be pure client-side logic. Expose tools.platecalc.query event for optional telemetry.

Machine notes persisted via machine.notes.submitted event. machine-service subscribes and persists notes keyed by brand/model. Notes may be public or private.

UI option to set plate inventory + default units; supported machines: barbell, hack squat, leg press, hex bar, cable machines.

Events produced

tools.platecalc.query (optional), machine.notes.submitted, machine.notes.persisted

Services

Frontend, Notification, machine-service

Story 12 — Feed access control: public vs passcode

As a user
I want to make my feed public or lock it with a passcode I control
Acceptance criteria / engineering notes

During settings, user can set feed_visibility: public | passcode. If passcode, store passcode_hash in settings-service; frontend does not get raw passcode.

When a viewer accesses a passcode-protected feed, frontend prompts for passkey and verifies locally against a temporary token service or via Notification event feed.access.request which validates the passcode via settings-service (but prefer local check: if the owner caches pass_hash? — do not leak). Best practice: viewer submits passcode to Notification service which publishes feed.access.attempt event for auth-service or settings-service to validate and emit feed.access.granted with short-lived access token. Store token in cookie/localStorage per user choice.

Passkey storage on client: localStorage or cookie (explicit UI option). Use secure handling instructions already in repo.

Public feeds skip this check.

Events produced

feed.access.attempt, feed.access.granted, feed.access.denied

Services

Frontend, Notification, settings-service, auth-service

Story 13 — Locks, idempotency & multi-replica safety

As an engineer
I want services to be safe to run in multiple replicas, avoiding double-processing
Acceptance criteria / engineering notes

Use idempotency keys in all inbound events (client_generated_id).

Acquire distributed locks for critical write workflows: prefer existing repo-standard lock mechanism (e.g., Postgres advisory locks if used in repo). If no lock mechanism exists, use Postgres advisory locks as default. Document lock acquisition/release behavior in INSTRUCTION.md.

Event consumers must ack messages only after successful DB commit. Use RabbitMQ manual ack. Consumers must handle redeliveries gracefully.

Story 14 — OpenAPI & Event Schema generation

As an implementer
I want machine-readable OpenAPI for service ingress endpoints + JSON Schemas for events
Acceptance criteria / engineering notes

For each service that interfaces with Notification (ingress) and read endpoints used by frontend, produce OpenAPI specs (yaml) and place them in /specs/ in the repo.

Produce event JSON Schemas for the 20+ events above, placed in /events/schemas/.

Ensure schemas include client_generated_id, user_id, timestamp, source_service, and schema_version.

Implementer must add event-schema validation in the consumer pipeline (schema validation step).

Story 15 — Prompt templates migration

As an implementer
I want prompt templates to be inserted into Postgres at service startup via migration script
Acceptance criteria / engineering notes

Create a migration that inserts templates into ai_prompt_templates table with columns: id, name, description, template_body, version, created_at. Migration runs on deployment.

Provide a fallback: if migration has already run, skip. Use upsert semantics.

Template store must be readable by AI-agent service.

Story 16 — TODO file & resume semantics

As an implementer agent
I want to generate a TODOS_IMPLEMENTER.md before changes and resume from it if interrupted
Acceptance criteria / engineering notes

Before making code changes, create TODOS_IMPLEMENTER.md at repo root listing discrete tasks (one per to-do). Commit it.

Tasks must have states: todo, in-progress, done, blocked. Agent updates file as it progresses and commits changes.

On startup, agent reads file and resumes incomplete tasks. If file missing, agent creates it.

Story 17 — Pod health checks & log policy

As an implementer agent
I want to periodically check pod health and log errors to a central place
Acceptance criteria / engineering notes

Agent should run a periodic check after a feature commit: kubectl get pods -n <service-namespace> and kubectl logs for pods in crashloop for a short window (use existing cluster access credentials present in repo/agent environment).

Record any non-ready pods to infra-health.log in repo (append) and emit event infra.pods.issue_detected to Notification.

All services must log meaningful info at info and error levels. No tests required.