DROP INDEX IF EXISTS idx_media_idempotency_keys_event_type;
DROP INDEX IF EXISTS idx_media_uploads_status;
DROP INDEX IF EXISTS idx_media_uploads_user_id;
DROP INDEX IF EXISTS idx_videos_visibility;
DROP INDEX IF EXISTS idx_videos_movement_label;

DROP TABLE IF EXISTS media_idempotency_keys;
DROP TABLE IF EXISTS media_uploads;

ALTER TABLE videos
DROP COLUMN IF EXISTS visibility,
DROP COLUMN IF EXISTS comment_text,
DROP COLUMN IF EXISTS rpe,
DROP COLUMN IF EXISTS weight,
DROP COLUMN IF EXISTS movement_label;

DROP TYPE IF EXISTS visibility_enum;
DROP TYPE IF EXISTS movement_label_enum;
