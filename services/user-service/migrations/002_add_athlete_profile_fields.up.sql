ALTER TABLE athlete_profiles
ADD COLUMN IF NOT EXISTS bio TEXT,
ADD COLUMN IF NOT EXISTS target_weight_class VARCHAR(20),
ADD COLUMN IF NOT EXISTS preferred_federation VARCHAR(50);
