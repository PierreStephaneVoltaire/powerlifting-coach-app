CREATE TABLE IF NOT EXISTS user_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    theme VARCHAR(20) DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto')),
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    units VARCHAR(10) DEFAULT 'metric' CHECK (units IN ('metric', 'imperial')),
    notifications JSONB DEFAULT '{"email": true, "push": true, "sms": false}',
    privacy JSONB DEFAULT '{"profile_public": false, "videos_public": false}',
    training_preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS app_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) NOT NULL UNIQUE,
    value JSONB NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);
CREATE INDEX idx_app_settings_key ON app_settings(key);
CREATE INDEX idx_app_settings_category ON app_settings(category);
CREATE INDEX idx_app_settings_is_public ON app_settings(is_public);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
CREATE TRIGGER update_user_settings_updated_at 
    BEFORE UPDATE ON user_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_app_settings_updated_at 
    BEFORE UPDATE ON app_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default app settings
INSERT INTO app_settings (key, value, description, category, is_public) VALUES
('max_video_size_mb', '100', 'Maximum video file size in MB', 'video', true),
('supported_video_formats', '["mp4", "mov", "avi", "mkv", "webm"]', 'Supported video file formats', 'video', true),
('ai_model_endpoint', '"http://litellm:4000"', 'AI model endpoint for form analysis', 'ai', false),
('maintenance_mode', 'false', 'Global maintenance mode flag', 'system', true),
('app_version', '"1.0.0"', 'Current application version', 'system', true),
('terms_version', '"1.0.0"', 'Current terms of service version', 'legal', true),
('privacy_version', '"1.0.0"', 'Current privacy policy version', 'legal', true);