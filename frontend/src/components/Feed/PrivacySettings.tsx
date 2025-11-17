import React, { useEffect, useState } from 'react';

interface PrivacySettings {
  default_privacy: 'private' | 'coach_only' | 'public';
  allow_coach_share: boolean;
  auto_post_workouts: boolean;
}

export const PrivacySettings: React.FC = () => {
  const [settings, setSettings] = useState<PrivacySettings>({
    default_privacy: 'private',
    allow_coach_share: true,
    auto_post_workouts: false,
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/feed/privacy-settings', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        if (data) {
          setSettings(data);
        }
      }
    } catch (err) {
      console.error('Failed to load privacy settings:', err);
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async () => {
    try {
      setSaving(true);
      setError(null);
      setSuccess(false);

      const response = await fetch('/api/v1/feed/privacy-settings', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(settings),
      });

      if (!response.ok) {
        throw new Error('Failed to save privacy settings');
      }

      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        <p className="text-gray-600 dark:text-gray-400">Loading privacy settings...</p>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
        Feed Privacy Settings
      </h2>

      <div className="space-y-6">
        {/* Default Privacy */}
        <div>
          <label className="block text-sm font-medium text-gray-900 dark:text-white mb-3">
            Default Post Privacy
          </label>
          <div className="space-y-2">
            {[
              { value: 'private', label: 'Private', description: 'Only you can see your posts' },
              { value: 'coach_only', label: 'Coach Only', description: 'Only you and your coach can see' },
              { value: 'public', label: 'Public', description: 'Anyone can see your posts' },
            ].map((option) => (
              <label
                key={option.value}
                className="flex items-start p-4 border border-gray-200 dark:border-gray-700 rounded-lg cursor-pointer
                         hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
              >
                <input
                  type="radio"
                  name="default_privacy"
                  value={option.value}
                  checked={settings.default_privacy === option.value}
                  onChange={(e) => setSettings({ ...settings, default_privacy: e.target.value as any })}
                  className="mt-1 h-4 w-4 text-blue-600 focus:ring-blue-500"
                />
                <div className="ml-3">
                  <div className="font-medium text-gray-900 dark:text-white">{option.label}</div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">{option.description}</div>
                </div>
              </label>
            ))}
          </div>
        </div>

        {/* Auto-post Workouts */}
        <div className="flex items-start justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="flex-1">
            <h3 className="font-medium text-gray-900 dark:text-white mb-1">
              Auto-post Completed Workouts
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Automatically share your workouts to your feed when you complete them
            </p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer ml-4">
            <input
              type="checkbox"
              checked={settings.auto_post_workouts}
              onChange={(e) => setSettings({ ...settings, auto_post_workouts: e.target.checked })}
              className="sr-only peer"
            />
            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300
                          dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700
                          peer-checked:after:translate-x-full peer-checked:after:border-white
                          after:content-[''] after:absolute after:top-[2px] after:left-[2px]
                          after:bg-white after:border-gray-300 after:border after:rounded-full
                          after:h-5 after:w-5 after:transition-all dark:border-gray-600
                          peer-checked:bg-blue-600"></div>
          </label>
        </div>

        {/* Coach Sharing */}
        <div className="flex items-start justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="flex-1">
            <h3 className="font-medium text-gray-900 dark:text-white mb-1">
              Allow Coach to Share
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Let your coach share your posts and progress with their audience (with your permission each time)
            </p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer ml-4">
            <input
              type="checkbox"
              checked={settings.allow_coach_share}
              onChange={(e) => setSettings({ ...settings, allow_coach_share: e.target.checked })}
              className="sr-only peer"
            />
            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300
                          dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700
                          peer-checked:after:translate-x-full peer-checked:after:border-white
                          after:content-[''] after:absolute after:top-[2px] after:left-[2px]
                          after:bg-white after:border-gray-300 after:border after:rounded-full
                          after:h-5 after:w-5 after:transition-all dark:border-gray-600
                          peer-checked:bg-blue-600"></div>
          </label>
        </div>

        {/* Save Button */}
        <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
              <p className="text-red-600 dark:text-red-400 text-sm">{error}</p>
            </div>
          )}
          {success && (
            <div className="mb-4 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
              <p className="text-green-600 dark:text-green-400 text-sm">Settings saved successfully!</p>
            </div>
          )}
          <button
            onClick={saveSettings}
            disabled={saving}
            className="w-full px-4 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium
                     disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {saving ? 'Saving...' : 'Save Privacy Settings'}
          </button>
        </div>
      </div>
    </div>
  );
};
