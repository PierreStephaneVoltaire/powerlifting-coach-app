import { OnboardingSettings } from '@/types';

export const MAX_WEIGHT_KG = 635;
export const MAX_SQUAT_KG = 700;
export const MAX_BENCH_KG = 700;
export const MAX_DEADLIFT_KG = 510;
export const MAX_TOTAL_KG = 1400;
export const MIN_LIFT_KG = 25;
export const MIN_COMPETITION_DAYS = 14;
export const MAX_HEIGHT_CM = 272;
export const MIN_HEIGHT_CM = 91;
export const SESSION_MIN_LENGTH = 30;

export const CANADIAN_FEDS = [
  { value: 'CPU', label: 'Canadian Powerlifting Union (CPU)' },
  { value: 'OPA', label: 'Ontario Powerlifting Association (OPA)' },
  { value: 'BCPA', label: 'BC Powerlifting Association (BCPA)' },
  { value: 'FQForce', label: 'Fédération Québécoise de Force (FQForce)' },
  { value: 'APU', label: 'Alberta Powerlifting Union (APU)' },
  { value: 'Non-Sanctioned', label: 'Non-Sanctioned / Local Meet' },
];

export interface FormData extends OnboardingSettings {
  has_competed: boolean;
  best_squat_kg: number;
  best_bench_kg: number;
  best_dead_kg: number;
  best_total_kg: number;
  comp_pr_date: string;
  comp_federation: string;
  squat_bar_position: 'high' | 'medium' | 'low' | 'french';
  injuries: string;
  knee_sleeve: string;
}

export interface StepProps {
  control: any;
  errors: any;
  watch: any;
  getValues: any;
}
