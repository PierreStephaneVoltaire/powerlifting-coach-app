interface AppConfig {
  apiUrl: string;
  authUrl: string;
}

const config: AppConfig = {
  apiUrl: 'https://api.nolift.training',
  authUrl: 'https://auth.nolift.training',
};

export async function loadConfig(): Promise<AppConfig> {
  return config;
}

export function getConfig(): AppConfig {
  return config;
}
