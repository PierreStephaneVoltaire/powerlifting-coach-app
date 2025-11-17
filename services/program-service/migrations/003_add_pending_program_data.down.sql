-- Remove pending program approval fields
DROP INDEX IF EXISTS idx_programs_athlete_status;
DROP INDEX IF EXISTS idx_programs_status;

ALTER TABLE programs
DROP COLUMN IF EXISTS pending_program_data,
DROP COLUMN IF EXISTS program_status;
