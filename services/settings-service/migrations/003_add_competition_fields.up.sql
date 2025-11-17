-- Add competition and best lift fields to user_settings
ALTER TABLE user_settings
ADD COLUMN IF NOT EXISTS has_competed BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS best_squat_kg NUMERIC CHECK (best_squat_kg >= 0),
ADD COLUMN IF NOT EXISTS best_bench_kg NUMERIC CHECK (best_bench_kg >= 0),
ADD COLUMN IF NOT EXISTS best_dead_kg NUMERIC CHECK (best_dead_kg >= 0),
ADD COLUMN IF NOT EXISTS best_total_kg NUMERIC CHECK (best_total_kg >= 0),
ADD COLUMN IF NOT EXISTS comp_pr_date DATE,
ADD COLUMN IF NOT EXISTS comp_federation VARCHAR(100),
ADD COLUMN IF NOT EXISTS squat_bar_position VARCHAR(20) CHECK (squat_bar_position IN ('high', 'medium', 'low', 'french')),
ADD COLUMN IF NOT EXISTS competition_date DATE;

-- Add indexes for frequently queried fields
CREATE INDEX IF NOT EXISTS idx_user_settings_competition_date ON user_settings(competition_date) WHERE competition_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_settings_has_competed ON user_settings(has_competed);

-- Add comments for new fields
COMMENT ON COLUMN user_settings.has_competed IS 'Whether user has competed in powerlifting before';
COMMENT ON COLUMN user_settings.best_squat_kg IS 'Best squat in kilograms (competition or gym PR)';
COMMENT ON COLUMN user_settings.best_bench_kg IS 'Best bench press in kilograms (competition or gym PR)';
COMMENT ON COLUMN user_settings.best_dead_kg IS 'Best deadlift in kilograms (competition or gym PR)';
COMMENT ON COLUMN user_settings.best_total_kg IS 'Best competition total in kilograms';
COMMENT ON COLUMN user_settings.comp_pr_date IS 'Date when competition PRs were set';
COMMENT ON COLUMN user_settings.comp_federation IS 'Federation where competition PRs were set (e.g., CPU, IPF, USAPL)';
COMMENT ON COLUMN user_settings.squat_bar_position IS 'Squat bar position preference (high bar, medium bar, low bar, or french)';
COMMENT ON COLUMN user_settings.competition_date IS 'Date of next/target competition';
