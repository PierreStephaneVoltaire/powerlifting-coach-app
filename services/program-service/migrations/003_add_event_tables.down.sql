DROP TRIGGER IF EXISTS update_workout_sessions_updated_at ON workout_sessions;

DROP INDEX IF EXISTS idx_idempotency_keys_processed_at;
DROP INDEX IF EXISTS idx_idempotency_keys_event_type;
DROP INDEX IF EXISTS idx_workout_sessions_status;
DROP INDEX IF EXISTS idx_workout_sessions_program_id;
DROP INDEX IF EXISTS idx_workout_sessions_user_id;

DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS workout_sessions;
