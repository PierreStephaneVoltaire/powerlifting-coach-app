/**
 * Fake Data Service for Dev Mode
 *
 * Why: Allows frontend development without backend dependency
 * Stores all data in localStorage for persistence across page reloads
 */

const STORAGE_PREFIX = 'powercoach_dev_';

interface User {
  id: string;
  email: string;
  name: string;
  user_type: string;
  onboarding_completed: boolean;
}

interface Program {
  id: string;
  program_name: string;
  competition_date?: string;
  exercises: any[];
}

interface Exercise {
  id: string;
  name: string;
  description?: string;
  lift_type: string;
  primary_muscles: string[];
  difficulty?: string;
  equipment_needed: string[];
  demo_video_url?: string;
  form_cues: string[];
}

export class FakeDataService {
  private storage: Storage;

  constructor() {
    this.storage = localStorage;
    this.initializeDefaultData();
  }

  private initializeDefaultData() {
    if (!this.getItem('initialized')) {
      this.setItem('initialized', 'true');
      this.setItem('user', {
        id: 'dev-user-1',
        email: 'dev@powercoach.com',
        name: 'Dev Athlete',
        user_type: 'athlete',
        onboarding_completed: true,
      });

      this.setItem('programs', [{
        id: 'prog-1',
        program_name: 'Competition Prep - 12 Week',
        competition_date: new Date(Date.now() + 90 * 24 * 60 * 60 * 1000).toISOString(),
        current_squat_max: 180,
        current_bench_max: 120,
        current_deadlift_max: 220,
        goal_squat: 190,
        goal_bench: 130,
        goal_deadlift: 230,
      }]);

      this.setItem('exercises', this.getDefaultExercises());
      this.setItem('workoutHistory', []);
      this.setItem('completedSets', []);
    }
  }

  private getItem<T>(key: string): T | null {
    const item = this.storage.getItem(STORAGE_PREFIX + key);
    return item ? JSON.parse(item) : null;
  }

  private setItem(key: string, value: any): void {
    this.storage.setItem(STORAGE_PREFIX + key, JSON.stringify(value));
  }

  private getDefaultExercises(): Exercise[] {
    return [
      {
        id: 'ex-1',
        name: 'Back Squat',
        description: 'Barbell squat with bar on upper back',
        lift_type: 'squat',
        primary_muscles: ['quadriceps', 'glutes'],
        difficulty: 'intermediate',
        equipment_needed: ['barbell', 'squat_rack'],
        demo_video_url: 'https://www.youtube.com/watch?v=ultWZbUMPL8',
        form_cues: ['Brace core', 'Keep chest up', 'Drive through heels'],
      },
      {
        id: 'ex-2',
        name: 'Bench Press',
        description: 'Barbell bench press',
        lift_type: 'bench',
        primary_muscles: ['pectorals', 'triceps'],
        difficulty: 'intermediate',
        equipment_needed: ['barbell', 'bench'],
        demo_video_url: 'https://www.youtube.com/watch?v=rT7DgCr-3pg',
        form_cues: ['Retract scapula', 'Leg drive', 'Touch chest'],
      },
      {
        id: 'ex-3',
        name: 'Deadlift',
        description: 'Conventional deadlift',
        lift_type: 'deadlift',
        primary_muscles: ['hamstrings', 'glutes', 'lower_back'],
        difficulty: 'intermediate',
        equipment_needed: ['barbell'],
        demo_video_url: 'https://www.youtube.com/watch?v=op9kVnSso6Q',
        form_cues: ['Neutral spine', 'Push floor away', 'Lock hips at top'],
      },
    ];
  }

  async login(email: string, password: string): Promise<{ token: string; user: User }> {
    await this.delay(500);
    const user = this.getItem<User>('user') || this.getDefaultUser();
    return {
      token: 'fake-jwt-token-' + Date.now(),
      user,
    };
  }

  async register(data: any): Promise<{ token: string; user: User }> {
    await this.delay(500);
    const user: User = {
      id: 'user-' + Date.now(),
      email: data.email,
      name: data.name,
      user_type: 'athlete',
      onboarding_completed: false,
    };
    this.setItem('user', user);
    return {
      token: 'fake-jwt-token-' + Date.now(),
      user,
    };
  }

  async getCurrentUser(): Promise<User> {
    await this.delay(200);
    return this.getItem<User>('user') || this.getDefaultUser();
  }

  async updateOnboarding(data: any): Promise<void> {
    await this.delay(300);
    const user = this.getItem<User>('user') || this.getDefaultUser();
    user.onboarding_completed = true;
    this.setItem('user', user);
    this.setItem('onboardingData', data);
  }

