import React from 'react';

interface WorkoutPreviewProps {
  session: any;
  onClose: () => void;
  onStart: () => void;
}

export const WorkoutPreview: React.FC<WorkoutPreviewProps> = ({ session, onClose, onStart }) => {
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
            onClick={onStart}
            className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Start Workout
          </button>
        </div>
      </div>
    </div>
  );
};
