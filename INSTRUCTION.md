Agent role: Claude — Implementer (strict execution rules)

Primary goal: implement the user stories in instruction.md , in the order specified (non-AI features first). Write minimal comments, include logs, do not write tests. Generate OpenAPI and event schema files. Insert AI prompt templates via migration. Use existing repo conventions; do not invent unrelated folders unless strictly necessary and justified in the commit message.

Work unit atomicity: Each change must be a small commit that updates TODOS_IMPLEMENTER.md (mark in-progress), then commit again when done. No multi-feature mega-commit.

Before coding: create TODOS_IMPLEMENTER.md listing tasks derived from USER_STORIES.md. Commit the TODO file.

Resume behavior: On restart, read TODOS_IMPLEMENTER.md and resume any in-progress or todo tasks. If a task is blocked, leave it and continue other todos.

Event-only comms: No direct REST calls between services for data writes. All writes must go through Notification service which publishes to RabbitMQ. Services persist based on subscriptions. (Exception: read-only queries that are served by read services are ok. communication to infra like db cache,authentication-service, keycloak and litellm are fine)

Idempotency: All outgoing events must include client_generated_id. Consumers must dedupe on that id.

Locks: Use Postgres advisory locks for multi-replica critical sections unless repo contains a different approved lock mechanism. Acquire lock using pg_try_advisory_lock(hash) before proceeding; release after transaction commit or on failure. Log lock acquisition/release.

RabbitMQ semantics: Use manual ack and requeue on transient failure. Only ack after DB transaction commit.

OpenAPI & schemas: Generate specs under /specs/ and events/schemas/. Each schema must declare schema_version and example payload. Add validation middleware in consumer where possible (schema validation fails → emit event.validation.failed).

Prompt templates migration: Add a migration (SQL or migrations-tool file) that INSERT ... ON CONFLICT DO UPDATE the ai_prompt_templates table. Migration must run idempotently at startup. Put templates under /ai/templates/ as source-of-truth too.

Logging: Use the repo’s logging library. If none, use structured JSON logs with level, ts, service, msg, ctx. Log on info for success paths and error for failures. Append infra pod-check outputs to infra-health.log.

Minimal comments: Do not add verbose comments. Favor clear variable names.

No tests required. Add TODOs for tests in TODOS_IMPLEMENTER.md.

External AI connectivity: Do not hardcode credentials. Use repo’s secrets mechanism. AI connector should call external OpenWebUI endpoint; implement wrapper client that can be stubbed until real creds are available.

Pod health check frequency: Run checks once after each PR/commit and when resuming an interrupted run. Log results as described.

Feed passkey UX: Frontend submits passcode via Notification service which emits feed.access.attempt. settings-service validates and emits feed.access.granted/denied. If granted, frontend stores short-lived token in cookie/localStorage as per repo convention. access control for feeds can use keycloack 

Database migrations: All schema changes (prompts or minor tables) must be added as migration files and committed. Keep migration atomic and idempotent.

Safety & privacy: Do not expose raw passcodes in logs. Mask any sensitive values in logs.

When blocked: If a missing repo convention or dependency blocks progress, mark task blocked and leave a single-line explanation in TODO file (no external communication). Continue other tasks.

What to do about existing infra: If the repo already contains a mechanism for locks, queues or storage, use it. Only add defaults (Postgres advisory locks, IndexedDB queue) if no existing solution is discoverable in repo. Document that decision in INTEGRATION_NOTES.md.

AI features: Implement stubs first (no external AI calls). After all non-AI features are green, implement AI flows: AI-agent service that subscribes to workout.completed and dm.message.sent (ai_coach_request). Insert prompt templates via migration before enabling AI worker.

Event schema citations: For any event referenced in code, include a short example JSON in /events/schemas/<event>.json with example field.

Testing locally: Add instructions to INTEGRATION_NOTES.md for simulating RabbitMQ events (e.g., rabbitmqadmin publish example), running consumer in dev mode, and running the Pod health check.

Time awareness: Don’t invent dates or versions. Use whatever the repo already has. If the repo has no migration tool or conventions, create a SQL migration.