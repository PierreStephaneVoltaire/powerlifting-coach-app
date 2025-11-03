DROP TRIGGER IF EXISTS update_coach_profiles_updated_at ON coach_profiles;
DROP TRIGGER IF EXISTS update_athlete_profiles_updated_at ON athlete_profiles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_coach_athlete_access_code;
DROP INDEX IF EXISTS idx_coach_athlete_access_athlete_id;
DROP INDEX IF EXISTS idx_coach_athlete_access_coach_id;
DROP INDEX IF EXISTS idx_coach_profiles_user_id;
DROP INDEX IF EXISTS idx_athlete_profiles_access_code;
DROP INDEX IF EXISTS idx_athlete_profiles_user_id;
DROP INDEX IF EXISTS idx_users_user_type;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_keycloak_id;

DROP TABLE IF EXISTS coach_athlete_access;
DROP TABLE IF EXISTS coach_profiles;
DROP TABLE IF EXISTS athlete_profiles;
DROP TABLE IF EXISTS users;