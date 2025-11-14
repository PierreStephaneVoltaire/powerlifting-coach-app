import { useState, useEffect } from 'react';

interface Exercise {
  id: string;
  name: string;
  description?: string;
  lift_type: string;
  [key: string]: any;
}

export const useExerciseFilters = (exercises: Exercise[]) => {
  const [filteredExercises, setFilteredExercises] = useState<Exercise[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedLiftType, setSelectedLiftType] = useState<string>('all');

  useEffect(() => {
    filterExercises();
  }, [exercises, searchQuery, selectedLiftType]);

  const filterExercises = () => {
    let filtered = exercises;

    if (selectedLiftType !== 'all') {
      filtered = filtered.filter(ex => ex.lift_type === selectedLiftType);
    }

    if (searchQuery) {
      filtered = filtered.filter(ex =>
        ex.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        ex.description?.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    setFilteredExercises(filtered);
  };

  return {
    filteredExercises,
    searchQuery,
    setSearchQuery,
    selectedLiftType,
    setSelectedLiftType,
  };
};
