import { FieldPath } from 'react-hook-form';
import { FormData, MIN_COMPETITION_DAYS } from './OnboardingFormTypes';

export const getMinCompetitionDate = () => {
  const date = new Date();
  date.setDate(date.getDate() + MIN_COMPETITION_DAYS);
  return date.toISOString().split('T')[0];
};

export const getWeightClassKg = (weightClass: string | undefined): number => {
  if (!weightClass) return 0;
  if (weightClass.includes('+')) return 1000;
  const match = weightClass.match(/(\d+)/);
  return match ? parseInt(match[0]) : 0;
};

export const weightPlanValidation = (
  value: string | undefined,
  currentWeightKg: number,
  targetKg: number
) => {
  if (!value || value === 'maintain') return true;

  if (targetKg === 0) return true;

  if (value === 'gain' && currentWeightKg > targetKg + 1.0) {
    return `Cannot select 'Gain'. Current weight (${currentWeightKg.toFixed(1)} kg) is more than 1 kg above the target weight class (${targetKg} kg). Select 'Maintain' or 'Lose'.`;
  }

  if (value === 'lose' && currentWeightKg < targetKg - 1.0) {
    return `Cannot select 'Lose'. Current weight (${currentWeightKg.toFixed(1)} kg) is more than 1 kg below the target weight class (${targetKg} kg). Select 'Maintain' or 'Gain'.`;
  }

  return true;
};

export const getStepFields = (
  step: number,
  hasCompeted: boolean,
  heightUnit: string,
  feedVisibility: string
): FieldPath<FormData>[] => {
  let step1Fields: FieldPath<FormData>[] = [
    'has_competed', 'age', 'weight.value', 'weight.unit',
    'height.unit', 'height.value', 'best_squat_kg', 'best_bench_kg', 'best_dead_kg'
  ];

  if (heightUnit === 'in') {
    step1Fields.push('height.feet', 'height.inches');
    step1Fields = step1Fields.filter(f => f !== 'height.value');
  } else {
    step1Fields = step1Fields.filter(f => f !== 'height.feet' && f !== 'height.inches');
  }

  if (hasCompeted) {
    step1Fields.push('best_total_kg', 'comp_pr_date', 'comp_federation');
  }

  switch (step) {
    case 1:
      return step1Fields;
    case 2:
      return ['target_weight_class', 'competition_date', 'squat_goal.value', 'bench_goal.value', 'dead_goal.value', 'most_important_lift', 'least_important_lift'];
    case 3:
      return ['training_days_per_week', 'session_length_minutes', 'volume_preference', 'recovery_rating_squat', 'recovery_rating_bench', 'recovery_rating_dead', 'squat_stance', 'deadlift_style', 'squat_bar_position'];
    case 4:
      const step4Fields: FieldPath<FormData>[] = ['weight_plan', 'feed_visibility', 'knee_sleeve'];
      if (feedVisibility === 'passcode') {
        step4Fields.push('passcode');
      }
      return step4Fields;
    default:
      return [];
  }
};

export const prepareApiPayload = (data: FormData) => {
  const apiPayload = JSON.parse(JSON.stringify(data));

  // Convert height to cm if in imperial
  if (apiPayload.height && apiPayload.height.unit === 'in') {
    const feet = apiPayload.height.feet || 0;
    const inches = apiPayload.height.inches || 0;
    apiPayload.height.value = Math.round((feet * 12 + inches) * 2.54);
    apiPayload.height.unit = 'cm';
  }

  if (apiPayload.height) {
    delete apiPayload.height.feet;
    delete apiPayload.height.inches;
  }

  // Ensure all goal units are kg
  if (apiPayload.squat_goal) apiPayload.squat_goal.unit = 'kg';
  if (apiPayload.bench_goal) apiPayload.bench_goal.unit = 'kg';
  if (apiPayload.dead_goal) apiPayload.dead_goal.unit = 'kg';

  // Clear competition data if hasn't competed
  if (!apiPayload.has_competed) {
    apiPayload.best_total_kg = 0;
    apiPayload.comp_pr_date = null;
    apiPayload.comp_federation = null;
  }

  return apiPayload;
};
