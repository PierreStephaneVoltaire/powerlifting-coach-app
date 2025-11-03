CREATE TYPE feedback_type AS ENUM ('program_adjustment', 'form_correction', 'general_note', 'motivation');
CREATE TYPE feedback_priority AS ENUM ('low', 'medium', 'high', 'urgent');

CREATE TABLE IF NOT EXISTS coach_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL,
    athlete_id UUID NOT NULL,
    feedback_type feedback_type NOT NULL DEFAULT 'general_note',
    priority feedback_priority NOT NULL DEFAULT 'medium',
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    reference_type VARCHAR(50), -- 'program', 'session', 'video', 'exercise'
    reference_id UUID,
    tags TEXT[],
    is_private BOOLEAN DEFAULT FALSE,
    incorporated_by_ai BOOLEAN DEFAULT FALSE,
    incorporated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS coach_athlete_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL,
    athlete_id UUID NOT NULL,
    note_type VARCHAR(50) NOT NULL DEFAULT 'general', -- 'general', 'injury', 'goal', 'preference'
    title VARCHAR(255),
    content TEXT NOT NULL,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS feedback_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feedback_id UUID NOT NULL REFERENCES coach_feedback(id) ON DELETE CASCADE,
    athlete_id UUID NOT NULL,
    response_text TEXT NOT NULL,
    is_acknowledgment BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS coach_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL,
    athlete_id UUID NOT NULL,
    notification_type VARCHAR(50) NOT NULL, -- 'new_video', 'missed_session', 'program_completed', 'question'
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    reference_type VARCHAR(50),
    reference_id UUID,
    is_read BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS athlete_progress_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL,
    athlete_id UUID NOT NULL,
    tracking_date DATE NOT NULL,
    body_weight_kg DECIMAL(5,2),
    squat_max_kg DECIMAL(5,2),
    bench_max_kg DECIMAL(5,2),
    deadlift_max_kg DECIMAL(5,2),
    total_kg DECIMAL(6,2),
    notes TEXT,
    measurements JSONB DEFAULT '{}', -- Body measurements, RPE averages, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_coach_feedback_coach_id ON coach_feedback(coach_id);
CREATE INDEX idx_coach_feedback_athlete_id ON coach_feedback(athlete_id);
CREATE INDEX idx_coach_feedback_type ON coach_feedback(feedback_type);
CREATE INDEX idx_coach_feedback_priority ON coach_feedback(priority);
CREATE INDEX idx_coach_feedback_incorporated ON coach_feedback(incorporated_by_ai);
CREATE INDEX idx_coach_feedback_created_at ON coach_feedback(created_at);

CREATE INDEX idx_coach_athlete_notes_coach_id ON coach_athlete_notes(coach_id);
CREATE INDEX idx_coach_athlete_notes_athlete_id ON coach_athlete_notes(athlete_id);
CREATE INDEX idx_coach_athlete_notes_type ON coach_athlete_notes(note_type);

CREATE INDEX idx_feedback_responses_feedback_id ON feedback_responses(feedback_id);
CREATE INDEX idx_feedback_responses_athlete_id ON feedback_responses(athlete_id);

CREATE INDEX idx_coach_notifications_coach_id ON coach_notifications(coach_id);
CREATE INDEX idx_coach_notifications_athlete_id ON coach_notifications(athlete_id);
CREATE INDEX idx_coach_notifications_type ON coach_notifications(notification_type);
CREATE INDEX idx_coach_notifications_is_read ON coach_notifications(is_read);

CREATE INDEX idx_athlete_progress_coach_id ON athlete_progress_tracking(coach_id);
CREATE INDEX idx_athlete_progress_athlete_id ON athlete_progress_tracking(athlete_id);
CREATE INDEX idx_athlete_progress_date ON athlete_progress_tracking(tracking_date);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
CREATE TRIGGER update_coach_feedback_updated_at 
    BEFORE UPDATE ON coach_feedback 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_coach_athlete_notes_updated_at 
    BEFORE UPDATE ON coach_athlete_notes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to create notification when feedback is created
CREATE OR REPLACE FUNCTION create_feedback_notification()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO coach_notifications (
        coach_id, 
        athlete_id, 
        notification_type, 
        title, 
        message, 
        reference_type, 
        reference_id
    ) VALUES (
        NEW.coach_id,
        NEW.athlete_id,
        'feedback_given',
        'New Feedback: ' || NEW.title,
        SUBSTRING(NEW.content, 1, 200),
        'feedback',
        NEW.id
    );
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to create notification when feedback is given
CREATE TRIGGER create_feedback_notification_trigger 
    AFTER INSERT ON coach_feedback 
    FOR EACH ROW EXECUTE FUNCTION create_feedback_notification();