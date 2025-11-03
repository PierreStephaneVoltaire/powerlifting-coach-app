DROP TRIGGER IF EXISTS update_app_settings_updated_at ON app_settings;
DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_app_settings_is_public;
DROP INDEX IF EXISTS idx_app_settings_category;
DROP INDEX IF EXISTS idx_app_settings_key;
DROP INDEX IF EXISTS idx_user_settings_user_id;

DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS user_settings;