import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { useAuthStore } from '@/store/authStore';
import { AuthTokens } from '@/types';
import { offlineQueue } from './offlineQueue';

import { generateUUID } from '@/utils/uuid';
const API_BASE_URL = process.env.REACT_APP_API_URL || '';
const DEFAULT_TIMEOUT = 30000;
const WRITE_TIMEOUT = 60000;

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: DEFAULT_TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
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

              originalRequest.headers.Authorization = `Bearer ${newTokens.access_token}`;
              return this.client(originalRequest);
            }
          } catch (refreshError) {
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

  async getFeed(limit = 20, cursor?: string, visibility = 'public') {
    const params: any = { limit, visibility };
    if (cursor) params.cursor = cursor;

    const response = await this.client.get('/api/v1/feed', { params });
    return response.data;
  }

  async getFeedPost(postId: string) {
    const response = await this.client.get(`/api/v1/feed/${postId}`);
    return response.data;
  }

  async getPostComments(postId: string) {
    const response = await this.client.get(`/api/v1/posts/${postId}/comments`);
    return response.data;
  }

  async getPostLikes(postId: string) {
    const response = await this.client.get(`/api/v1/posts/${postId}/likes`);
    return response.data;
  }

  async submitEvent(event: any, options: { useOfflineQueue?: boolean } = {}) {
    const { useOfflineQueue: shouldUseQueue = true } = options;

    try {
      const response = await this.client.post('/api/v1/notify/events', event, {
        timeout: WRITE_TIMEOUT,
      });
      return response.data;
    } catch (error: any) {
      const isNetworkError = !error.response || error.code === 'ECONNABORTED' || error.code === 'ERR_NETWORK';

      if (shouldUseQueue && isNetworkError) {
        console.info('Network error, queuing event for offline submission', {
          event_type: event.event_type,
          error: error.message,
        });
        await offlineQueue.enqueue(event);
        return { queued: true, id: event.client_generated_id };
      }

      throw error;
    }
  }

  async submitOnboardingSettings(userId: string, settings: any) {
    const event = {
      schema_version: '1.0.0',
      event_type: 'user.settings.submitted',
      client_generated_id: generateUUID(),
      user_id: userId,
      timestamp: new Date().toISOString(),
      source_service: 'frontend',
      data: settings,
    };

    return this.submitEvent(event);
  }

  async submitComment(userId: string, postId: string, commentText: string, parentCommentId?: string) {
    const event = {
      schema_version: '1.0.0',
      event_type: 'comment.created',
      client_generated_id: generateUUID(),
      user_id: userId,
      timestamp: new Date().toISOString(),
      source_service: 'frontend',
      data: {
        post_id: postId,
        parent_comment_id: parentCommentId || null,
        comment_text: commentText,
      },
    };

    return this.submitEvent(event);
  }

  async submitLike(userId: string, targetType: string, targetId: string, action: 'like' | 'unlike') {
    const event = {
      schema_version: '1.0.0',
      event_type: 'interaction.liked',
      client_generated_id: generateUUID(),
      user_id: userId,
      timestamp: new Date().toISOString(),
      source_service: 'frontend',
      data: {
        target_type: targetType,
        target_id: targetId,
        action,
      },
    };

    return this.submitEvent(event);
  }

  async submitFeedAccessAttempt(userId: string, feedOwnerID: string, passcode: string) {
    const event = {
      schema_version: '1.0.0',
      event_type: 'feed.access.attempt',
      client_generated_id: generateUUID(),
      user_id: userId,
      timestamp: new Date().toISOString(),
      source_service: 'frontend',
      data: {
        feed_owner_id: feedOwnerID,
        passcode,
      },
    };

    return this.submitEvent(event);
  }

  async getConversationMessages(conversationId: string) {
    const response = await this.client.get(`/api/v1/dm/conversations/${conversationId}/messages`);
    return response.data;
  }

  async getActiveProgram() {
    const response = await this.client.get('/api/v1/programs/active');
    return response.data;
  }

  async getPendingProgram() {
    const response = await this.client.get('/api/v1/programs/pending');
    return response.data;
  }

  async createProgramFromChat(programData: any) {
    const response = await this.client.post('/api/v1/programs/from-chat', programData);
    return response.data;
  }

  async approveProgram(programId: string) {
    const response = await this.client.post(`/api/v1/programs/${programId}/approve`);
    return response.data;
  }

  async rejectProgram(programId: string) {
    const response = await this.client.post(`/api/v1/programs/${programId}/reject`);
    return response.data;
  }

  async getProgram(programId: string) {
    const response = await this.client.get(`/api/v1/programs/${programId}`);
    return response.data;
  }

  async getMyPrograms() {
    const response = await this.client.get('/api/v1/programs');
    return response.data;
  }

  async getPendingEventsCount(): Promise<number> {
    return offlineQueue.getPendingCount();
  }

  startOfflineQueueProcessor() {
    offlineQueue.startAutoProcess();
  }

  async getPreviousSets(exerciseName: string, limit = 5) {
    const response = await this.client.get(`/api/v1/exercises/${encodeURIComponent(exerciseName)}/previous`, {
      params: { limit }
    });
    return response.data;
  }

  async generateWarmups(workingWeightKg: number, liftType: string) {
    const response = await this.client.post('/api/v1/exercises/warmups/generate', {
      working_weight_kg: workingWeightKg,
      lift_type: liftType,
    });
    return response.data;
  }

  async getExerciseLibrary(liftType?: string) {
    const params = liftType ? { lift_type: liftType } : {};
    const response = await this.client.get('/api/v1/exercises/library', { params });
    return response.data;
  }

  async createCustomExercise(exerciseData: any) {
    const response = await this.client.post('/api/v1/exercises/library', exerciseData);
    return response.data;
  }

  async getWorkoutTemplates() {
    const response = await this.client.get('/api/v1/templates/workouts');
    return response.data;
  }

  async createWorkoutTemplate(templateData: any) {
    const response = await this.client.post('/api/v1/templates/workouts', templateData);
    return response.data;
  }

  async getVolumeData(startDate: string, endDate: string, exerciseName?: string) {
    const response = await this.client.post('/api/v1/analytics/volume', {
      start_date: startDate,
      end_date: endDate,
      exercise_name: exerciseName,
    });
    return response.data;
  }

  async getE1RMData(startDate: string, endDate: string, liftType?: string) {
    const response = await this.client.post('/api/v1/analytics/e1rm', {
      start_date: startDate,
      end_date: endDate,
      lift_type: liftType,
    });
    return response.data;
  }

  async getSessionHistory(startDate?: string, endDate?: string, limit = 50) {
    const params: any = { limit };
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;

    const response = await this.client.get('/api/v1/sessions/history', { params });
    return response.data;
  }

  async deleteSession(sessionId: string, reason?: string) {
    const response = await this.client.delete(`/api/v1/sessions/${sessionId}`, {
      data: { reason }
    });
    return response.data;
  }

  async proposeChange(programId: string, changes: any, description?: string) {
    const response = await this.client.post('/api/v1/programs/changes/propose', {
      program_id: programId,
      proposed_changes: changes,
      change_description: description,
    });
    return response.data;
  }

  async getPendingChanges(programId: string) {
    const response = await this.client.get(`/api/v1/programs/${programId}/changes/pending`);
    return response.data;
  }

  async applyChange(changeId: string) {
    const response = await this.client.post(`/api/v1/programs/changes/${changeId}/apply`);
    return response.data;
  }

  async rejectChange(changeId: string) {
    const response = await this.client.post(`/api/v1/programs/changes/${changeId}/reject`);
    return response.data;
  }

  async chatWithAI(message: string, programId?: string, coachContextEnable = false) {
    const response = await this.client.post('/api/v1/programs/chat', {
      message,
      program_id: programId,
      coach_context_enable: coachContextEnable,
    });
    return response.data;
  }
}

export const apiClient = new ApiClient();

// Why: Re-export for dev mode support - wrapper handles routing to fake data
export { api } from './apiWrapper';