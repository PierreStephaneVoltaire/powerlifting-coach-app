DROP TRIGGER IF EXISTS create_feedback_notification_trigger ON coach_feedback;
DROP TRIGGER IF EXISTS update_coach_athlete_notes_updated_at ON coach_athlete_notes;
DROP TRIGGER IF EXISTS update_coach_feedback_updated_at ON coach_feedback;

DROP FUNCTION IF EXISTS create_feedback_notification();
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_athlete_progress_date;
DROP INDEX IF EXISTS idx_athlete_progress_athlete_id;
DROP INDEX IF EXISTS idx_athlete_progress_coach_id;
DROP INDEX IF EXISTS idx_coach_notifications_is_read;
DROP INDEX IF EXISTS idx_coach_notifications_type;
DROP INDEX IF EXISTS idx_coach_notifications_athlete_id;
DROP INDEX IF EXISTS idx_coach_notifications_coach_id;
DROP INDEX IF EXISTS idx_feedback_responses_athlete_id;
DROP INDEX IF EXISTS idx_feedback_responses_feedback_id;
DROP INDEX IF EXISTS idx_coach_athlete_notes_type;
DROP INDEX IF EXISTS idx_coach_athlete_notes_athlete_id;
DROP INDEX IF EXISTS idx_coach_athlete_notes_coach_id;
DROP INDEX IF EXISTS idx_coach_feedback_created_at;
DROP INDEX IF EXISTS idx_coach_feedback_incorporated;
DROP INDEX IF EXISTS idx_coach_feedback_priority;
DROP INDEX IF EXISTS idx_coach_feedback_type;
DROP INDEX IF EXISTS idx_coach_feedback_athlete_id;
DROP INDEX IF EXISTS idx_coach_feedback_coach_id;

DROP TABLE IF EXISTS athlete_progress_tracking;
DROP TABLE IF EXISTS coach_notifications;
DROP TABLE IF EXISTS feedback_responses;
DROP TABLE IF EXISTS coach_athlete_notes;
DROP TABLE IF EXISTS coach_feedback;

DROP TYPE IF EXISTS feedback_priority;
DROP TYPE IF EXISTS feedback_type;