import React, { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';

import { generateUUID } from '@/utils/uuid';
interface ExerciseSet {
  weight: number;
  reps: number;
  rpe?: number;
  completed: boolean;
}

interface ExerciseSummary {
  name: string;
  sets: ExerciseSet[];
}

export const WorkoutSession: React.FC = () => {
  const navigate = useNavigate();
  const { workoutId } = useParams<{ workoutId: string }>();
  const { user } = useAuthStore();
  const [startTime] = useState(new Date());
  const [exercises, setExercises] = useState<ExerciseSummary[]>([
    {
      name: 'Squat',
      sets: [
        { weight: 0, reps: 0, rpe: undefined, completed: false },
        { weight: 0, reps: 0, rpe: undefined, completed: false },
        { weight: 0, reps: 0, rpe: undefined, completed: false },
      ],
    },
  ]);
  const [notes, setNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSetChange = (exerciseIdx: number, setIdx: number, field: keyof ExerciseSet, value: any) => {
    const newExercises = [...exercises];
    newExercises[exerciseIdx].sets[setIdx] = {
      ...newExercises[exerciseIdx].sets[setIdx],
      [field]: value,
    };
    setExercises(newExercises);
  };

  const handleCompleteWorkout = async () => {
    if (!user || !workoutId) return;

    setIsSubmitting(true);
    setError(null);

    try {
      const endTime = new Date();
      const durationMinutes = Math.round((endTime.getTime() - startTime.getTime()) / 1000 / 60);

      const event = {
        schema_version: '1.0.0',
        event_type: 'workout.completed',
        client_generated_id: generateUUID(),
        user_id: user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          workout_id: workoutId,
          duration_minutes: durationMinutes,
          exercises_summary: exercises,
          notes,
        },
      };

      await apiClient.submitEvent(event);
      console.info('Workout completed', { workout_id: workoutId, duration_minutes: durationMinutes });
      navigate('/program/list');
    } catch (err: any) {
      console.error('Failed to complete workout', err);
      if (err.queued) {
        navigate('/program/list');
      } else {
        setError(err.response?.data?.error || 'Failed to save workout. Please try again.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const addExercise = () => {
    setExercises([
      ...exercises,
      {
        name: '',
        sets: [
          { weight: 0, reps: 0, rpe: undefined, completed: false },
        ],
      },
    ]);
  };

  const addSet = (exerciseIdx: number) => {
    const newExercises = [...exercises];
    newExercises[exerciseIdx].sets.push({
      weight: 0,
      reps: 0,
      rpe: undefined,
      completed: false,
    });
    setExercises(newExercises);
  };

  const calculateDuration = () => {
    const now = new Date();
    const diffMinutes = Math.round((now.getTime() - startTime.getTime()) / 1000 / 60);
    return diffMinutes;
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Workout Session</h2>
            <p className="text-sm text-gray-600 mt-1">
              Duration: {calculateDuration()} minutes
            </p>
          </div>
          <button
            onClick={() => navigate('/program/list')}
            className="text-gray-600 hover:text-gray-900"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="space-y-6">
          {exercises.map((exercise, exerciseIdx) => (
            <div key={exerciseIdx} className="border border-gray-200 rounded-lg p-4">
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Exercise Name
                </label>
                <input
                  type="text"
                  value={exercise.name}
                  onChange={(e) => {
                    const newExercises = [...exercises];
                    newExercises[exerciseIdx].name = e.target.value;
                    setExercises(newExercises);
                  }}
                  placeholder="e.g., Squat, Bench Press"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                />
              </div>

              <div className="space-y-3">
                <div className="grid grid-cols-12 gap-2 text-xs font-medium text-gray-600 px-2">
                  <div className="col-span-1">Set</div>
                  <div className="col-span-3">Weight</div>
                  <div className="col-span-3">Reps</div>
                  <div className="col-span-3">RPE</div>
                  <div className="col-span-2">Done</div>
                </div>

                {exercise.sets.map((set, setIdx) => (
                  <div key={setIdx} className="grid grid-cols-12 gap-2 items-center">
                    <div className="col-span-1 text-center text-sm font-medium text-gray-700">
                      {setIdx + 1}
                    </div>
                    <div className="col-span-3">
                      <input
                        type="number"
                        step="0.5"
                        value={set.weight || ''}
                        onChange={(e) => handleSetChange(exerciseIdx, setIdx, 'weight', parseFloat(e.target.value) || 0)}
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        placeholder="kg"
                      />
                    </div>
                    <div className="col-span-3">
                      <input
                        type="number"
                        value={set.reps || ''}
                        onChange={(e) => handleSetChange(exerciseIdx, setIdx, 'reps', parseInt(e.target.value) || 0)}
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        placeholder="reps"
                      />
                    </div>
                    <div className="col-span-3">
                      <input
                        type="number"
                        step="0.5"
                        min="1"
                        max="10"
                        value={set.rpe || ''}
                        onChange={(e) => handleSetChange(exerciseIdx, setIdx, 'rpe', parseFloat(e.target.value) || undefined)}
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        placeholder="1-10"
                      />
                    </div>
                    <div className="col-span-2 flex justify-center">
                      <input
                        type="checkbox"
                        checked={set.completed}
                        onChange={(e) => handleSetChange(exerciseIdx, setIdx, 'completed', e.target.checked)}
                        className="w-5 h-5 rounded border-gray-300 text-blue-600"
                      />
                    </div>
                  </div>
                ))}

                <button
                  onClick={() => addSet(exerciseIdx)}
                  className="text-sm text-blue-600 hover:text-blue-700 font-medium"
                >
                  + Add Set
                </button>
              </div>
            </div>
          ))}

          <button
            onClick={addExercise}
            className="w-full px-4 py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-gray-400 hover:text-gray-700"
          >
            + Add Exercise
          </button>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Workout Notes
            </label>
            <textarea
              rows={4}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="How did the workout feel? Any adjustments needed?"
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            />
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-4">
            <button
              onClick={() => navigate('/program/list')}
              className="px-6 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleCompleteWorkout}
              disabled={isSubmitting}
              className="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
            >
              {isSubmitting ? 'Saving...' : 'Complete Workout'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
