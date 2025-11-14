-- Rollback coach-athlete relationship features

-- Drop indices
DROP INDEX IF EXISTS idx_relationships_coach;
DROP INDEX IF EXISTS idx_relationships_athlete;
DROP INDEX IF EXISTS idx_relationships_status;
DROP INDEX IF EXISTS idx_relationships_cooldown;
DROP INDEX IF EXISTS idx_permission_log_relationship;
DROP INDEX IF EXISTS idx_coach_certifications_coach;
DROP INDEX IF EXISTS idx_coach_success_stories_coach;
DROP INDEX IF EXISTS idx_feed_privacy_athlete;
DROP INDEX IF EXISTS idx_feed_post_privacy_post;

-- Drop tables
DROP TABLE IF EXISTS feed_post_privacy;
DROP TABLE IF EXISTS feed_privacy_settings;
DROP TABLE IF EXISTS coach_success_stories;
DROP TABLE IF EXISTS coach_certifications;
DROP TABLE IF EXISTS relationship_permission_log;
DROP TABLE IF EXISTS coach_athlete_relationships;

COMMIT;
