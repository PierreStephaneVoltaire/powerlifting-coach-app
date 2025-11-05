import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';

interface Workout {
  workout_id: string;
  date: string;
  name: string;
  exercises: Array<{
    name: string;
    sets: number;
    reps: string;
  }>;
  status: 'pending' | 'in_progress' | 'completed';
  notes?: string;
}

export const WorkoutList: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [workouts] = useState<Workout[]>([]);
  const [compDate] = useState<string | null>(null);

  const calculateWeeksUntilComp = () => {
    if (!compDate) return 0;
    const today = new Date();
    const comp = new Date(compDate);
    const diffTime = comp.getTime() - today.getTime();
    const diffWeeks = Math.ceil(diffTime / (1000 * 60 * 60 * 24 * 7));
    return diffWeeks;
  };

  const handleStartWorkout = async (workoutId: string) => {
    if (!user) return;

    try {
      const event = {
        schema_version: '1.0.0',
        event_type: 'workout.started',
        client_generated_id: crypto.randomUUID(),
        user_id: user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          workout_id: workoutId,
          start_timestamp: new Date().toISOString(),
        },
      };

      await apiClient.submitEvent(event);
      console.info('Workout started', { workout_id: workoutId });
      navigate(`/workout/${workoutId}`);
    } catch (err: any) {
      console.error('Failed to start workout', err);
      if (err.queued) {
        navigate(`/workout/${workoutId}`);
      }
    }
  };

  const weeksUntilComp = calculateWeeksUntilComp();

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">My Training Plan</h2>
          {weeksUntilComp > 0 && (
            <p className="text-sm text-gray-600 mt-1">
              {weeksUntilComp} weeks until competition
            </p>
          )}
        </div>
        <button
          onClick={() => navigate('/program/create')}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          New Plan
        </button>
      </div>

      {workouts.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
            />
          </svg>
          <h3 className="mt-2 text-lg font-medium text-gray-900">No training plan</h3>
          <p className="mt-1 text-sm text-gray-500">
            Create your first training plan to get started
          </p>
          <button
            onClick={() => navigate('/program/create')}
            className="mt-6 px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Create Plan
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          {workouts.map((workout) => (
            <div
              key={workout.workout_id}
              className="bg-white rounded-lg shadow p-6"
            >
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">
                    {workout.name || `Workout - ${new Date(workout.date).toLocaleDateString()}`}
                  </h3>
                  <p className="text-sm text-gray-600 mt-1">
                    {new Date(workout.date).toLocaleDateString('en-US', {
                      weekday: 'long',
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </p>
                </div>
                <span
                  className={`px-3 py-1 rounded-full text-xs font-medium ${
                    workout.status === 'completed'
                      ? 'bg-green-100 text-green-800'
                      : workout.status === 'in_progress'
                      ? 'bg-yellow-100 text-yellow-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}
                >
                  {workout.status === 'completed'
                    ? 'Completed'
                    : workout.status === 'in_progress'
                    ? 'In Progress'
                    : 'Pending'}
                </span>
              </div>

              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">Exercises</h4>
                <div className="space-y-2">
                  {workout.exercises.map((exercise, idx) => (
                    <div key={idx} className="flex items-center text-sm text-gray-600">
                      <span className="w-8 h-8 flex items-center justify-center bg-gray-100 rounded-full mr-3">
                        {idx + 1}
                      </span>
                      <span className="flex-1">{exercise.name}</span>
                      <span className="text-gray-500">
                        {exercise.sets} x {exercise.reps}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {workout.notes && (
                <p className="text-sm text-gray-600 mb-4 italic">{workout.notes}</p>
              )}

              {workout.status === 'pending' && (
                <button
                  onClick={() => handleStartWorkout(workout.workout_id)}
                  className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  Start Workout
                </button>
              )}

              {workout.status === 'in_progress' && (
                <button
                  onClick={() => navigate(`/workout/${workout.workout_id}`)}
                  className="w-full px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700"
                >
                  Continue Workout
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
