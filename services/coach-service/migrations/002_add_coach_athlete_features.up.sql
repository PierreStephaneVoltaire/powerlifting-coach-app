-- Add coach-athlete relationship management and enhanced profiles

-- Create coach-athlete relationships table
CREATE TABLE IF NOT EXISTS coach_athlete_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL,
    athlete_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'active', 'terminated'
    request_message TEXT,
    requested_at TIMESTAMP DEFAULT NOW(),
    accepted_at TIMESTAMP,
    terminated_at TIMESTAMP,
    terminated_by UUID, -- who terminated the relationship
    termination_reason TEXT,
    cooldown_until TIMESTAMP, -- prevents immediate re-establishment
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(coach_id, athlete_id) -- one relationship per pair
);

-- Create relationship permission log for audit trail
CREATE TABLE IF NOT EXISTS relationship_permission_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relationship_id UUID NOT NULL REFERENCES coach_athlete_relationships(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL, -- 'access_granted', 'access_revoked', 'data_viewed', 'program_modified'
    resource_type VARCHAR(100), -- 'program', 'chat', 'feed', 'video', 'history'
    resource_id UUID,
    performed_by UUID NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    logged_at TIMESTAMP DEFAULT NOW()
);

-- Enhanced coach profiles
CREATE TABLE IF NOT EXISTS coach_certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL,
    certification_name VARCHAR(255) NOT NULL,
    issuing_organization VARCHAR(255),
    issue_date DATE,
    expiry_date DATE,
    verification_status VARCHAR(50) DEFAULT 'unverified', -- 'unverified', 'pending', 'verified'
    verification_document_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS coach_success_stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL,
    athlete_name VARCHAR(255), -- can be anonymized
    achievement TEXT NOT NULL,
    competition_name VARCHAR(255),
    competition_date DATE,
    total_kg DECIMAL(6,2),
    weight_class VARCHAR(50),
    federation VARCHAR(100),
    placement INTEGER,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Add enhanced fields to existing coach profiles (if table exists)
-- These would be added to the coach_profiles table in user-service
-- Documenting here for reference:
-- ALTER TABLE coach_profiles ADD COLUMN IF NOT EXISTS years_experience INTEGER;
-- ALTER TABLE coach_profiles ADD COLUMN IF NOT EXISTS coaching_philosophy TEXT;
-- ALTER TABLE coach_profiles ADD COLUMN IF NOT EXISTS federation_affiliations JSONB DEFAULT '[]'::jsonb;
-- ALTER TABLE coach_profiles ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE;
-- ALTER TABLE coach_profiles ADD COLUMN IF NOT EXISTS availability_status VARCHAR(50) DEFAULT 'accepting'; -- 'accepting', 'waitlist', 'closed'
-- ALTER TABLE coach_profiles ADD COLUMN IF NOT EXISTS monthly_rate DECIMAL(8,2);

-- Feed privacy settings (extends video-service feed functionality)
CREATE TABLE IF NOT EXISTS feed_privacy_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL UNIQUE,
    default_privacy VARCHAR(50) DEFAULT 'private', -- 'private', 'coach_only', 'public'
    allow_coach_share BOOLEAN DEFAULT TRUE,
    auto_post_workouts BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS feed_post_privacy (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL UNIQUE, -- references feed_posts from video-service
    privacy_level VARCHAR(50) NOT NULL, -- 'private', 'coach_only', 'public'
    visible_to_coaches JSONB DEFAULT '[]'::jsonb, -- array of coach UUIDs
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indices for performance
CREATE INDEX idx_relationships_coach ON coach_athlete_relationships(coach_id, status);
CREATE INDEX idx_relationships_athlete ON coach_athlete_relationships(athlete_id, status);
CREATE INDEX idx_relationships_status ON coach_athlete_relationships(status);
CREATE INDEX idx_relationships_cooldown ON coach_athlete_relationships(cooldown_until) WHERE cooldown_until IS NOT NULL;
CREATE INDEX idx_permission_log_relationship ON relationship_permission_log(relationship_id, logged_at DESC);
CREATE INDEX idx_coach_certifications_coach ON coach_certifications(coach_id);
CREATE INDEX idx_coach_success_stories_coach ON coach_success_stories(coach_id);
CREATE INDEX idx_feed_privacy_athlete ON feed_privacy_settings(athlete_id);
CREATE INDEX idx_feed_post_privacy_post ON feed_post_privacy(post_id);

COMMIT;
