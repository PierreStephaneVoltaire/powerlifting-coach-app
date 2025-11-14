export const getDifficultyColor = (difficulty?: string) => {
  switch (difficulty) {
    case 'beginner':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300';
    case 'intermediate':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300';
    case 'advanced':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
  }
};

export const getLiftTypeColor = (liftType: string) => {
  switch (liftType) {
    case 'squat':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300';
    case 'bench':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300';
    case 'deadlift':
      return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300';
    case 'accessory':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
  }
};

export const getMuscleGroupColor = (muscle: string) => {
  const muscleColors: Record<string, string> = {
    'quadriceps': 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
    'hamstrings': 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
    'glutes': 'bg-pink-100 text-pink-800 dark:bg-pink-900/30 dark:text-pink-300',
    'chest': 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
    'back': 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
    'shoulders': 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300',
    'triceps': 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
    'biceps': 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-300',
    'core': 'bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-300',
  };

  return muscleColors[muscle.toLowerCase()] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
};
