DROP TRIGGER IF EXISTS update_ai_conversations_updated_at ON ai_conversations;
DROP TRIGGER IF EXISTS update_training_sessions_updated_at ON training_sessions;
DROP TRIGGER IF EXISTS update_programs_updated_at ON programs;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_program_templates_experience_level;
DROP INDEX IF EXISTS idx_program_templates_category;
DROP INDEX IF EXISTS idx_ai_conversations_program_id;
DROP INDEX IF EXISTS idx_ai_conversations_athlete_id;
DROP INDEX IF EXISTS idx_completed_sets_exercise_id;
DROP INDEX IF EXISTS idx_exercises_lift_type;
DROP INDEX IF EXISTS idx_exercises_session_id;
DROP INDEX IF EXISTS idx_training_sessions_scheduled_date;
DROP INDEX IF EXISTS idx_training_sessions_athlete_id;
DROP INDEX IF EXISTS idx_training_sessions_program_id;
DROP INDEX IF EXISTS idx_programs_dates;
DROP INDEX IF EXISTS idx_programs_active;
DROP INDEX IF EXISTS idx_programs_coach_id;
DROP INDEX IF EXISTS idx_programs_athlete_id;

DROP TABLE IF EXISTS program_templates;
DROP TABLE IF EXISTS ai_conversations;
DROP TABLE IF EXISTS completed_sets;
DROP TABLE IF EXISTS exercises;
DROP TABLE IF EXISTS training_sessions;
DROP TABLE IF EXISTS programs;

DROP TYPE IF EXISTS program_phase;
DROP TYPE IF EXISTS lift_type;