-- Add fields to support pending program approval workflow
ALTER TABLE programs
ADD COLUMN pending_program_data JSONB DEFAULT NULL,
ADD COLUMN program_status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (program_status IN ('draft', 'pending_approval', 'approved', 'rejected'));

-- Add index for efficient querying
CREATE INDEX idx_programs_status ON programs(program_status);
CREATE INDEX idx_programs_athlete_status ON programs(athlete_id, program_status);

-- Add a comment to explain the fields
COMMENT ON COLUMN programs.pending_program_data IS 'Stores AI-generated program data awaiting user approval. Once approved, this data is moved to program_data.';
COMMENT ON COLUMN programs.program_status IS 'Tracks program approval status: draft (being created), pending_approval (awaiting user confirmation), approved (active), rejected (user declined)';
