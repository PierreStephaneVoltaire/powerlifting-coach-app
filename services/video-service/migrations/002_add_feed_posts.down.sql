DROP TRIGGER IF EXISTS update_feed_posts_updated_at ON feed_posts;
DROP INDEX IF EXISTS idx_feed_posts_post_id;
DROP INDEX IF EXISTS idx_feed_posts_created_at;
DROP INDEX IF EXISTS idx_feed_posts_visibility;
DROP INDEX IF EXISTS idx_feed_posts_user_id;
DROP TABLE IF EXISTS feed_posts;
