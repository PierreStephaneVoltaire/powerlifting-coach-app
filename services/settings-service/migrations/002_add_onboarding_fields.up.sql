ALTER TABLE user_settings
ADD COLUMN IF NOT EXISTS weight_value NUMERIC,
ADD COLUMN IF NOT EXISTS weight_unit VARCHAR(10) CHECK (weight_unit IN ('kg', 'lb')),
ADD COLUMN IF NOT EXISTS age INTEGER CHECK (age >= 13 AND age <= 120),
ADD COLUMN IF NOT EXISTS target_weight_class VARCHAR(50),
ADD COLUMN IF NOT EXISTS weeks_until_comp INTEGER CHECK (weeks_until_comp >= 0),
ADD COLUMN IF NOT EXISTS squat_goal_value NUMERIC,
ADD COLUMN IF NOT EXISTS squat_goal_unit VARCHAR(10) CHECK (squat_goal_unit IN ('kg', 'lb')),
ADD COLUMN IF NOT EXISTS bench_goal_value NUMERIC,
ADD COLUMN IF NOT EXISTS bench_goal_unit VARCHAR(10) CHECK (bench_goal_unit IN ('kg', 'lb')),
ADD COLUMN IF NOT EXISTS dead_goal_value NUMERIC,
ADD COLUMN IF NOT EXISTS dead_goal_unit VARCHAR(10) CHECK (dead_goal_unit IN ('kg', 'lb')),
ADD COLUMN IF NOT EXISTS most_important_lift VARCHAR(20) CHECK (most_important_lift IN ('squat', 'bench', 'deadlift')),
ADD COLUMN IF NOT EXISTS least_important_lift VARCHAR(20) CHECK (least_important_lift IN ('squat', 'bench', 'deadlift')),
ADD COLUMN IF NOT EXISTS recovery_rating_squat INTEGER CHECK (recovery_rating_squat >= 1 AND recovery_rating_squat <= 5),
ADD COLUMN IF NOT EXISTS recovery_rating_bench INTEGER CHECK (recovery_rating_bench >= 1 AND recovery_rating_bench <= 5),
ADD COLUMN IF NOT EXISTS recovery_rating_dead INTEGER CHECK (recovery_rating_dead >= 1 AND recovery_rating_dead <= 5),
ADD COLUMN IF NOT EXISTS training_days_per_week INTEGER CHECK (training_days_per_week >= 1 AND training_days_per_week <= 7),
ADD COLUMN IF NOT EXISTS session_length_minutes INTEGER CHECK (session_length_minutes >= 15 AND session_length_minutes <= 300),
ADD COLUMN IF NOT EXISTS weight_plan VARCHAR(20) CHECK (weight_plan IN ('gain', 'lose', 'maintain')),
ADD COLUMN IF NOT EXISTS form_issues TEXT[],
ADD COLUMN IF NOT EXISTS injuries TEXT,
ADD COLUMN IF NOT EXISTS evaluate_feasibility BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS federation VARCHAR(50),
ADD COLUMN IF NOT EXISTS knee_sleeve VARCHAR(50),
ADD COLUMN IF NOT EXISTS deadlift_style VARCHAR(20) CHECK (deadlift_style IN ('sumo', 'conventional')),
ADD COLUMN IF NOT EXISTS squat_stance VARCHAR(20) CHECK (squat_stance IN ('wide', 'narrow', 'medium')),
ADD COLUMN IF NOT EXISTS add_per_month VARCHAR(20) CHECK (add_per_month IN ('2.5kg', '5kg', 'none')),
ADD COLUMN IF NOT EXISTS volume_preference VARCHAR(20) CHECK (volume_preference IN ('low', 'high', 'medium')),
ADD COLUMN IF NOT EXISTS recovers_from_heavy_deads BOOLEAN,
ADD COLUMN IF NOT EXISTS height_value NUMERIC,
ADD COLUMN IF NOT EXISTS height_unit VARCHAR(10) CHECK (height_unit IN ('cm', 'in')),
ADD COLUMN IF NOT EXISTS past_competitions JSONB DEFAULT '[]',
ADD COLUMN IF NOT EXISTS feed_visibility VARCHAR(20) DEFAULT 'public' CHECK (feed_visibility IN ('public', 'passcode')),
ADD COLUMN IF NOT EXISTS passcode_hash VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_user_settings_feed_visibility ON user_settings(feed_visibility);
CREATE INDEX IF NOT EXISTS idx_user_settings_weeks_until_comp ON user_settings(weeks_until_comp) WHERE weeks_until_comp IS NOT NULL;

