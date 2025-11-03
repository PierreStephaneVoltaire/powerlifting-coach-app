CREATE TYPE machine_type_enum AS ENUM ('barbell', 'hack_squat', 'leg_press', 'hex_bar', 'cable', 'other');
CREATE TYPE visibility_enum AS ENUM ('public', 'private');

CREATE TABLE IF NOT EXISTS machine_notes (
    note_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    brand VARCHAR(255) NOT NULL,
    model VARCHAR(255),
    machine_type machine_type_enum NOT NULL,
    settings TEXT NOT NULL,
    visibility visibility_enum NOT NULL DEFAULT 'private',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_machine_notes_user_id ON machine_notes(user_id);
CREATE INDEX IF NOT EXISTS idx_machine_notes_machine_type ON machine_notes(machine_type);
CREATE INDEX IF NOT EXISTS idx_machine_notes_visibility ON machine_notes(visibility);
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_event_type ON idempotency_keys(event_type);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_machine_notes_updated_at
    BEFORE UPDATE ON machine_notes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
