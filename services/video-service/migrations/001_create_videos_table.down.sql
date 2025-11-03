DROP TRIGGER IF EXISTS set_video_share_token ON videos;
DROP TRIGGER IF EXISTS update_videos_updated_at ON videos;

DROP FUNCTION IF EXISTS set_share_token();
DROP FUNCTION IF EXISTS generate_share_token();
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_video_shares_shared_by;
DROP INDEX IF EXISTS idx_video_shares_video_id;
DROP INDEX IF EXISTS idx_form_feedback_video_id;
DROP INDEX IF EXISTS idx_videos_created_at;
DROP INDEX IF EXISTS idx_videos_public_share_token;
DROP INDEX IF EXISTS idx_videos_status;
DROP INDEX IF EXISTS idx_videos_athlete_id;

DROP TABLE IF EXISTS video_shares;
DROP TABLE IF EXISTS form_feedback;
DROP TABLE IF EXISTS videos;

DROP TYPE IF EXISTS video_status;