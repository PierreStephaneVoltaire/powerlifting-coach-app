-- Rollback enhanced workout logging features

-- Remove default exercises
DELETE FROM exercise_library WHERE is_custom = FALSE;

-- Drop indices
DROP INDEX IF EXISTS idx_completed_sets_set_type;
DROP INDEX IF EXISTS idx_training_sessions_deleted_at;
DROP INDEX IF EXISTS idx_training_sessions_adhoc;
DROP INDEX IF EXISTS idx_exercise_library_lift_type;
DROP INDEX IF EXISTS idx_exercise_library_created_by;
DROP INDEX IF EXISTS idx_workout_templates_athlete;
DROP INDEX IF EXISTS idx_program_changes_program;
DROP INDEX IF EXISTS idx_program_changes_status;
DROP INDEX IF EXISTS idx_exercises_name_athlete;
DROP INDEX IF EXISTS idx_training_sessions_athlete_completed;

-- Drop new tables
DROP TABLE IF EXISTS program_changes;
DROP TABLE IF EXISTS workout_templates;

-- Remove foreign key and column from exercises
ALTER TABLE exercises DROP COLUMN IF EXISTS exercise_library_id;

-- Drop exercise library
DROP TABLE IF EXISTS exercise_library;

-- Remove added columns from exercises
ALTER TABLE exercises DROP COLUMN IF EXISTS athlete_notes;

-- Remove added columns from training_sessions
ALTER TABLE training_sessions
    DROP COLUMN IF EXISTS is_adhoc,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_reason;

-- Remove competition date from programs
ALTER TABLE programs DROP COLUMN IF EXISTS competition_date;

-- Remove columns from completed_sets
ALTER TABLE completed_sets
    DROP COLUMN IF EXISTS set_type,
    DROP COLUMN IF EXISTS media_urls,
    DROP COLUMN IF EXISTS exercise_notes;

-- Drop set type enum
DROP TYPE IF EXISTS set_type;

COMMIT;
