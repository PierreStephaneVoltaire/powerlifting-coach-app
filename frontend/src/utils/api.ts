import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { useAuthStore } from '@/store/authStore';
import { AuthTokens } from '@/types';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        const { tokens } = useAuthStore.getState();
        if (tokens?.access_token) {
          config.headers.Authorization = `Bearer ${tokens.access_token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor to handle token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            const { tokens, refreshTokens, logout } = useAuthStore.getState();
            
            if (tokens?.refresh_token) {
              const newTokens = await this.refreshToken(tokens.refresh_token);
              refreshTokens(newTokens);
              
              // Retry original request with new token
              originalRequest.headers.Authorization = `Bearer ${newTokens.access_token}`;
              return this.client(originalRequest);
            }
          } catch (refreshError) {
            // Refresh failed, logout user
            useAuthStore.getState().logout();
            window.location.href = '/login';
          }
        }

        return Promise.reject(error);
      }
    );
  }

  private async refreshToken(refreshToken: string): Promise<AuthTokens> {
    const response = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
      refresh_token: refreshToken,
    });
    return response.data;
  }

  // Auth endpoints
  async login(email: string, password: string) {
    const response = await this.client.post('/api/v1/auth/login', {
      email,
      password,
    });
    return response.data;
  }

  async register(email: string, password: string, name: string, userType: 'athlete' | 'coach') {
    const response = await this.client.post('/api/v1/auth/register', {
      email,
      password,
      name,
      user_type: userType,
    });
    return response.data;
  }

  async logout(refreshToken: string) {
    await this.client.post('/api/v1/auth/logout', {
      refresh_token: refreshToken,
    });
  }

  async getUserInfo() {
    const response = await this.client.get('/api/v1/auth/user');
    return response.data;
  }

  // User endpoints
  async getProfile() {
    const response = await this.client.get('/api/v1/users/profile');
    return response.data;
  }

  async updateAthleteProfile(data: any) {
    const response = await this.client.put('/api/v1/users/athlete/profile', data);
    return response.data;
  }

  async updateCoachProfile(data: any) {
    const response = await this.client.put('/api/v1/users/coach/profile', data);
    return response.data;
  }

  async generateAccessCode(expiresInWeeks?: number) {
    const response = await this.client.post('/api/v1/users/athlete/access-code', {
      expires_in_weeks: expiresInWeeks,
    });
    return response.data;
  }

  async grantCoachAccess(accessCode: string) {
    const response = await this.client.post('/api/v1/users/coach/grant-access', {
      access_code: accessCode,
    });
    return response.data;
  }

  async getMyAthletes() {
    const response = await this.client.get('/api/v1/users/coach/athletes');
    return response.data;
  }

  // Video endpoints
  async getUploadUrl(filename: string, fileSize: number) {
    const response = await this.client.post('/api/v1/videos/upload', {
      filename,
      file_size: fileSize,
    });
    return response.data;
  }

  async completeUpload(videoId: string) {
    const response = await this.client.post(`/api/v1/videos/${videoId}/complete`);
    return response.data;
  }

  async getMyVideos(page = 1, pageSize = 20) {
    const response = await this.client.get('/api/v1/videos', {
      params: { page, page_size: pageSize },
    });
    return response.data;
  }

  async getVideo(videoId: string) {
    const response = await this.client.get(`/api/v1/videos/${videoId}`);
    return response.data;
  }

  async deleteVideo(videoId: string) {
    await this.client.delete(`/api/v1/videos/${videoId}`);
  }

  async getSharedVideo(shareToken: string) {
    const response = await this.client.get(`/api/v1/videos/shared/${shareToken}`);
    return response.data;
  }

  // Settings endpoints
  async getUserSettings() {
    const response = await this.client.get('/api/v1/settings/user');
    return response.data;
  }

  async updateUserSettings(settings: any) {
    const response = await this.client.put('/api/v1/settings/user', settings);
    return response.data;
  }

  async getPublicAppSettings() {
    const response = await this.client.get('/api/v1/settings/app/public');
    return response.data;
  }

  // Upload file to presigned URL
  async uploadFile(url: string, file: File, onProgress?: (progress: number) => void) {
    const response = await axios.put(url, file, {
      headers: {
        'Content-Type': file.type,
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
    return response;
  }
}

export const apiClient = new ApiClient();