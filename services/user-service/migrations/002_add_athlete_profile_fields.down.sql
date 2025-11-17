ALTER TABLE athlete_profiles
DROP COLUMN IF EXISTS bio,
DROP COLUMN IF EXISTS target_weight_class,
DROP COLUMN IF EXISTS preferred_federation;
