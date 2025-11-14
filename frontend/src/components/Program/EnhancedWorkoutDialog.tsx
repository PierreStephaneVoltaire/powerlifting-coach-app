import React, { useState, useEffect } from 'react';
import { api } from '../../utils/api';

interface EnhancedWorkoutDialogProps {
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
  set_type: 'warm_up' | 'working' | 'backoff' | 'amrap' | 'failure' | 'drop_set' | 'cluster' | 'pause' | 'tempo' | 'custom';
  media_urls?: string[];
  exercise_notes?: string;
}

export const EnhancedWorkoutDialog: React.FC<EnhancedWorkoutDialogProps> = ({ session, onClose, onComplete }) => {
  const [isStarted, setIsStarted] = useState(!!session.completed_at);
  const [currentExerciseIdx, setCurrentExerciseIdx] = useState(0);
  const [exercises, setExercises] = useState<any[]>([]);
  const [exerciseLogs, setExerciseLogs] = useState<Map<string, SetLog[]>>(new Map());
  const [workoutNotes, setWorkoutNotes] = useState('');
  const [rpeRating, setRpeRating] = useState<number | ''>('');
  const [showWarmupGenerator, setShowWarmupGenerator] = useState(false);
  const [previousSets, setPreviousSets] = useState<any[]>([]);
  const [exerciseNotes, setExerciseNotes] = useState('');

  useEffect(() => {
    const initialExercises = [...session.exercises].map((ex: any, idx: number) => ({
      ...ex,
      order: idx,
    }));
    setExercises(initialExercises);

    const initialLogs = new Map();
    initialExercises.forEach((ex: any) => {
      const sets: SetLog[] = [];
      for (let i = 1; i <= ex.sets; i++) {
        sets.push({
          set_number: i,
          reps_completed: 0,
          weight_kg: 0,
          completed: false,
          set_type: 'working',
        });
      }
      initialLogs.set(ex.id || `ex-${ex.order}`, sets);
    });
    setExerciseLogs(initialLogs);

    loadSavedProgress();
  }, [session]);

  useEffect(() => {
    if (currentExercise && isStarted) {
      fetchPreviousSets(currentExercise.name);
    }
  }, [currentExerciseIdx, isStarted]);

  const fetchPreviousSets = async (exerciseName: string) => {
    try {
      const response = await api.get(`/exercises/${encodeURIComponent(exerciseName)}/previous?limit=1`);
      if (response.data.previous_sets && response.data.previous_sets.length > 0) {
        setPreviousSets(response.data.previous_sets);
      }
    } catch (error) {
      console.error('Failed to fetch previous sets:', error);
    }
  };

  const autofillFromPrevious = () => {
    if (previousSets.length === 0) return;

    const exId = currentExercise.id || `ex-${currentExercise.order}`;
    const logs = [...(exerciseLogs.get(exId) || [])];

    previousSets.forEach((prevSet: any) => {
      const setLog = logs.find(log => log.set_number === prevSet.set_number);
      if (setLog && !setLog.completed) {
        setLog.weight_kg = prevSet.weight_kg;
        setLog.reps_completed = prevSet.reps_completed;
        setLog.rpe_actual = prevSet.rpe_actual;
        setLog.set_type = prevSet.set_type || 'working';
      }
    });

    exerciseLogs.set(exId, logs);
    setExerciseLogs(new Map(exerciseLogs));
  };

  const generateWarmups = async () => {
    const exId = currentExercise.id || `ex-${currentExercise.order}`;
    const logs = [...(exerciseLogs.get(exId) || [])];

    const firstWorkingSet = logs.find(s => s.set_type === 'working' && s.weight_kg > 0);
    if (!firstWorkingSet) {
      alert('Please enter a working weight first');
      return;
    }

    try {
      const response = await api.post('/exercises/warmups/generate', {
        working_weight_kg: firstWorkingSet.weight_kg,
        lift_type: currentExercise.lift_type || 'accessory',
      });

      const warmups = response.data.warmup_sets || [];

      const warmupLogs: SetLog[] = warmups.map((wu: any, idx: number) => ({
        set_number: idx + 1,
        reps_completed: wu.reps,
        weight_kg: wu.weight_kg,
        set_type: 'warm_up',
        completed: false,
        notes: wu.plate_setup,
      }));

      const updatedLogs = [
        ...warmupLogs,
        ...logs.map((log, idx) => ({ ...log, set_number: warmupLogs.length + idx + 1 })),
      ];

      exerciseLogs.set(exId, updatedLogs);
      setExerciseLogs(new Map(exerciseLogs));
      setShowWarmupGenerator(false);
    } catch (error) {
      console.error('Failed to generate warmups:', error);
      alert('Failed to generate warmup sets');
    }
  };

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

  const handleCompleteWorkout = async () => {
    if (!confirm('Are you sure you want to complete this workout?')) return;

    try {
      const workoutData = {
        session_id: session.id,
        exercises: exercises.map(ex => {
          const exId = ex.id || `ex-${ex.order}`;
          const sets = exerciseLogs.get(exId) || [];
          return {
            exercise_id: ex.id,
            sets: sets.map(s => ({
              set_number: s.set_number,
              reps_completed: s.reps_completed,
              weight_kg: s.weight_kg,
              rpe_actual: s.rpe_actual,
              notes: s.notes,
              set_type: s.set_type,
              media_urls: s.media_urls || [],
              exercise_notes: s.exercise_notes,
            })),
            notes: exerciseNotes,
          };
        }),
        notes: workoutNotes,
        rpe_rating: rpeRating || null,
      };

      await api.post('/programs/log-workout', workoutData);

      localStorage.removeItem(`workout-progress-${session.id}`);
      onComplete();
    } catch (error) {
      console.error('Failed to complete workout:', error);
      alert('Failed to save workout. Please try again.');
    }
  };

  const SET_TYPES = [
    { value: 'warm_up', label: 'Warm-up', color: 'bg-yellow-100 text-yellow-800' },
    { value: 'working', label: 'Working', color: 'bg-green-100 text-green-800' },
    { value: 'backoff', label: 'Backoff', color: 'bg-blue-100 text-blue-800' },
    { value: 'amrap', label: 'AMRAP', color: 'bg-purple-100 text-purple-800' },
    { value: 'failure', label: 'To Failure', color: 'bg-red-100 text-red-800' },
    { value: 'drop_set', label: 'Drop Set', color: 'bg-orange-100 text-orange-800' },
    { value: 'cluster', label: 'Cluster', color: 'bg-indigo-100 text-indigo-800' },
    { value: 'pause', label: 'Pause', color: 'bg-pink-100 text-pink-800' },
    { value: 'tempo', label: 'Tempo', color: 'bg-teal-100 text-teal-800' },
  ];

  if (!isStarted) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6">
          <h2 className="text-2xl font-bold mb-4 text-gray-900 dark:text-white">{session.session_name}</h2>
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            📅 {new Date(session.scheduled_date).toLocaleDateString()}
          </p>

          <div className="mb-6">
            <h3 className="font-semibold mb-3 text-gray-900 dark:text-white">Exercises ({session.exercises?.length || 0})</h3>
            <div className="space-y-2">
              {session.exercises?.map((ex: any, idx: number) => (
                <div key={idx} className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-700 rounded">
                  <div className="flex-1">
                    <div className="font-medium text-gray-900 dark:text-white">{ex.name}</div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">
                      {ex.sets}×{ex.reps} @ {ex.intensity || `RPE ${ex.rpe}`}
                      {ex.notes && <span className="ml-2 text-gray-500 dark:text-gray-500">• {ex.notes}</span>}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-3">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-900 dark:text-white"
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
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">{session.session_name}</h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Exercise {currentExerciseIdx + 1} of {exercises.length}
              </p>
            </div>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div className="mt-3 w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all"
              style={{
                width: `${((currentExerciseIdx + 1) / exercises.length) * 100}%`,
              }}
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4">
          {currentExercise && (
            <div>
              <div className="flex justify-between items-start mb-4">
                <div className="flex-1">
                  <h3 className="text-2xl font-bold text-gray-900 dark:text-white">{currentExercise.name}</h3>
                  <p className="text-gray-600 dark:text-gray-400 mt-1">
                    Target: {currentExercise.sets}×{currentExercise.reps} @ {currentExercise.intensity || `RPE ${currentExercise.rpe}`}
                  </p>
                  {currentExercise.notes && (
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">💡 {currentExercise.notes}</p>
                  )}
                </div>

                <div className="flex gap-2">
                  {previousSets.length > 0 && (
                    <button
                      onClick={autofillFromPrevious}
                      className="px-3 py-1 text-sm bg-purple-600 text-white rounded hover:bg-purple-700"
                    >
                      📋 Autofill Previous
                    </button>
                  )}
                  <button
                    onClick={() => setShowWarmupGenerator(true)}
                    className="px-3 py-1 text-sm bg-yellow-600 text-white rounded hover:bg-yellow-700"
                  >
                    🔥 Add Warm-ups
                  </button>
                </div>
              </div>

              <div className="mb-4">
                <label className="text-sm text-gray-600 dark:text-gray-400 block mb-1">Exercise Notes</label>
                <textarea
                  value={exerciseNotes}
                  onChange={(e) => setExerciseNotes(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded text-sm"
                  placeholder="How did this exercise feel? Any technique notes..."
                  rows={2}
                />
              </div>

              <div className="space-y-2">
                {currentLogs.map((setLog) => (
                  <div
                    key={setLog.set_number}
                    className={`border-2 rounded-lg p-4 ${
                      setLog.completed ? 'border-green-500 bg-green-50 dark:bg-green-900/20' : 'border-gray-200 dark:border-gray-700'
                    }`}
                  >
                    <div className="flex items-center gap-4 mb-2">
                      <div className="w-16 text-center">
                        <div className="font-semibold text-gray-700 dark:text-gray-300">Set {setLog.set_number}</div>
                      </div>

                      <div className="flex-1">
                        <label className="text-xs text-gray-600 dark:text-gray-400 block mb-1">Type</label>
                        <select
                          value={setLog.set_type}
                          onChange={(e) => updateSetLog(setLog.set_number, 'set_type', e.target.value)}
                          className="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded text-sm"
                        >
                          {SET_TYPES.map(type => (
                            <option key={type.value} value={type.value}>{type.label}</option>
                          ))}
                        </select>
                      </div>

                      <div className="flex-1">
                        <label className="text-xs text-gray-600 dark:text-gray-400 block mb-1">Weight (kg)</label>
                        <input
                          type="number"
                          step="0.5"
                          value={setLog.weight_kg || ''}
                          onChange={(e) => updateSetLog(setLog.set_number, 'weight_kg', parseFloat(e.target.value) || 0)}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded"
                          placeholder="0"
                        />
                      </div>

                      <div className="flex-1">
                        <label className="text-xs text-gray-600 dark:text-gray-400 block mb-1">Reps</label>
                        <input
                          type="number"
                          value={setLog.reps_completed || ''}
                          onChange={(e) => updateSetLog(setLog.set_number, 'reps_completed', parseInt(e.target.value) || 0)}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded"
                          placeholder="0"
                        />
                      </div>

                      <div className="flex-1">
                        <label className="text-xs text-gray-600 dark:text-gray-400 block mb-1">RPE</label>
                        <input
                          type="number"
                          step="0.5"
                          min="1"
                          max="10"
                          value={setLog.rpe_actual || ''}
                          onChange={(e) => updateSetLog(setLog.set_number, 'rpe_actual', parseFloat(e.target.value) || undefined)}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded"
                          placeholder="6.5"
                        />
                      </div>

                      <button
                        onClick={() => toggleSetCompleted(setLog.set_number)}
                        className={`px-4 py-2 rounded font-medium ${
                          setLog.completed
                            ? 'bg-green-600 text-white hover:bg-green-700'
                            : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                        }`}
                      >
                        {setLog.completed ? '✓ Done' : 'Mark Complete'}
                      </button>
                    </div>

                    <div className="mt-2">
                      <input
                        type="text"
                        value={setLog.notes || ''}
                        onChange={(e) => updateSetLog(setLog.set_number, 'notes', e.target.value)}
                        className="w-full px-3 py-1 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded text-sm"
                        placeholder="Set notes (e.g., felt heavy, good bar speed, etc.)"
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
          <div className="flex justify-between items-center">
            <button
              onClick={() => setCurrentExerciseIdx(Math.max(0, currentExerciseIdx - 1))}
              disabled={currentExerciseIdx === 0}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-white dark:hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed text-gray-900 dark:text-white"
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
              <label className="text-xs text-gray-600 dark:text-gray-400 block mb-1">Overall RPE (1-10)</label>
              <input
                type="number"
                step="0.5"
                min="1"
                max="10"
                value={rpeRating}
                onChange={(e) => setRpeRating(parseFloat(e.target.value) || '')}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded"
                placeholder="How hard was this workout?"
              />
            </div>
            <div>
              <label className="text-xs text-gray-600 dark:text-gray-400 block mb-1">Workout Notes</label>
              <input
                type="text"
                value={workoutNotes}
                onChange={(e) => setWorkoutNotes(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded"
                placeholder="How did you feel?"
              />
            </div>
          </div>
        </div>
      </div>

      {showWarmupGenerator && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[60]">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-bold mb-4 text-gray-900 dark:text-white">Generate Warm-up Sets</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
              Enter a working weight for the first working set, then click generate to create progressive warm-up sets.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setShowWarmupGenerator(false)}
                className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-900 dark:text-white"
              >
                Cancel
              </button>
              <button
                onClick={generateWarmups}
                className="px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700"
              >
                Generate Warm-ups
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
