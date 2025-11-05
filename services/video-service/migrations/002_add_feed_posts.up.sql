CREATE TABLE IF NOT EXISTS feed_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    video_id UUID REFERENCES videos(id) ON DELETE CASCADE,
    visibility VARCHAR(20) DEFAULT 'public' CHECK (visibility IN ('public', 'passcode')),
    movement_label VARCHAR(100),
    weight_value NUMERIC,
    weight_unit VARCHAR(10) CHECK (weight_unit IN ('kg', 'lb')),
    rpe NUMERIC CHECK (rpe >= 1 AND rpe <= 10),
    comment_text TEXT,
    comments_count INTEGER DEFAULT 0,
    likes_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_feed_posts_user_id ON feed_posts(user_id);
CREATE INDEX idx_feed_posts_visibility ON feed_posts(visibility);
CREATE INDEX idx_feed_posts_created_at ON feed_posts(created_at DESC);
CREATE INDEX idx_feed_posts_post_id ON feed_posts(post_id);

CREATE TRIGGER update_feed_posts_updated_at
    BEFORE UPDATE ON feed_posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE feed_posts IS 'Denormalized feed entries for video posts';
COMMENT ON COLUMN feed_posts.post_id IS 'Unique post identifier from event';
COMMENT ON COLUMN feed_posts.visibility IS 'Post visibility: public or passcode-protected';
COMMENT ON COLUMN feed_posts.movement_label IS 'Type of lift (squat, bench, deadlift, etc)';
COMMENT ON COLUMN feed_posts.comments_count IS 'Cached comment count';
COMMENT ON COLUMN feed_posts.likes_count IS 'Cached like count';
