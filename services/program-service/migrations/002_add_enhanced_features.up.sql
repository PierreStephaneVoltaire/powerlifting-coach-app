-- Add enhanced workout logging features

-- Set type enum for set tagging
CREATE TYPE set_type AS ENUM (
    'warm_up',
    'working',
    'backoff',
    'amrap',
    'failure',
    'drop_set',
    'cluster',
    'pause',
    'tempo',
    'custom'
);

-- Add columns to completed_sets for enhanced logging
ALTER TABLE completed_sets
    ADD COLUMN set_type set_type DEFAULT 'working',
    ADD COLUMN media_urls JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN exercise_notes TEXT;

-- Add competition date to programs
ALTER TABLE programs
    ADD COLUMN competition_date TIMESTAMP;

-- Add session-level features
ALTER TABLE training_sessions
    ADD COLUMN is_adhoc BOOLEAN DEFAULT FALSE,
    ADD COLUMN deleted_at TIMESTAMP,
    ADD COLUMN deleted_reason TEXT;

-- Add exercise-level notes support
ALTER TABLE exercises
    ADD COLUMN athlete_notes TEXT;

-- Create exercise library table
CREATE TABLE IF NOT EXISTS exercise_library (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    lift_type lift_type NOT NULL,
    primary_muscles JSONB DEFAULT '[]'::jsonb,
    secondary_muscles JSONB DEFAULT '[]'::jsonb,
    difficulty VARCHAR(50), -- beginner, intermediate, advanced
    equipment_needed JSONB DEFAULT '[]'::jsonb,
    demo_video_url TEXT,
    instructions TEXT,
    form_cues JSONB DEFAULT '[]'::jsonb,
    is_custom BOOLEAN DEFAULT FALSE,
    created_by UUID, -- athlete who created custom exercise
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Link exercises to library (optional)
ALTER TABLE exercises
    ADD COLUMN exercise_library_id UUID REFERENCES exercise_library(id);

-- Create workout templates table
CREATE TABLE IF NOT EXISTS workout_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    template_data JSONB NOT NULL, -- stores exercise structure
    is_public BOOLEAN DEFAULT FALSE,
    times_used INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create program change log for git-like management
CREATE TABLE IF NOT EXISTS program_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL, -- 'propose', 'approve', 'reject', 'apply'
    proposed_changes JSONB NOT NULL, -- diff of changes
    change_description TEXT,
    proposed_by VARCHAR(50) DEFAULT 'ai', -- 'ai', 'coach', 'athlete'
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'applied'
    created_at TIMESTAMP DEFAULT NOW(),
    applied_at TIMESTAMP
);

