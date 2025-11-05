DROP TRIGGER IF EXISTS update_reminders_updated_at ON reminders;

DROP INDEX IF EXISTS idx_idempotency_keys_event_type;
DROP INDEX IF EXISTS idx_reminders_scheduled_for;
DROP INDEX IF EXISTS idx_reminders_status;
DROP INDEX IF EXISTS idx_reminders_program_id;
DROP INDEX IF EXISTS idx_reminders_user_id;

DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS reminders;

DROP TYPE IF EXISTS reminder_status_enum;
