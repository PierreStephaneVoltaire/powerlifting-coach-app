import React, { useState } from 'react';

interface ProgramApprovalViewProps {
  pendingProgram: any;
  currentProgram: any;
  onApprove: () => Promise<void>;
  onReject: () => Promise<void>;
}

export const ProgramApprovalView: React.FC<ProgramApprovalViewProps> = ({
  pendingProgram,
  currentProgram,
  onApprove,
  onReject,
}) => {
  const [isApproving, setIsApproving] = useState(false);
  const [isRejecting, setIsRejecting] = useState(false);

  const handleApprove = async () => {
    setIsApproving(true);
    try {
      await onApprove();
    } finally {
      setIsApproving(false);
    }
  };

  const handleReject = async () => {
    if (!confirm('Are you sure you want to reject this program? You will need to create a new one.')) {
      return;
    }

    setIsRejecting(true);
    try {
      await onReject();
    } finally {
      setIsRejecting(false);
    }
  };

  const pendingData = pendingProgram.pending_program_data || {};
  const phases = pendingData.phases || [];
  const weeklyWorkouts = pendingData.weeklyWorkouts || [];
  const summary = pendingData.summary || {};

  return (
    <div className="max-w-7xl mx-auto p-6">
      <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-6">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-yellow-800">
              Program Awaiting Approval
            </h3>
            <p className="text-sm text-yellow-700 mt-2">
              Your AI coach has generated a new training program. Review it carefully and approve or go back to the chat to make changes.
            </p>
          </div>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 bg-gray-50">
          <h2 className="text-2xl font-bold text-gray-900">{pendingProgram.name}</h2>
          {pendingProgram.description && (
            <p className="text-gray-600 mt-1">{pendingProgram.description}</p>
          )}
          <div className="mt-3 flex flex-wrap gap-4 text-sm text-gray-600">
            <span>📅 {summary.totalWeeks || pendingProgram.weeks_total} weeks</span>
            <span>🏋️ {pendingProgram.days_per_week} days/week</span>
            <span>🎯 Competition: Week {summary.competitionWeek || pendingProgram.weeks_total + 1}</span>
          </div>
        </div>

        <div className="p-6">
          {/* Program Phases */}
          {phases.length > 0 && (
            <div className="mb-8">
              <h3 className="text-lg font-semibold mb-4">Training Phases</h3>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Phase</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Weeks</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Focus</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Characteristics</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {phases.map((phase: any, idx: number) => (
                      <tr key={idx}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{phase.name}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          Week {Math.min(...phase.weeks)}-{Math.max(...phase.weeks)}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">{phase.focus}</td>
                        <td className="px-6 py-4 text-sm text-gray-500">{phase.characteristics}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Weekly Workout Summary */}
          {weeklyWorkouts.length > 0 && (
            <div className="mb-8">
              <h3 className="text-lg font-semibold mb-4">Weekly Workout Overview</h3>
              <div className="space-y-4">
                {weeklyWorkouts.slice(0, 4).map((week: any) => (
                  <div key={week.week} className="border border-gray-200 rounded-lg p-4">
                    <h4 className="font-semibold text-gray-900 mb-3">Week {week.week}</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                      {week.workouts?.map((workout: any, idx: number) => (
                        <div key={idx} className="bg-gray-50 rounded p-3">
                          <div className="text-sm font-medium text-gray-700 mb-2">
                            Day {workout.day}: {workout.name}
                          </div>
                          <div className="text-xs text-gray-600 space-y-1">
                            {workout.exercises?.slice(0, 3).map((ex: any, exIdx: number) => (
                              <div key={exIdx}>
                                • {ex.name}: {ex.sets}x{ex.reps} @ {ex.intensity || ex.rpe}
                              </div>
                            ))}
                            {workout.exercises?.length > 3 && (
                              <div className="text-gray-500 italic">
                                + {workout.exercises.length - 3} more exercises
                              </div>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
                {weeklyWorkouts.length > 4 && (
                  <p className="text-sm text-gray-500 text-center">
                    ... and {weeklyWorkouts.length - 4} more weeks
                  </p>
                )}
              </div>
            </div>
          )}

          {/* Main Lifts Summary */}
          <div className="mb-8">
            <h3 className="text-lg font-semibold mb-4">Main Lifts Progression</h3>
            <p className="text-sm text-gray-600 mb-3">
              This shows the heaviest/most challenging sets for each of the main competition lifts per week.
            </p>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Week</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Squat</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Bench</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Deadlift</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {weeklyWorkouts.slice(0, 12).map((week: any) => {
                    const squatSets = getMainLiftSets(week.workouts, 'squat');
                    const benchSets = getMainLiftSets(week.workouts, 'bench');
                    const deadliftSets = getMainLiftSets(week.workouts, 'deadlift');

                    return (
                      <tr key={week.week}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                          Week {week.week}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{squatSets || '-'}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{benchSets || '-'}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{deadliftSets || '-'}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex justify-end gap-4 pt-6 border-t border-gray-200">
            <button
              onClick={handleReject}
              disabled={isRejecting || isApproving}
              className="px-6 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isRejecting ? 'Rejecting...' : 'Go Back to Chat'}
            </button>
            <button
              onClick={handleApprove}
              disabled={isApproving || isRejecting}
              className="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isApproving ? 'Approving...' : 'Approve & Start Program'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

// Helper function to extract main lift sets from workouts
function getMainLiftSets(workouts: any[], liftType: string): string {
  if (!workouts) return '';

  for (const workout of workouts) {
    const exercises = workout.exercises || [];
    for (const exercise of exercises) {
      const name = exercise.name.toLowerCase();
      const isMainLift =
        (liftType === 'squat' && (name.includes('squat') && name.includes('competition'))) ||
        (liftType === 'bench' && (name.includes('bench') && name.includes('competition'))) ||
        (liftType === 'deadlift' && (name.includes('deadlift') && name.includes('competition')));

      if (isMainLift) {
        return `${exercise.sets}x${exercise.reps} @ ${exercise.intensity || `RPE ${exercise.rpe}`}`;
      }
    }
  }

  return '';
}