COMMENT ON COLUMN user_settings.weight_value IS 'User body weight value';
COMMENT ON COLUMN user_settings.weight_unit IS 'Unit for body weight (kg or lb)';
COMMENT ON COLUMN user_settings.age IS 'User age in years';
COMMENT ON COLUMN user_settings.target_weight_class IS 'Target competition weight class';
COMMENT ON COLUMN user_settings.weeks_until_comp IS 'Weeks until next competition';
COMMENT ON COLUMN user_settings.squat_goal_value IS 'Squat goal weight value';
COMMENT ON COLUMN user_settings.squat_goal_unit IS 'Unit for squat goal (kg or lb)';
COMMENT ON COLUMN user_settings.bench_goal_value IS 'Bench press goal weight value';
COMMENT ON COLUMN user_settings.bench_goal_unit IS 'Unit for bench goal (kg or lb)';
COMMENT ON COLUMN user_settings.dead_goal_value IS 'Deadlift goal weight value';
COMMENT ON COLUMN user_settings.dead_goal_unit IS 'Unit for deadlift goal (kg or lb)';
COMMENT ON COLUMN user_settings.most_important_lift IS 'User most important lift priority';
COMMENT ON COLUMN user_settings.least_important_lift IS 'User least important lift priority';
COMMENT ON COLUMN user_settings.recovery_rating_squat IS 'Squat recovery rating 1-5';
COMMENT ON COLUMN user_settings.recovery_rating_bench IS 'Bench recovery rating 1-5';
COMMENT ON COLUMN user_settings.recovery_rating_dead IS 'Deadlift recovery rating 1-5';
COMMENT ON COLUMN user_settings.training_days_per_week IS 'Number of training days per week';
COMMENT ON COLUMN user_settings.session_length_minutes IS 'Average training session length in minutes';
COMMENT ON COLUMN user_settings.weight_plan IS 'Weight management plan (gain/lose/maintain)';
COMMENT ON COLUMN user_settings.form_issues IS 'Array of user-reported form issues';
COMMENT ON COLUMN user_settings.injuries IS 'User-reported injuries or imbalances';
COMMENT ON COLUMN user_settings.evaluate_feasibility IS 'Whether coach should evaluate goal feasibility';
COMMENT ON COLUMN user_settings.federation IS 'Competition federation';
COMMENT ON COLUMN user_settings.knee_sleeve IS 'Preferred knee sleeve brand/type';
COMMENT ON COLUMN user_settings.deadlift_style IS 'Deadlift style (sumo or conventional)';
COMMENT ON COLUMN user_settings.squat_stance IS 'Squat stance width (wide/narrow/medium)';
COMMENT ON COLUMN user_settings.add_per_month IS 'Expected monthly strength gain';
COMMENT ON COLUMN user_settings.volume_preference IS 'Training volume preference (low/high/medium)';
COMMENT ON COLUMN user_settings.recovers_from_heavy_deads IS 'Whether user recovers well from heavy deadlifts';
COMMENT ON COLUMN user_settings.height_value IS 'User height value';
COMMENT ON COLUMN user_settings.height_unit IS 'Unit for height (cm or in)';
COMMENT ON COLUMN user_settings.past_competitions IS 'JSONB array of past competition results';
COMMENT ON COLUMN user_settings.feed_visibility IS 'Feed visibility setting (public or passcode)';
COMMENT ON COLUMN user_settings.passcode_hash IS 'Hashed passcode for feed access (bcrypt)';
