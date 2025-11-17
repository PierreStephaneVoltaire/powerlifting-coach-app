import React, { useState, useEffect } from 'react';
import { apiClient } from '../../utils/api';
import { format, differenceInDays, differenceInWeeks } from 'date-fns';

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

export const CompPrepDashboard: React.FC = () => {
  const [data, setData] = useState<CompPrepData>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCompPrepData();
  }, []);

  const fetchCompPrepData = async () => {
    setLoading(true);
    try {
      // Fetch current program to get competition date
      const programResponse = await apiClient.get('/programs/current');
      const program = programResponse.data.program;

      // Mock data for now - in production this would come from athlete profile and program
      setData({
        competition_date: program?.competition_date,
        competition_name: program?.program_name || 'Upcoming Meet',
        weight_class: 93, // kg
        current_squat_max: 180,
        current_bench_max: 120,
        current_deadlift_max: 220,
        goal_squat: 190,
        goal_bench: 130,
        goal_deadlift: 230,
        current_bodyweight: 89.5,
        qualifying_total: 500,
        target_total: 550,
      });
    } catch (error) {
      console.error('Failed to fetch comp prep data:', error);
    } finally {
      setLoading(false);
    }
  };

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

    // Simple readiness formula:
    // - Progress toward goal (0-60 points)
    // - Time remaining appropriateness (0-40 points)
    const progressScore = Math.min((currentTotal / goalTotal) * 60, 60);

    // Optimal is 8-12 weeks out
    let timeScore = 40;
    const weeksUntil = calculateWeeksUntilComp() || 0;
    if (weeksUntil < 4) timeScore = 20; // Too close
    else if (weeksUntil > 16) timeScore = 30; // Too far

    return Math.round(progressScore + timeScore);
  };

  const suggestOpener = (currentMax: number, percentage = 0.90) => {
    return Math.floor(currentMax * percentage / 2.5) * 2.5; // Round to nearest 2.5kg
  };

  const getProgressPercentage = (current: number, goal: number) => {
    return Math.min((current / goal) * 100, 100);
  };

  const getProgressColor = (percentage: number) => {
    if (percentage >= 95) return 'text-green-600';
    if (percentage >= 85) return 'text-yellow-600';
    return 'text-blue-600';
  };

  const daysUntil = calculateDaysUntilComp();
  const weeksUntil = calculateWeeksUntilComp();
  const currentTotal = calculateCurrentTotal();
  const goalTotal = calculateGoalTotal();
  const readinessScore = calculateReadinessScore();

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (!data.competition_date) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center">
          <h2 className="text-2xl font-bold mb-4 text-gray-900 dark:text-white">No Competition Scheduled</h2>
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            Set a competition date in your program to track your prep progress
          </p>
          <button className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            Schedule Competition
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Header with Countdown */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg shadow-lg p-8 mb-8 text-white">
        <div className="text-center">
          <h1 className="text-4xl font-bold mb-2">{data.competition_name}</h1>
          <p className="text-xl mb-4">{format(new Date(data.competition_date), 'MMMM dd, yyyy')}</p>
          <div className="flex justify-center gap-8">
            <div>
              <div className="text-5xl font-bold">{weeksUntil}</div>
              <div className="text-sm uppercase tracking-wide">Weeks Out</div>
            </div>
            <div>
              <div className="text-5xl font-bold">{daysUntil}</div>
              <div className="text-sm uppercase tracking-wide">Days Out</div>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        {/* Readiness Score */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Readiness Score</h3>
          <div className="relative w-32 h-32 mx-auto">
            <svg className="transform -rotate-90 w-32 h-32">
              <circle
                cx="64"
                cy="64"
                r="56"
                stroke="currentColor"
                strokeWidth="8"
                fill="transparent"
                className="text-gray-200 dark:text-gray-700"
              />
              <circle
                cx="64"
                cy="64"
                r="56"
                stroke="currentColor"
                strokeWidth="8"
                fill="transparent"
                strokeDasharray={`${2 * Math.PI * 56}`}
                strokeDashoffset={`${2 * Math.PI * 56 * (1 - readinessScore / 100)}`}
                className={readinessScore >= 80 ? 'text-green-600' : readinessScore >= 60 ? 'text-yellow-600' : 'text-blue-600'}
                strokeLinecap="round"
              />
            </svg>
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-3xl font-bold text-gray-900 dark:text-white">{readinessScore}</span>
            </div>
          </div>
          <p className="text-center mt-4 text-sm text-gray-600 dark:text-gray-400">
            {readinessScore >= 80 ? 'Peak Ready' : readinessScore >= 60 ? 'On Track' : 'Build Phase'}
          </p>
        </div>

        {/* Current vs Goal Total */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">Total (SBD)</h3>
          <div className="flex items-baseline gap-2 mb-4">
            <span className="text-4xl font-bold text-gray-900 dark:text-white">{currentTotal}</span>
            <span className="text-2xl text-gray-500 dark:text-gray-400">/ {goalTotal} kg</span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3 mb-2">
            <div
              className="bg-blue-600 h-3 rounded-full transition-all"
              style={{ width: `${getProgressPercentage(currentTotal, goalTotal)}%` }}
            />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            {goalTotal - currentTotal} kg to goal
          </p>
          {data.qualifying_total && (
            <p className="text-sm text-green-600 dark:text-green-400 mt-2">
              ✓ Qualified (need {data.qualifying_total} kg)
            </p>
          )}
        </div>

        {/* Weight Class */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">Weight Class</h3>
          <div className="flex items-baseline gap-2 mb-4">
            <span className="text-4xl font-bold text-gray-900 dark:text-white">{data.current_bodyweight}</span>
            <span className="text-2xl text-gray-500 dark:text-gray-400">/ {data.weight_class} kg</span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3 mb-2">
            <div
              className={`h-3 rounded-full transition-all ${
                (data.current_bodyweight || 0) > (data.weight_class || 0) ? 'bg-red-600' : 'bg-green-600'
              }`}
              style={{ width: `${getProgressPercentage(data.current_bodyweight || 0, data.weight_class || 1)}%` }}
            />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            {(data.weight_class || 0) - (data.current_bodyweight || 0) >= 0
              ? `${((data.weight_class || 0) - (data.current_bodyweight || 0)).toFixed(1)} kg under limit`
              : `${Math.abs((data.weight_class || 0) - (data.current_bodyweight || 0)).toFixed(1)} kg OVER limit`}
          </p>
        </div>
      </div>

      {/* Individual Lift Progress Rings */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {/* Squat */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Squat</h3>
          <div className="flex items-center justify-between mb-4">
            <div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">{data.current_squat_max} kg</div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Goal: {data.goal_squat} kg</div>
            </div>
            <div className="relative w-20 h-20">
              <svg className="transform -rotate-90 w-20 h-20">
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="6"
                  fill="transparent"
                  className="text-gray-200 dark:text-gray-700"
                />
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="6"
                  fill="transparent"
                  strokeDasharray={`${2 * Math.PI * 36}`}
                  strokeDashoffset={`${2 * Math.PI * 36 * (1 - getProgressPercentage(data.current_squat_max || 0, data.goal_squat || 1) / 100)}`}
                  className="text-red-600"
                  strokeLinecap="round"
                />
              </svg>
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-sm font-bold text-gray-900 dark:text-white">
                  {Math.round(getProgressPercentage(data.current_squat_max || 0, data.goal_squat || 1))}%
                </span>
              </div>
            </div>
          </div>
          <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">Suggested Opener:</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {suggestOpener(data.current_squat_max || 0)} kg
              </span>
            </div>
          </div>
        </div>

        {/* Bench Press */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Bench Press</h3>
          <div className="flex items-center justify-between mb-4">
            <div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">{data.current_bench_max} kg</div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Goal: {data.goal_bench} kg</div>
            </div>
            <div className="relative w-20 h-20">
              <svg className="transform -rotate-90 w-20 h-20">
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="6"
                  fill="transparent"
                  className="text-gray-200 dark:text-gray-700"
                />
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="6"
                  fill="transparent"
                  strokeDasharray={`${2 * Math.PI * 36}`}
                  strokeDashoffset={`${2 * Math.PI * 36 * (1 - getProgressPercentage(data.current_bench_max || 0, data.goal_bench || 1) / 100)}`}
                  className="text-blue-600"
                  strokeLinecap="round"
                />
              </svg>
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-sm font-bold text-gray-900 dark:text-white">
                  {Math.round(getProgressPercentage(data.current_bench_max || 0, data.goal_bench || 1))}%
                </span>
              </div>
            </div>
          </div>
          <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">Suggested Opener:</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {suggestOpener(data.current_bench_max || 0)} kg
              </span>
            </div>
          </div>
        </div>

        {/* Deadlift */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Deadlift</h3>
          <div className="flex items-center justify-between mb-4">
            <div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">{data.current_deadlift_max} kg</div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Goal: {data.goal_deadlift} kg</div>
            </div>
            <div className="relative w-20 h-20">
              <svg className="transform -rotate-90 w-20 h-20">
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="6"
                  fill="transparent"
                  className="text-gray-200 dark:text-gray-700"
                />
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="6"
                  fill="transparent"
                  strokeDasharray={`${2 * Math.PI * 36}`}
                  strokeDashoffset={`${2 * Math.PI * 36 * (1 - getProgressPercentage(data.current_deadlift_max || 0, data.goal_deadlift || 1) / 100)}`}
                  className="text-green-600"
                  strokeLinecap="round"
                />
              </svg>
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-sm font-bold text-gray-900 dark:text-white">
                  {Math.round(getProgressPercentage(data.current_deadlift_max || 0, data.goal_deadlift || 1))}%
                </span>
              </div>
            </div>
          </div>
          <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
            <div className="flex justify-between text-sm">
              <span className="text-gray-600 dark:text-gray-400">Suggested Opener:</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {suggestOpener(data.current_deadlift_max || 0)} kg
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Attempt Strategy Table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h3 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Competition Attempt Strategy</h3>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Lift
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Opener (90%)
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  2nd Attempt (95%)
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  3rd Attempt (Goal)
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <tr>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                  Squat
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {suggestOpener(data.current_squat_max || 0)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {suggestOpener(data.current_squat_max || 0, 0.95)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-blue-600">
                  {data.goal_squat} kg
                </td>
              </tr>
              <tr>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                  Bench Press
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {suggestOpener(data.current_bench_max || 0)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {suggestOpener(data.current_bench_max || 0, 0.95)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-blue-600">
                  {data.goal_bench} kg
                </td>
              </tr>
              <tr>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                  Deadlift
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {suggestOpener(data.current_deadlift_max || 0)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {suggestOpener(data.current_deadlift_max || 0, 0.95)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-blue-600">
                  {data.goal_deadlift} kg
                </td>
              </tr>
              <tr className="bg-gray-50 dark:bg-gray-700">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-bold text-gray-900 dark:text-white">
                  TOTAL
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900 dark:text-white">
                  {suggestOpener(data.current_squat_max || 0) +
                    suggestOpener(data.current_bench_max || 0) +
                    suggestOpener(data.current_deadlift_max || 0)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900 dark:text-white">
                  {suggestOpener(data.current_squat_max || 0, 0.95) +
                    suggestOpener(data.current_bench_max || 0, 0.95) +
                    suggestOpener(data.current_deadlift_max || 0, 0.95)} kg
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-bold text-blue-600">
                  {goalTotal} kg
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
