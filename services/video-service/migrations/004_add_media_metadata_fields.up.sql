CREATE TYPE movement_label_enum AS ENUM ('squat', 'bench', 'deadlift', 'accessory', 'other');
CREATE TYPE visibility_enum AS ENUM ('public', 'private');

ALTER TABLE videos
ADD COLUMN IF NOT EXISTS movement_label movement_label_enum,
ADD COLUMN IF NOT EXISTS weight DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS rpe DECIMAL(3,1) CHECK (rpe >= 1 AND rpe <= 10),
ADD COLUMN IF NOT EXISTS comment_text TEXT,
ADD COLUMN IF NOT EXISTS visibility visibility_enum NOT NULL DEFAULT 'public';

CREATE INDEX IF NOT EXISTS idx_videos_movement_label ON videos(movement_label);
CREATE INDEX IF NOT EXISTS idx_videos_visibility ON videos(visibility);

CREATE TABLE IF NOT EXISTS media_uploads (
    upload_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    video_id UUID REFERENCES videos(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    upload_status VARCHAR(50) NOT NULL DEFAULT 'requested' CHECK (upload_status IN ('requested', 'uploaded', 'processing', 'completed', 'failed')),
    presigned_url TEXT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    uploaded_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS media_idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    upload_id UUID,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_media_uploads_user_id ON media_uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_media_uploads_status ON media_uploads(upload_status);
CREATE INDEX IF NOT EXISTS idx_media_idempotency_keys_event_type ON media_idempotency_keys(event_type);
