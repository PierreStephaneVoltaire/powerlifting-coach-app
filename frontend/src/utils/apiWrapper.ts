/**
 * API Wrapper with Dev Mode Support
 *
 * Why: Centralizes dev mode logic - all API calls check dev mode first
 * before deciding whether to use fake data or real backend
 */

import { fakeDataService } from '@/services/fakeDataService';
import { apiClient as realApiClient } from './api';

const isDevMode = (): boolean => {
  return localStorage.getItem('powercoach_dev_mode') === 'true';
};

export const api = new Proxy(realApiClient, {
  get(target, prop: string) {
    const devMode = isDevMode();

    // Why: Route to fake data service when in dev mode
    if (devMode && typeof fakeDataService[prop as keyof typeof fakeDataService] === 'function') {
      return (...args: any[]) => {
        console.log(`[DEV MODE] ${prop}`, args);
        return (fakeDataService[prop as keyof typeof fakeDataService] as any)(...args);
      };
    }

    // Why: Use real API client in production mode
    const value = target[prop as keyof typeof target];
    if (typeof value === 'function') {
      return value.bind(target);
    }
    return value;
  }
});
