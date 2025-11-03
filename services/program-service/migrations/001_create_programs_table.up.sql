CREATE TYPE lift_type AS ENUM ('squat', 'bench', 'deadlift', 'accessory');
CREATE TYPE program_phase AS ENUM ('hypertrophy', 'strength', 'peaking', 'deload', 'off_season');

CREATE TABLE IF NOT EXISTS programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL,
    coach_id UUID,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    phase program_phase NOT NULL DEFAULT 'strength',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    weeks_total INTEGER NOT NULL,
    days_per_week INTEGER NOT NULL DEFAULT 3,
    program_data JSONB NOT NULL DEFAULT '{}',
    ai_generated BOOLEAN DEFAULT TRUE,
    ai_model VARCHAR(100),
    ai_prompt TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS training_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    athlete_id UUID NOT NULL,
    week_number INTEGER NOT NULL,
    day_number INTEGER NOT NULL,
    session_name VARCHAR(255),
    scheduled_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    rpe_rating DECIMAL(2,1) CHECK (rpe_rating >= 1 AND rpe_rating <= 10),
    duration_minutes INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    exercise_order INTEGER NOT NULL,
    lift_type lift_type NOT NULL,
    exercise_name VARCHAR(255) NOT NULL,
    target_sets INTEGER NOT NULL,
    target_reps VARCHAR(50), -- Can be "8-10", "5", "AMRAP", etc.
    target_weight_kg DECIMAL(5,2),
    target_rpe DECIMAL(2,1),
    target_percentage DECIMAL(5,2), -- Percentage of 1RM
    rest_seconds INTEGER,
    notes TEXT,
    tempo VARCHAR(20), -- e.g., "3-1-1-0"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS completed_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    set_number INTEGER NOT NULL,
    reps_completed INTEGER NOT NULL,
    weight_kg DECIMAL(5,2) NOT NULL,
    rpe_actual DECIMAL(2,1),
    video_id UUID, -- Reference to video in video service
    notes TEXT,
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL,
    program_id UUID REFERENCES programs(id) ON DELETE SET NULL,
    conversation_type VARCHAR(50) NOT NULL DEFAULT 'program_generation',
    messages JSONB NOT NULL DEFAULT '[]',
    coach_context_enabled BOOLEAN DEFAULT FALSE,
    coach_feedback_incorporated JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS program_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    experience_level VARCHAR(50) CHECK (experience_level IN ('beginner', 'intermediate', 'advanced', 'elite')),
    phase program_phase NOT NULL,
    weeks_duration INTEGER NOT NULL,
    days_per_week INTEGER NOT NULL,
    template_data JSONB NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_programs_athlete_id ON programs(athlete_id);
CREATE INDEX idx_programs_coach_id ON programs(coach_id);
CREATE INDEX idx_programs_active ON programs(is_active);
CREATE INDEX idx_programs_dates ON programs(start_date, end_date);
CREATE INDEX idx_training_sessions_program_id ON training_sessions(program_id);
CREATE INDEX idx_training_sessions_athlete_id ON training_sessions(athlete_id);
CREATE INDEX idx_training_sessions_scheduled_date ON training_sessions(scheduled_date);
CREATE INDEX idx_exercises_session_id ON exercises(session_id);
CREATE INDEX idx_exercises_lift_type ON exercises(lift_type);
CREATE INDEX idx_completed_sets_exercise_id ON completed_sets(exercise_id);
CREATE INDEX idx_ai_conversations_athlete_id ON ai_conversations(athlete_id);
CREATE INDEX idx_ai_conversations_program_id ON ai_conversations(program_id);
CREATE INDEX idx_program_templates_category ON program_templates(category);
CREATE INDEX idx_program_templates_experience_level ON program_templates(experience_level);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
CREATE TRIGGER update_programs_updated_at 
    BEFORE UPDATE ON programs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_training_sessions_updated_at 
    BEFORE UPDATE ON training_sessions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ai_conversations_updated_at 
    BEFORE UPDATE ON ai_conversations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert some default program templates
INSERT INTO program_templates (name, description, category, experience_level, phase, weeks_duration, days_per_week, template_data, is_public) VALUES
('Beginner Linear Progression', 'Simple linear progression for new lifters', 'Strength', 'beginner', 'strength', 12, 3, '{"exercises": [{"name": "Squat", "sets": 3, "reps": "5", "progression": "linear"}, {"name": "Bench Press", "sets": 3, "reps": "5", "progression": "linear"}, {"name": "Deadlift", "sets": 1, "reps": "5", "progression": "linear"}]}', true),
('Intermediate 5/3/1', 'Wendler 5/3/1 for intermediate lifters', 'Strength', 'intermediate', 'strength', 16, 4, '{"exercises": [{"name": "Squat", "sets": "3-5", "reps": "5/3/1", "progression": "percentage"}, {"name": "Bench Press", "sets": "3-5", "reps": "5/3/1", "progression": "percentage"}, {"name": "Deadlift", "sets": "3-5", "reps": "5/3/1", "progression": "percentage"}, {"name": "Overhead Press", "sets": "3-5", "reps": "5/3/1", "progression": "percentage"}]}', true),
('Peaking Program', 'Competition preparation program', 'Peaking', 'advanced', 'peaking', 8, 5, '{"exercises": [{"name": "Competition Squat", "sets": "1-3", "reps": "1-3", "progression": "opener_second_third"}, {"name": "Competition Bench", "sets": "1-3", "reps": "1-3", "progression": "opener_second_third"}, {"name": "Competition Deadlift", "sets": "1-3", "reps": "1-3", "progression": "opener_second_third"}]}', true);