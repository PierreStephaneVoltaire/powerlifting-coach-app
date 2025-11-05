export interface User {
  id: string;
  email: string;
  name: string;
  user_type: 'athlete' | 'coach';
  created_at: string;
  updated_at: string;
}

export interface AthleteProfile {
  id: string;
  user_id: string;
  weight_kg?: number;
  experience_level?: 'beginner' | 'intermediate' | 'advanced' | 'elite';
  competition_date?: string;
  access_code?: string;
  access_code_expires_at?: string;
  squat_max_kg?: number;
  bench_max_kg?: number;
  deadlift_max_kg?: number;
  training_frequency?: number;
  goals?: string;
  injuries?: string;
  created_at: string;
  updated_at: string;
}

export interface CoachProfile {
  id: string;
  user_id: string;
  bio?: string;
  certifications?: string[];
  years_experience?: number;
  specializations?: string[];
  hourly_rate?: number;
  created_at: string;
  updated_at: string;
}

export interface UserResponse {
  user: User;
  athlete_profile?: AthleteProfile;
  coach_profile?: CoachProfile;
}

export interface Video {
  id: string;
  athlete_id: string;
  filename: string;
  original_filename: string;
  file_size: number;
  content_type: string;
  duration_seconds?: number;
  original_url?: string;
  processed_url?: string;
  thumbnail_url?: string;
  public_share_token: string;
  status: 'uploading' | 'processing' | 'ready' | 'failed';
  processing_error?: string;
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
  processed_at?: string;
}

export interface FormFeedback {
  id: string;
  video_id: string;
  feedback_text: string;
  confidence_score?: number;
  issues: FormIssue[];
  ai_model?: string;
  created_at: string;
}

export interface FormIssue {
  type: string;
  description: string;
  timestamp_seconds?: number;
  severity?: number;
}

export interface Program {
  id: string;
  athlete_id: string;
  name: string;
  description: string;
  start_date: string;
  end_date: string;
  program_data: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface TrainingSession {
  id: string;
  program_id: string;
  athlete_id: string;
  scheduled_date: string;
  completed_at?: string;
  exercises: Exercise[];
  notes?: string;
}

export interface Exercise {
  id: string;
  lift_type: 'squat' | 'bench' | 'deadlift' | 'accessory';
  name: string;
  sets: Set[];
  notes?: string;
}

export interface Set {
  reps: number;
  weight_kg: number;
  rpe?: number;
  completed: boolean;
  video_id?: string;
}

export interface UserSettings {
  id: string;
  user_id: string;
  theme: 'light' | 'dark' | 'auto';
  language: string;
  timezone: string;
  units: 'metric' | 'imperial';
  notifications: NotificationSettings;
  privacy: PrivacySettings;
  training_preferences: TrainingPreferences;
  created_at: string;
  updated_at: string;
}

export interface NotificationSettings {
  email: boolean;
  push: boolean;
  sms: boolean;
}

export interface PrivacySettings {
  profile_public: boolean;
  videos_public: boolean;
}

export interface TrainingPreferences {
  preferred_training_days?: string[];
  session_duration_mins?: number;
  rest_days_between?: number;
  max_sets_per_exercise?: number;
  preferred_time_of_day?: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  user_type: 'athlete' | 'coach';
}

export interface ApiError {
  error: string;
  message?: string;
  details?: Record<string, any>;
}

export interface VideoListResponse {
  videos: Video[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface UploadResponse {
  video_id: string;
  upload_url: string;
  expires_at: string;
}

export interface FeedPost {
  id: string;
  post_id: string;
  user_id: string;
  video_id?: string;
  visibility: 'public' | 'passcode';
  movement_label: string;
  weight?: {
    value: number;
    unit: 'kg' | 'lb';
  };
  rpe?: number;
  comment_text: string;
  comments_count: number;
  likes_count: number;
  created_at: string;
  updated_at: string;
}

export interface FeedResponse {
  posts: FeedPost[];
  next_cursor?: string;
}

export interface Comment {
  id: string;
  comment_id: string;
  post_id: string;
  user_id: string;
  parent_comment_id?: string;
  comment_text: string;
  created_at: string;
  updated_at: string;
}

export interface CommentsResponse {
  comments: Comment[];
}

export interface Like {
  id: string;
  user_id: string;
  target_type: string;
  target_id: string;
  created_at: string;
}

export interface LikesResponse {
  likes: Like[];
}

export interface OnboardingSettings {
  weight: {
    value: number;
    unit: 'kg' | 'lb';
  };
  age: number;
  target_weight_class?: string;
  weeks_until_comp?: number;
  squat_goal?: {
    value: number;
    unit: 'kg' | 'lb';
  };
  bench_goal?: {
    value: number;
    unit: 'kg' | 'lb';
  };
  dead_goal?: {
    value: number;
    unit: 'kg' | 'lb';
  };
  most_important_lift?: 'squat' | 'bench' | 'deadlift';
  least_important_lift?: 'squat' | 'bench' | 'deadlift';
  recovery_rating_squat?: number;
  recovery_rating_bench?: number;
  recovery_rating_dead?: number;
  training_days_per_week: number;
  session_length_minutes?: number;
  weight_plan?: 'gain' | 'lose' | 'maintain';
  form_issues?: string[];
  injuries?: string;
  evaluate_feasibility?: boolean;
  federation?: string;
  knee_sleeve?: string;
  deadlift_style?: 'sumo' | 'conventional';
  squat_stance?: 'wide' | 'narrow' | 'medium';
  add_per_month?: '2.5kg' | '5kg' | 'none';
  volume_preference?: 'low' | 'high' | 'medium';
  recovers_from_heavy_deads?: boolean;
  height?: {
    value: number;
    unit: 'cm' | 'in';
  };
  past_competitions?: Array<{
    date_range: string;
    squat_attempts: number[];
    bench_attempts: number[];
    deadlift_attempts: number[];
    total: number;
  }>;
  feed_visibility?: 'public' | 'passcode';
  passcode?: string;
}

export interface Event {
  schema_version: string;
  event_type: string;
  client_generated_id: string;
  user_id: string;
  timestamp: string;
  source_service: string;
  data: Record<string, any>;
}