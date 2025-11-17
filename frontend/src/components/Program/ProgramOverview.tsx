import React, { useEffect, useState } from 'react';
import { apiClient } from '@/utils/api';
import { EnhancedWorkoutDialog } from './EnhancedWorkoutDialog';

interface ProgramOverviewProps {
  program: any;
  onRefresh: () => void;
}

export const ProgramOverview: React.FC<ProgramOverviewProps> = ({ program, onRefresh }) => {
  const [sessions, setSessions] = useState<any[]>([]);
  const [selectedSession, setSelectedSession] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [currentWeek, setCurrentWeek] = useState(1);
  const [exporting, setExporting] = useState(false);

  useEffect(() => {
    loadSessions();
  }, [program.id]);

  const loadSessions = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getProgram(program.id);
      // The backend should return sessions, but for now we'll generate from program_data
      setSessions(generateSessionsFromProgramData(program.program_data));

      // Calculate current week based on start date
      const weeksSinceStart = Math.floor(
        (Date.now() - new Date(program.start_date).getTime()) / (7 * 24 * 60 * 60 * 1000)
      );
      setCurrentWeek(Math.max(1, Math.min(weeksSinceStart + 1, program.weeks_total)));
    } catch (error) {
      console.error('Failed to load sessions:', error);
    } finally {
      setLoading(false);
    }
  };

  const generateSessionsFromProgramData = (programData: any): any[] => {
    // This is a temporary solution - the backend will generate these
    const weeklyWorkouts = programData.weeklyWorkouts || [];
    const allSessions: any[] = [];

    weeklyWorkouts.forEach((week: any) => {
      week.workouts?.forEach((workout: any) => {
        allSessions.push({
          id: `${week.week}-${workout.day}`,
          week_number: week.week,
          day_number: workout.day,
          session_name: workout.name,
          scheduled_date: calculateScheduledDate(program.start_date, week.week, workout.day),
          completed_at: null,
          exercises: workout.exercises || [],
        });
      });
    });

    return allSessions;
  };

  const calculateScheduledDate = (startDate: string, weekNumber: number, dayNumber: number): string => {
    const start = new Date(startDate);
    const daysToAdd = (weekNumber - 1) * 7 + (dayNumber - 1);
    const scheduled = new Date(start);
    scheduled.setDate(scheduled.getDate() + daysToAdd);
    return scheduled.toISOString().split('T')[0];
  };

  const programData = program.program_data || {};
  const phases = programData.phases || [];
  const weeklyWorkouts = programData.weeklyWorkouts || [];

  // Calculate days until competition
  const daysUntilComp = Math.floor(
    (new Date(program.end_date).getTime() - Date.now()) / (24 * 60 * 60 * 1000)
  );
  const weeksUntilComp = Math.ceil(daysUntilComp / 7);

  const handleExportProgram = async () => {
    try {
      setExporting(true);
      const blob = await apiClient.exportProgram(program.id);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${program.name.replace(/\s+/g, '_')}_${new Date().toISOString().split('T')[0]}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export program:', error);
      alert('Failed to export program. Please try again.');
    } finally {
      setExporting(false);
    }
  };

  return (
    <div className="max-w-7xl mx-auto p-6">
      {/* Header Section */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 mb-2">{program.name}</h1>
            {program.description && (
              <p className="text-gray-600">{program.description}</p>
            )}
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleExportProgram}
              disabled={exporting}
              className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 text-sm disabled:opacity-50"
            >
              {exporting ? 'Exporting...' : 'Export to Excel'}
            </button>
            <button
              onClick={() => window.open('/chat', '_blank')}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
            >
              Chat with Coach
            </button>
          </div>
        </div>

        <div className="mt-6 grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-blue-50 rounded-lg p-4">
            <div className="text-sm text-blue-600 font-medium">Current Week</div>
            <div className="text-2xl font-bold text-blue-900">Week {currentWeek}</div>
            <div className="text-xs text-blue-600">of {program.weeks_total}</div>
          </div>

          <div className="bg-green-50 rounded-lg p-4">
            <div className="text-sm text-green-600 font-medium">Until Competition</div>
            <div className="text-2xl font-bold text-green-900">{Math.max(0, weeksUntilComp)}w</div>
            <div className="text-xs text-green-600">{Math.max(0, daysUntilComp)} days</div>
          </div>

          <div className="bg-purple-50 rounded-lg p-4">
            <div className="text-sm text-purple-600 font-medium">Training Days</div>
            <div className="text-2xl font-bold text-purple-900">{program.days_per_week}</div>
            <div className="text-xs text-purple-600">per week</div>
          </div>

          <div className="bg-orange-50 rounded-lg p-4">
            <div className="text-sm text-orange-600 font-medium">Current Phase</div>
            <div className="text-lg font-bold text-orange-900 capitalize">{program.phase}</div>
            <div className="text-xs text-orange-600">
              {getCurrentPhaseInfo(phases, currentWeek)}
            </div>
          </div>
        </div>
      </div>

      {/* Program Phases */}
      {phases.length > 0 && (
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Training Phases</h2>
          <div className="space-y-3">
            {phases.map((phase: any, idx: number) => {
              const isCurrentPhase = currentWeek >= Math.min(...phase.weeks) && currentWeek <= Math.max(...phase.weeks);
              return (
                <div
                  key={idx}
                  className={`border rounded-lg p-4 ${
                    isCurrentPhase ? 'border-blue-500 bg-blue-50' : 'border-gray-200'
                  }`}
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <h3 className={`font-semibold ${isCurrentPhase ? 'text-blue-900' : 'text-gray-900'}`}>
                        {phase.name}
                        {isCurrentPhase && (
                          <span className="ml-2 text-xs bg-blue-600 text-white px-2 py-1 rounded">Current</span>
                        )}
                      </h3>
                      <p className={`text-sm mt-1 ${isCurrentPhase ? 'text-blue-700' : 'text-gray-600'}`}>
                        {phase.focus}
                      </p>
                      <p className={`text-xs mt-2 ${isCurrentPhase ? 'text-blue-600' : 'text-gray-500'}`}>
                        {phase.characteristics}
                      </p>
                    </div>
                    <div className={`text-sm font-medium ${isCurrentPhase ? 'text-blue-600' : 'text-gray-500'}`}>
                      Weeks {Math.min(...phase.weeks)}-{Math.max(...phase.weeks)}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Main Lifts Summary */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Main Lifts Progression</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Week</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Phase</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Squat</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Bench</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Deadlift</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {weeklyWorkouts.map((week: any) => {
                const phaseName = getPhaseForWeek(phases, week.week);
                const isCurrentWeekRow = week.week === currentWeek;

                return (
                  <tr key={week.week} className={isCurrentWeekRow ? 'bg-blue-50' : ''}>
                    <td className="px-4 py-3 text-sm font-medium">
                      Week {week.week}
                      {isCurrentWeekRow && <span className="ml-2 text-blue-600">←</span>}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">{phaseName}</td>
                    <td className="px-4 py-3 text-sm">{getMainLiftInfo(week.workouts, 'squat')}</td>
                    <td className="px-4 py-3 text-sm">{getMainLiftInfo(week.workouts, 'bench')}</td>
                    <td className="px-4 py-3 text-sm">{getMainLiftInfo(week.workouts, 'deadlift')}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Calendar View */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-xl font-semibold mb-4">Training Calendar</h2>
        <div className="mb-4 flex gap-2">
          <button
            onClick={() => setCurrentWeek(Math.max(1, currentWeek - 1))}
            disabled={currentWeek === 1}
            className="px-3 py-1 border border-gray-300 rounded text-sm hover:bg-gray-50 disabled:opacity-50"
          >
            ← Previous Week
          </button>
          <button
            onClick={() => setCurrentWeek(Math.min(program.weeks_total, currentWeek + 1))}
            disabled={currentWeek === program.weeks_total}
            className="px-3 py-1 border border-gray-300 rounded text-sm hover:bg-gray-50 disabled:opacity-50"
          >
            Next Week →
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {sessions
            .filter((session) => session.week_number === currentWeek)
            .sort((a, b) => a.day_number - b.day_number)
            .map((session) => (
              <WorkoutCard
                key={session.id}
                session={session}
                onClick={() => setSelectedSession(session)}
              />
            ))}
        </div>
      </div>

      {/* Workout Dialog */}
      {selectedSession && (
        <EnhancedWorkoutDialog
          session={selectedSession}
          onClose={() => setSelectedSession(null)}
          onComplete={() => {
            setSelectedSession(null);
            loadSessions();
          }}
        />
      )}
    </div>
  );
};

// Helper Components
const WorkoutCard: React.FC<{ session: any; onClick: () => void }> = ({ session, onClick }) => {
  const isCompleted = !!session.completed_at;
  const isPast = new Date(session.scheduled_date) < new Date();

  return (
    <div
      onClick={onClick}
      className={`border-2 rounded-lg p-4 cursor-pointer transition-all hover:shadow-md ${
        isCompleted
          ? 'border-green-500 bg-green-50'
          : isPast
          ? 'border-red-300 bg-red-50'
          : 'border-blue-300 bg-blue-50'
      }`}
    >
      <div className="flex justify-between items-start mb-2">
        <div className="text-sm font-medium text-gray-700">Day {session.day_number}</div>
        <div className={`text-xs px-2 py-1 rounded ${
          isCompleted ? 'bg-green-600 text-white' : isPast ? 'bg-red-600 text-white' : 'bg-blue-600 text-white'
        }`}>
          {isCompleted ? 'Completed' : isPast ? 'Missed' : 'Upcoming'}
        </div>
      </div>
      <h4 className="font-semibold text-gray-900 mb-2">{session.session_name}</h4>
      <div className="text-xs text-gray-600 space-y-1">
        <div>📅 {new Date(session.scheduled_date).toLocaleDateString()}</div>
        <div>💪 {session.exercises?.length || 0} exercises</div>
      </div>
    </div>
  );
};

// Helper Functions
function getCurrentPhaseInfo(phases: any[], currentWeek: number): string {
  const phase = phases.find((p: any) =>
    currentWeek >= Math.min(...p.weeks) && currentWeek <= Math.max(...p.weeks)
  );
  return phase ? phase.name : '';
}

function getPhaseForWeek(phases: any[], weekNumber: number): string {
  const phase = phases.find((p: any) =>
    weekNumber >= Math.min(...p.weeks) && weekNumber <= Math.max(...p.weeks)
  );
  return phase?.name || '-';
}

function getMainLiftInfo(workouts: any[], liftType: string): string {
  if (!workouts) return '-';

  for (const workout of workouts) {
    const exercises = workout.exercises || [];
    for (const exercise of exercises) {
      const name = exercise.name.toLowerCase();
      const isMainLift =
        (liftType === 'squat' && name.includes('squat') && name.includes('competition')) ||
        (liftType === 'bench' && name.includes('bench') && name.includes('competition')) ||
        (liftType === 'deadlift' && name.includes('deadlift') && name.includes('competition'));

      if (isMainLift) {
        return `${exercise.sets}×${exercise.reps} @ ${exercise.intensity || `RPE ${exercise.rpe}`}`;
      }
    }
  }

  return '-';
}
