DROP TRIGGER IF EXISTS update_machine_notes_updated_at ON machine_notes;

DROP INDEX IF EXISTS idx_idempotency_keys_event_type;
DROP INDEX IF EXISTS idx_machine_notes_visibility;
DROP INDEX IF EXISTS idx_machine_notes_machine_type;
DROP INDEX IF EXISTS idx_machine_notes_user_id;

DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS machine_notes;

DROP TYPE IF EXISTS visibility_enum;
DROP TYPE IF EXISTS machine_type_enum;
