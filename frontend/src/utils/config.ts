interface AppConfig {
  apiUrl: string;
  authUrl: string;
}

let config: AppConfig | null = null;

export async function loadConfig(): Promise<AppConfig> {
  if (config) return config;

  try {
    const response = await fetch('/config.json');
    if (response.ok) {
      config = await response.json();
      return config!;
    }
  } catch (e) {
    console.warn('Failed to load config.json, using env vars');
  }

  config = {
    apiUrl: process.env.REACT_APP_API_URL || 'https://api.nolift.training',
    authUrl: process.env.REACT_APP_AUTH_URL || 'https://auth.nolift.training',
  };
  return config;
}

export function getConfig(): AppConfig {
  if (!config) {
    throw new Error('Config not loaded. Call loadConfig() first.');
  }
  return config;
}
