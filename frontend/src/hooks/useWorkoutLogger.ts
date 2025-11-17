import { useState, useEffect } from 'react';
import { api } from '@/utils/apiWrapper';

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

export const useWorkoutLogger = (session: any, onComplete: () => void) => {
  const [isStarted, setIsStarted] = useState(!!session.completed_at);
  const [currentExerciseIdx, setCurrentExerciseIdx] = useState(0);
  const [exercises, setExercises] = useState<any[]>([]);
  const [exerciseLogs, setExerciseLogs] = useState<Map<string, SetLog[]>>(new Map());
  const [workoutNotes, setWorkoutNotes] = useState('');
  const [rpeRating, setRpeRating] = useState<number | ''>('');
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

  useEffect(() => {
    if (isStarted) {
      saveProgress();
    }
  }, [exerciseLogs, currentExerciseIdx, workoutNotes, rpeRating, isStarted]);

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

  return {
    isStarted,
    setIsStarted,
    currentExerciseIdx,
    setCurrentExerciseIdx,
    exercises,
    exerciseLogs,
    currentExercise,
    currentLogs,
    workoutNotes,
    setWorkoutNotes,
    rpeRating,
    setRpeRating,
    previousSets,
    exerciseNotes,
    setExerciseNotes,
    updateSetLog,
    toggleSetCompleted,
    autofillFromPrevious,
    generateWarmups,
    handleCompleteWorkout,
  };
};
