import { differenceInDays, differenceInWeeks } from 'date-fns';

interface CompPrepData {
  competition_date?: string;
  competition_name?: string;
  weight_class?: number;
  current_squat_max?: number;
  current_bench_max?: number;
  current_deadlift_max?: number;
  goal_squat?: number;
  goal_bench?: number;
  goal_deadlift?: number;
  current_bodyweight?: number;
  qualifying_total?: number;
  target_total?: number;
}

export const useCompPrepCalculations = (data: CompPrepData) => {
  const calculateDaysUntilComp = () => {
    if (!data.competition_date) return null;
    return differenceInDays(new Date(data.competition_date), new Date());
  };

  const calculateWeeksUntilComp = () => {
    if (!data.competition_date) return null;
    return differenceInWeeks(new Date(data.competition_date), new Date());
  };

  const calculateCurrentTotal = () => {
    return (data.current_squat_max || 0) + (data.current_bench_max || 0) + (data.current_deadlift_max || 0);
  };

  const calculateGoalTotal = () => {
    return (data.goal_squat || 0) + (data.goal_bench || 0) + (data.goal_deadlift || 0);
  };

  const calculateReadinessScore = () => {
    const daysUntil = calculateDaysUntilComp();
    if (!daysUntil) return 0;

    const currentTotal = calculateCurrentTotal();
    const goalTotal = calculateGoalTotal();

    const progressScore = Math.min((currentTotal / goalTotal) * 60, 60);

    let timeScore = 40;
    const weeksUntil = calculateWeeksUntilComp() || 0;
    if (weeksUntil < 4) timeScore = 20;
    else if (weeksUntil > 16) timeScore = 30;

    return Math.round(progressScore + timeScore);
  };

  const suggestOpener = (currentMax: number, percentage = 0.90) => {
    return Math.floor(currentMax * percentage / 2.5) * 2.5;
  };

  const getProgressPercentage = (current: number, goal: number) => {
    return Math.min((current / goal) * 100, 100);
  };

  const getProgressColor = (percentage: number) => {
    if (percentage >= 95) return 'text-green-600';
    if (percentage >= 85) return 'text-yellow-600';
    return 'text-blue-600';
  };

  return {
    daysUntil: calculateDaysUntilComp(),
    weeksUntil: calculateWeeksUntilComp(),
    currentTotal: calculateCurrentTotal(),
    goalTotal: calculateGoalTotal(),
    readinessScore: calculateReadinessScore(),
    suggestOpener,
    getProgressPercentage,
    getProgressColor,
  };
};