-- Create indices for performance
CREATE INDEX idx_completed_sets_set_type ON completed_sets(set_type);
CREATE INDEX idx_training_sessions_deleted_at ON training_sessions(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_training_sessions_adhoc ON training_sessions(is_adhoc);
CREATE INDEX idx_exercise_library_lift_type ON exercise_library(lift_type);
CREATE INDEX idx_exercise_library_created_by ON exercise_library(created_by);
CREATE INDEX idx_workout_templates_athlete ON workout_templates(athlete_id);
CREATE INDEX idx_program_changes_program ON program_changes(program_id);
CREATE INDEX idx_program_changes_status ON program_changes(status);

-- Add index for previous set autofill queries
CREATE INDEX idx_exercises_name_athlete ON exercises(exercise_name, session_id);
CREATE INDEX idx_training_sessions_athlete_completed ON training_sessions(athlete_id, completed_at DESC)
    WHERE completed_at IS NOT NULL;

-- Insert default powerlifting exercises into library
INSERT INTO exercise_library (name, description, lift_type, primary_muscles, secondary_muscles, difficulty, equipment_needed, instructions, form_cues) VALUES
    ('Back Squat', 'Competition-style back squat', 'squat', '["quadriceps", "glutes"]', '["hamstrings", "core", "erectors"]', 'intermediate', '["barbell", "squat_rack"]', 'Bar positioned on upper traps, descend until hip crease below knee, drive up maintaining tension', '["Brace core hard", "Break at knees and hips simultaneously", "Keep chest up", "Drive through whole foot", "Squeeze glutes at top"]'),
    ('Bench Press', 'Competition-style bench press', 'bench', '["chest", "triceps"]', '["shoulders", "lats"]', 'intermediate', '["barbell", "bench"]', 'Retract scapula, arch back, lower to chest with elbows at 45 degrees, press explosively', '["Leg drive", "Tight upper back", "Touch chest", "Bar path slightly back", "Lockout completely"]'),
    ('Deadlift', 'Conventional deadlift', 'deadlift', '["hamstrings", "glutes", "erectors"]', '["lats", "traps", "core", "grip"]', 'advanced', '["barbell"]', 'Hip hinge with neutral spine, pull slack from bar, drive through floor, lock out hips and knees', '["Lats engaged", "Neutral spine", "Push floor away", "Hip hinge not squat", "Full lockout"]'),
    ('Sumo Deadlift', 'Wide stance deadlift', 'deadlift', '["quads", "glutes", "adductors"]', '["hamstrings", "erectors", "grip"]', 'advanced', '["barbell"]', 'Wide stance, vertical shins, pull slack, drive knees out and hips forward', '["Vertical torso", "Knees out over toes", "Pull bar close", "Lead with chest"]'),
    ('Front Squat', 'Front-loaded squat variation', 'squat', '["quadriceps", "core"]', '["glutes", "upper back"]', 'advanced', '["barbell", "squat_rack"]', 'Clean grip or cross-arms, maintain upright torso, descend keeping elbows high', '["Elbows high", "Upright torso", "Core braced", "Drive through heels"]'),
    ('Pause Squat', 'Squat with pause at bottom', 'squat', '["quadriceps", "glutes"]', '["hamstrings", "core"]', 'intermediate', '["barbell", "squat_rack"]', 'Descend to depth, pause 2-3 seconds, explode up maintaining tightness', '["Stay tight in pause", "No relaxation", "Explosive out of hole"]'),
    ('Pause Bench', 'Bench press with pause on chest', 'bench', '["chest", "triceps"]', '["shoulders"]', 'intermediate', '["barbell", "bench"]', 'Lower to chest, pause 2-3 seconds, press explosively', '["Maintain tension in pause", "No bouncing", "Controlled descent"]'),
    ('Romanian Deadlift', 'Hip hinge deadlift variation', 'deadlift', '["hamstrings", "glutes"]', '["erectors", "lats"]', 'intermediate', '["barbell"]', 'Slight knee bend, hip hinge lowering bar along shins, feel stretch in hamstrings', '["Soft knees", "Hinge at hips", "Bar close to body", "Stretch hamstrings"]'),
    ('Deficit Deadlift', 'Deadlift from elevated platform', 'deadlift', '["hamstrings", "quads", "erectors"]', '["glutes", "grip"]', 'advanced', '["barbell", "platform"]', 'Stand on 1-3 inch platform, maintain form through increased range', '["Longer pull", "Maintain back position", "Controlled setup"]'),
    ('Close Grip Bench', 'Narrow grip bench press', 'bench', '["triceps", "chest"]', '["shoulders"]', 'intermediate', '["barbell", "bench"]', 'Hands shoulder-width or narrower, elbows closer to body, press with tricep emphasis', '["Elbows in", "Tricep focus", "Full lockout"]'),
    ('Overhead Press', 'Standing barbell press', 'accessory', '["shoulders", "triceps"]', '["core", "upper back"]', 'intermediate', '["barbell"]', 'Bar at clavicle, press overhead keeping core tight, lock out overhead', '["Brace core", "Drive with legs", "Head through at top", "Full lockout"]'),
    ('Barbell Row', 'Bent-over barbell row', 'accessory', '["lats", "upper back"]', '["biceps", "erectors"]', 'intermediate', '["barbell"]', 'Hip hinge position, pull bar to lower chest, squeeze shoulder blades', '["Flat back", "Pull to sternum", "Squeeze at top", "Control eccentric"]'),
    ('Leg Press', 'Machine leg press', 'accessory', '["quadriceps", "glutes"]', '["hamstrings"]', 'beginner', '["leg_press"]', 'Feet shoulder-width, lower with control, press through full foot', '["Full range of motion", "Dont lock knees", "Control descent"]'),
    ('Dumbbell Bench', 'Dumbbell bench press', 'bench', '["chest", "triceps"]', '["shoulders"]', 'intermediate', '["dumbbells", "bench"]', 'Dumbbells at chest level, press up bringing together at top', '["Deeper stretch", "Stabilization", "Full ROM"]'),
    ('Pin Squat', 'Squat from pins at depth', 'squat', '["quadriceps", "glutes"]', '["hamstrings"]', 'advanced', '["barbell", "squat_rack", "pins"]', 'Start from bottom position on pins, remove slack, explode up', '["Dead start", "No stretch reflex", "Maximum power"]');

COMMIT;
