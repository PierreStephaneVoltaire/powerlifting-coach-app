import React from 'react';
import { getDifficultyColor, getLiftTypeColor, getMuscleGroupColor } from '@/utils/exerciseUtils';

interface Exercise {
  id: string;
  name: string;
  description?: string;
  lift_type: string;
  primary_muscles: string[];
  difficulty?: string;
  equipment_needed: string[];
  is_custom: boolean;
}

interface ExerciseCardProps {
  exercise: Exercise;
  onClick: () => void;
}

export const ExerciseCard: React.FC<ExerciseCardProps> = ({ exercise, onClick }) => {
  return (
    <div
      onClick={onClick}
      className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer"
    >
      <div className="flex justify-between items-start mb-2">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{exercise.name}</h3>
        {exercise.is_custom && (
          <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300 rounded">
            Custom
          </span>
        )}
      </div>

      {exercise.description && (
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-3 line-clamp-2">
          {exercise.description}
        </p>
      )}

      <div className="flex flex-wrap gap-2 mb-3">
        <span className={`px-2 py-1 text-xs rounded ${getLiftTypeColor(exercise.lift_type)}`}>
          {exercise.lift_type}
        </span>
        {exercise.difficulty && (
          <span className={`px-2 py-1 text-xs rounded ${getDifficultyColor(exercise.difficulty)}`}>
            {exercise.difficulty}
          </span>
        )}
      </div>

      <div className="flex flex-wrap gap-1 mb-2">
        {exercise.primary_muscles?.slice(0, 3).map((muscle, idx) => (
          <span
            key={idx}
            className={`px-2 py-1 text-xs rounded ${getMuscleGroupColor(muscle)}`}
          >
            {muscle}
          </span>
        ))}
        {exercise.primary_muscles?.length > 3 && (
          <span className="px-2 py-1 text-xs text-gray-600 dark:text-gray-400">
            +{exercise.primary_muscles.length - 3} more
          </span>
        )}
      </div>

      {exercise.equipment_needed?.length > 0 && (
        <div className="text-xs text-gray-500 dark:text-gray-400">
          Equipment: {exercise.equipment_needed.join(', ')}
        </div>
      )}
    </div>
  );
};
