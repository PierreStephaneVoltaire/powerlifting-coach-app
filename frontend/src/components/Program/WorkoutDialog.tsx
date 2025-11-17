import React, { useState, useEffect } from 'react';

interface WorkoutDialogProps {
  session: any;
  onClose: () => void;
  onComplete: () => void;
}

interface SetLog {
  set_number: number;
  reps_completed: number;
  weight_kg: number;
  rpe_actual?: number;
  notes?: string;
  video_id?: string;
  completed: boolean;
}

export const WorkoutDialog: React.FC<WorkoutDialogProps> = ({ session, onClose, onComplete }) => {
  const [isStarted, setIsStarted] = useState(!!session.completed_at);
  const [currentExerciseIdx, setCurrentExerciseIdx] = useState(0);
  const [exercises, setExercises] = useState<any[]>([]);
  const [exerciseLogs, setExerciseLogs] = useState<Map<string, SetLog[]>>(new Map());
  const [workoutNotes, setWorkoutNotes] = useState('');
  const [rpeRating, setRpeRating] = useState<number | ''>('');

  useEffect(() => {
    // Initialize exercises with completed sets if any
    const initialExercises = [...session.exercises].map((ex: any, idx: number) => ({
      ...ex,
      order: idx,
    }));
    setExercises(initialExercises);

    // Initialize logs for each exercise
    const initialLogs = new Map();
    initialExercises.forEach((ex: any) => {
      const sets: SetLog[] = [];
      for (let i = 1; i <= ex.sets; i++) {
        sets.push({
          set_number: i,
          reps_completed: 0,
          weight_kg: 0,
          completed: false,
        });
      }
      initialLogs.set(ex.id || `ex-${ex.order}`, sets);
    });
    setExerciseLogs(initialLogs);

    // Load saved progress from localStorage
    loadSavedProgress();
  }, [session]);

  const loadSavedProgress = () => {
    const savedKey = `workout-progress-${session.id}`;
    const saved = localStorage.getItem(savedKey);
    if (saved) {
      try {
        const data = JSON.parse(saved);
        if (data.exerciseLogs) {
          setExerciseLogs(new Map(Object.entries(data.exerciseLogs)));
        }
        if (data.currentExerciseIdx !== undefined) {
          setCurrentExerciseIdx(data.currentExerciseIdx);
        }
        if (data.workoutNotes) {
          setWorkoutNotes(data.workoutNotes);
        }
        if (data.rpeRating) {
          setRpeRating(data.rpeRating);
        }
        setIsStarted(true);
      } catch (error) {
        console.error('Failed to load saved progress:', error);
      }
    }
  };

  const saveProgress = () => {
    const savedKey = `workout-progress-${session.id}`;
    const data = {
      exerciseLogs: Object.fromEntries(exerciseLogs),
      currentExerciseIdx,
      workoutNotes,
      rpeRating,
    };
    localStorage.setItem(savedKey, JSON.stringify(data));
  };

  useEffect(() => {
    if (isStarted) {
      saveProgress();
    }
  }, [exerciseLogs, currentExerciseIdx, workoutNotes, rpeRating, isStarted]);

  const handleStartWorkout = () => {
    setIsStarted(true);
  };

  const currentExercise = exercises[currentExerciseIdx];
  const currentLogs = exerciseLogs.get(currentExercise?.id || `ex-${currentExercise?.order}`) || [];

  const updateSetLog = (setNumber: number, field: keyof SetLog, value: any) => {
    const exId = currentExercise.id || `ex-${currentExercise.order}`;
    const logs = [...(exerciseLogs.get(exId) || [])];
    const setLog = logs.find((log) => log.set_number === setNumber);
    if (setLog) {
      (setLog as any)[field] = value;
      exerciseLogs.set(exId, logs);
      setExerciseLogs(new Map(exerciseLogs));
    }
  };

  const toggleSetCompleted = (setNumber: number) => {
    const exId = currentExercise.id || `ex-${currentExercise.order}`;
    const logs = [...(exerciseLogs.get(exId) || [])];
    const setLog = logs.find((log) => log.set_number === setNumber);
    if (setLog) {
      setLog.completed = !setLog.completed;
      exerciseLogs.set(exId, logs);
      setExerciseLogs(new Map(exerciseLogs));
    }
  };

  const moveExercise = (fromIdx: number, toIdx: number) => {
    if (toIdx < 0 || toIdx >= exercises.length) return;
    const newExercises = [...exercises];
    const [removed] = newExercises.splice(fromIdx, 1);
    newExercises.splice(toIdx, 0, removed);
    setExercises(newExercises);
    setCurrentExerciseIdx(toIdx);
  };

  const handleCompleteWorkout = async () => {
    if (!confirm('Are you sure you want to complete this workout?')) return;

    try {
      // TODO: Submit workout to backend API
      // await apiClient.logWorkout(session.id, exerciseLogs, workoutNotes, rpeRating);

      // Clear saved progress
      localStorage.removeItem(`workout-progress-${session.id}`);

      onComplete();
    } catch (error) {
      console.error('Failed to complete workout:', error);
      alert('Failed to save workout. Please try again.');
    }
  };

  if (!isStarted) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
        <div className="bg-white rounded-lg max-w-2xl w-full p-6">
          <h2 className="text-2xl font-bold mb-4">{session.session_name}</h2>
          <p className="text-gray-600 mb-6">
            📅 {new Date(session.scheduled_date).toLocaleDateString()}
          </p>

          <div className="mb-6">
            <h3 className="font-semibold mb-3">Exercises ({session.exercises?.length || 0})</h3>
            <div className="space-y-2">
              {session.exercises?.map((ex: any, idx: number) => (
                <div key={idx} className="flex items-center gap-3 p-3 bg-gray-50 rounded">
                  <div className="flex-1">
                    <div className="font-medium">{ex.name}</div>
                    <div className="text-sm text-gray-600">
                      {ex.sets}×{ex.reps} @ {ex.intensity || `RPE ${ex.rpe}`}
                      {ex.notes && <span className="ml-2 text-gray-500">• {ex.notes}</span>}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-3">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleStartWorkout}
              className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Start Workout
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-xl font-bold">{session.session_name}</h2>
              <p className="text-sm text-gray-600">
                Exercise {currentExerciseIdx + 1} of {exercises.length}
              </p>
            </div>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Progress Bar */}
          <div className="mt-3 w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all"
              style={{
                width: `${((currentExerciseIdx + 1) / exercises.length) * 100}%`,
              }}
            />
          </div>
        </div>

        {/* Exercise Content */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {currentExercise && (
            <div>
              <div className="flex justify-between items-start mb-4">
                <div className="flex-1">
                  <h3 className="text-2xl font-bold text-gray-900">{currentExercise.name}</h3>
                  <p className="text-gray-600 mt-1">
                    Target: {currentExercise.sets}×{currentExercise.reps} @ {currentExercise.intensity || `RPE ${currentExercise.rpe}`}
                  </p>
                  {currentExercise.notes && (
                    <p className="text-sm text-gray-500 mt-2">💡 {currentExercise.notes}</p>
                  )}
                </div>

                {/* Exercise Reorder Buttons */}
                <div className="flex gap-1">
                  <button
                    onClick={() => moveExercise(currentExerciseIdx, currentExerciseIdx - 1)}
                    disabled={currentExerciseIdx === 0}
                    className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30"
                    title="Move up"
                  >
                    ↑
                  </button>
                  <button
                    onClick={() => moveExercise(currentExerciseIdx, currentExerciseIdx + 1)}
                    disabled={currentExerciseIdx === exercises.length - 1}
                    className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30"
                    title="Move down"
                  >
                    ↓
                  </button>
                </div>
              </div>

              {/* Sets Table */}
              <div className="space-y-2">
                {currentLogs.map((setLog) => (
                  <div
                    key={setLog.set_number}
                    className={`border-2 rounded-lg p-4 ${
                      setLog.completed ? 'border-green-500 bg-green-50' : 'border-gray-200'
                    }`}
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-16 text-center">
                        <div className="font-semibold text-gray-700">Set {setLog.set_number}</div>
                      </div>

                      <div className="flex-1 grid grid-cols-3 gap-3">
                        <div>
                          <label className="text-xs text-gray-600 block mb-1">Weight (kg)</label>
                          <input
                            type="number"
                            step="0.5"
                            value={setLog.weight_kg || ''}
                            onChange={(e) => updateSetLog(setLog.set_number, 'weight_kg', parseFloat(e.target.value) || 0)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                            placeholder="0"
                          />
                        </div>

                        <div>
                          <label className="text-xs text-gray-600 block mb-1">Reps</label>
                          <input
                            type="number"
                            value={setLog.reps_completed || ''}
                            onChange={(e) => updateSetLog(setLog.set_number, 'reps_completed', parseInt(e.target.value) || 0)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                            placeholder="0"
                          />
                        </div>

                        <div>
                          <label className="text-xs text-gray-600 block mb-1">RPE (optional)</label>
                          <input
                            type="number"
                            step="0.5"
                            min="1"
                            max="10"
                            value={setLog.rpe_actual || ''}
                            onChange={(e) => updateSetLog(setLog.set_number, 'rpe_actual', parseFloat(e.target.value) || undefined)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                            placeholder="6.5"
                          />
                        </div>
                      </div>

                      <button
                        onClick={() => toggleSetCompleted(setLog.set_number)}
                        className={`px-4 py-2 rounded font-medium ${
                          setLog.completed
                            ? 'bg-green-600 text-white hover:bg-green-700'
                            : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                        }`}
                      >
                        {setLog.completed ? '✓ Done' : 'Mark Complete'}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer Navigation */}
        <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
          <div className="flex justify-between items-center">
            <button
              onClick={() => setCurrentExerciseIdx(Math.max(0, currentExerciseIdx - 1))}
              disabled={currentExerciseIdx === 0}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-white disabled:opacity-50 disabled:cursor-not-allowed"
            >
              ← Previous Exercise
            </button>

            {currentExerciseIdx < exercises.length - 1 ? (
              <button
                onClick={() => setCurrentExerciseIdx(currentExerciseIdx + 1)}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Next Exercise →
              </button>
            ) : (
              <button
                onClick={handleCompleteWorkout}
                className="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
              >
                Complete Workout
              </button>
            )}
          </div>

          <div className="mt-4 grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs text-gray-600 block mb-1">Overall RPE (1-10)</label>
              <input
                type="number"
                step="0.5"
                min="1"
                max="10"
                value={rpeRating}
                onChange={(e) => setRpeRating(parseFloat(e.target.value) || '')}
                className="w-full px-3 py-2 border border-gray-300 rounded"
                placeholder="How hard was this workout?"
              />
            </div>
            <div>
              <label className="text-xs text-gray-600 block mb-1">Workout Notes</label>
              <input
                type="text"
                value={workoutNotes}
                onChange={(e) => setWorkoutNotes(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded"
                placeholder="How did you feel?"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