  async getCurrentProgram(): Promise<any> {
    await this.delay(300);
    const programs = this.getItem<any[]>('programs') || [];
    return { program: programs[0] || null };
  }

  async getExerciseLibrary(liftType?: string): Promise<{ exercises: Exercise[] }> {
    await this.delay(200);
    let exercises = this.getItem<Exercise[]>('exercises') || [];
    if (liftType && liftType !== 'all') {
      exercises = exercises.filter(ex => ex.lift_type === liftType);
    }
    return { exercises };
  }

  async createCustomExercise(exerciseData: any): Promise<Exercise> {
    await this.delay(300);
    const exercises = this.getItem<Exercise[]>('exercises') || [];
    const newExercise: Exercise = {
      id: 'custom-' + Date.now(),
      ...exerciseData,
    };
    exercises.push(newExercise);
    this.setItem('exercises', exercises);
    return newExercise;
  }

  async getVolumeData(startDate: string, endDate: string): Promise<any[]> {
    await this.delay(300);
    // Generate fake volume data
    const data = [];
    const start = new Date(startDate);
    const end = new Date(endDate);
    const diffDays = Math.floor((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));

    for (let i = 0; i < Math.min(diffDays, 30); i++) {
      const date = new Date(start.getTime() + i * 24 * 60 * 60 * 1000);
      data.push({
        date: date.toISOString(),
        exercise_name: 'Back Squat',
        total_volume: Math.random() * 5000 + 3000,
      });
    }
    return data;
  }

  async getE1RMData(startDate: string, endDate: string): Promise<any[]> {
    await this.delay(300);
    const data = [];
    const start = new Date(startDate);
    const end = new Date(endDate);
    const diffDays = Math.floor((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));

    for (let i = 0; i < Math.min(diffDays, 30); i++) {
      const date = new Date(start.getTime() + i * 24 * 60 * 60 * 1000);
      data.push({
        date: date.toISOString(),
        exercise_name: 'Back Squat',
        lift_type: 'squat',
        estimated_1rm: 180 + Math.random() * 20,
      });
    }
    return data;
  }

  async getSessionHistory(startDate?: string, endDate?: string, limit = 50): Promise<any[]> {
    await this.delay(300);
    return this.getItem<any[]>('workoutHistory') || [];
  }

  async getPreviousSets(exerciseName: string, limit = 5): Promise<{ previous_sets: any[] }> {
    await this.delay(200);
    const completedSets = this.getItem<any[]>('completedSets') || [];
    const previousSets = completedSets
      .filter((set: any) => set.exercise_name === exerciseName)
      .slice(0, limit);
    return { previous_sets: previousSets };
  }

  async generateWarmups(workingWeightKg: number, liftType: string): Promise<{ warmup_sets: any[] }> {
    await this.delay(200);
    const percentages = [0, 0.4, 0.5, 0.6, 0.7, 0.85, 0.95];
    const warmupSets = percentages.map((pct, idx) => ({
      set_number: idx + 1,
      weight_kg: Math.round(workingWeightKg * pct / 2.5) * 2.5,
      reps: idx === 0 ? 10 : 5,
      set_type: 'warm_up',
    }));
    return { warmup_sets: warmupSets };
  }

  async logWorkout(workoutData: any): Promise<void> {
    await this.delay(400);
    const history = this.getItem<any[]>('workoutHistory') || [];
    history.unshift({
      id: 'session-' + Date.now(),
      ...workoutData,
      completed_at: new Date().toISOString(),
    });
    this.setItem('workoutHistory', history);

    // Also store individual sets
    const completedSets = this.getItem<any[]>('completedSets') || [];
    workoutData.exercises?.forEach((ex: any) => {
      ex.sets?.forEach((set: any) => {
        completedSets.push({
          ...set,
          exercise_name: ex.exercise_name,
          completed_at: new Date().toISOString(),
        });
      });
    });
    this.setItem('completedSets', completedSets);
  }

  async chatWithAI(message: string): Promise<{ response: string; artifact?: any }> {
    await this.delay(1000);
    return {
      response: `This is a fake AI response in dev mode. You said: "${message}". In production, this would connect to LiteLLM.`,
    };
  }

  clearAllData(): void {
    Object.keys(this.storage)
      .filter(key => key.startsWith(STORAGE_PREFIX))
      .forEach(key => this.storage.removeItem(key));
    this.initializeDefaultData();
  }

  private getDefaultUser(): User {
    return {
      id: 'dev-user-1',
      email: 'dev@powercoach.com',
      name: 'Dev Athlete',
      user_type: 'athlete',
      onboarding_completed: true,
    };
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

export const fakeDataService = new FakeDataService();
