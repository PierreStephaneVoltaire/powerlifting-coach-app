# Event Schemas

This directory contains JSON schema definitions for all events used in the powerlifting coach app event-driven architecture.

## Schema Structure

Each event schema must include:
- `schema_version`: Version of the schema
- `client_generated_id`: UUID for idempotency
- `user_id`: User associated with the event
- `timestamp`: ISO 8601 timestamp
- `source_service`: Service that generated the event
- Event-specific fields

## Event Categories

### Auth Events
- `auth.user.logged_in.json`

### Settings Events
- `user.settings.submitted.json`
- `user.settings.persisted.json`
- `user.settings.failed.json`

### Feed Events
- `feed.post.created.json`
- `feed.post.updated.json`
- `feed.post.deleted.json`
- `feed.access.attempt.json`
- `feed.access.granted.json`
- `feed.access.denied.json`

### Media Events
- `media.upload_requested.json`
- `media.uploaded.json`
- `media.processed.json`

### Comment Events
- `comment.created.json`
- `comment.persisted.json`
- `interaction.liked.json`

### Program Events
- `program.plan.created.json`
- `program.plan.updated.json`
- `program.plan.persisted.json`
- `workout.started.json`
- `workout.completed.json`

### DM Events
- `dm.message.sent.json`
- `dm.message.persisted.json`
- `dm.pin.attempts.json`

### Tools Events
- `tools.platecalc.query.json`
- `machine.notes.submitted.json`
- `machine.notes.persisted.json`

### AI Events
- `ai_coach.requested.json`
- `ai_coach.response.json`

### Infrastructure Events
- `infra.pods.issue_detected.json`
- `event.validation.failed.json`
