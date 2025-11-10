-- Drop indexes
DROP INDEX IF EXISTS idx_user_settings_competition_date;
DROP INDEX IF EXISTS idx_user_settings_has_competed;

-- Remove competition and best lift fields from user_settings
ALTER TABLE user_settings
DROP COLUMN IF EXISTS has_competed,
DROP COLUMN IF EXISTS best_squat_kg,
DROP COLUMN IF EXISTS best_bench_kg,
DROP COLUMN IF EXISTS best_dead_kg,
DROP COLUMN IF EXISTS best_total_kg,
DROP COLUMN IF EXISTS comp_pr_date,
DROP COLUMN IF EXISTS comp_federation,
DROP COLUMN IF EXISTS squat_bar_position,
DROP COLUMN IF EXISTS competition_date;
