import React, { useState, useEffect } from 'react';
import { apiClient } from '../../utils/api';
import { ExerciseDetail } from './ExerciseDetail';

interface Exercise {
  id: string;
  name: string;
  description?: string;
  lift_type: string;
  primary_muscles: string[];
  secondary_muscles: string[];
  difficulty?: string;
  equipment_needed: string[];
  demo_video_url?: string;
  instructions?: string;
  form_cues: string[];
  is_custom: boolean;
  is_public: boolean;
}

export const ExerciseLibrary: React.FC = () => {
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [filteredExercises, setFilteredExercises] = useState<Exercise[]>([]);
  const [selectedLiftType, setSelectedLiftType] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedExercise, setSelectedExercise] = useState<Exercise | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchExercises();
  }, [selectedLiftType]);

  useEffect(() => {
    filterExercises();
  }, [exercises, searchQuery, selectedLiftType]);

  const fetchExercises = async () => {
    setLoading(true);
    try {
      const liftType = selectedLiftType === 'all' ? undefined : selectedLiftType;
      const response = await apiClient.getExerciseLibrary(liftType);
      setExercises(response.exercises || []);
    } catch (error) {
      console.error('Failed to fetch exercises:', error);
    } finally {
      setLoading(false);
    }
  };

  const filterExercises = () => {
    let filtered = exercises;

    if (searchQuery) {
      filtered = filtered.filter(ex =>
        ex.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        ex.description?.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    setFilteredExercises(filtered);
  };

  const getDifficultyColor = (difficulty?: string) => {
    switch (difficulty) {
      case 'beginner':
        return 'bg-green-100 text-green-800';
      case 'intermediate':
        return 'bg-yellow-100 text-yellow-800';
      case 'advanced':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getLiftTypeColor = (liftType: string) => {
    switch (liftType) {
      case 'squat':
        return 'bg-red-100 text-red-800';
      case 'bench':
        return 'bg-blue-100 text-blue-800';
      case 'deadlift':
        return 'bg-green-100 text-green-800';
      case 'accessory':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Exercise Library</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Browse {filteredExercises.length} exercises with form cues and demo videos
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          + Add Custom Exercise
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Search</label>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search exercises..."
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Lift Type</label>
            <select
              value={selectedLiftType}
              onChange={(e) => setSelectedLiftType(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
            >
              <option value="all">All Types</option>
              <option value="squat">Squat</option>
              <option value="bench">Bench Press</option>
              <option value="deadlift">Deadlift</option>
              <option value="accessory">Accessory</option>
            </select>
          </div>
        </div>
      </div>

      {/* Exercise Grid */}
      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredExercises.map((exercise) => (
            <div
              key={exercise.id}
              onClick={() => setSelectedExercise(exercise)}
              className="bg-white dark:bg-gray-800 rounded-lg shadow hover:shadow-lg transition cursor-pointer overflow-hidden"
            >
              {/* Exercise Header */}
              <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                <div className="flex justify-between items-start mb-2">
                  <h3 className="font-bold text-lg text-gray-900 dark:text-white">{exercise.name}</h3>
                  {exercise.is_custom && (
                    <span className="px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 text-xs rounded">
                      Custom
                    </span>
                  )}
                </div>

                <div className="flex gap-2 flex-wrap">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${getLiftTypeColor(exercise.lift_type)}`}>
                    {exercise.lift_type}
                  </span>
                  {exercise.difficulty && (
                    <span className={`px-2 py-1 rounded text-xs font-medium ${getDifficultyColor(exercise.difficulty)}`}>
                      {exercise.difficulty}
                    </span>
                  )}
                </div>
              </div>

              {/* Exercise Info */}
              <div className="p-4">
                {exercise.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-3 line-clamp-2">
                    {exercise.description}
                  </p>
                )}

                {/* Primary Muscles */}
                {exercise.primary_muscles && exercise.primary_muscles.length > 0 && (
                  <div className="mb-3">
                    <div className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Primary Muscles</div>
                    <div className="flex gap-1 flex-wrap">
                      {exercise.primary_muscles.map((muscle, idx) => (
                        <span key={idx} className="px-2 py-0.5 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-xs rounded">
                          {muscle}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {/* Equipment */}
                {exercise.equipment_needed && exercise.equipment_needed.length > 0 && (
                  <div className="mb-3">
                    <div className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Equipment</div>
                    <div className="flex gap-1 flex-wrap">
                      {exercise.equipment_needed.slice(0, 3).map((eq, idx) => (
                        <span key={idx} className="px-2 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 text-xs rounded">
                          {eq.replace(/_/g, ' ')}
                        </span>
                      ))}
                      {exercise.equipment_needed.length > 3 && (
                        <span className="px-2 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 text-xs rounded">
                          +{exercise.equipment_needed.length - 3} more
                        </span>
                      )}
                    </div>
                  </div>
                )}

                {/* Form Cues Preview */}
                {exercise.form_cues && exercise.form_cues.length > 0 && (
                  <div>
                    <div className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Key Cue</div>
                    <div className="text-sm text-gray-600 dark:text-gray-400 italic">
                      "{exercise.form_cues[0]}"
                    </div>
                  </div>
                )}

                {exercise.demo_video_url && (
                  <div className="mt-3 text-sm text-blue-600 dark:text-blue-400 font-medium">
                    📹 Video available
                  </div>
                )}
              </div>

              {/* Footer */}
              <div className="px-4 py-3 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700">
                <button className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium">
                  View Details →
                </button>
              </div>
            </div>
          ))}

          {filteredExercises.length === 0 && (
            <div className="col-span-full text-center py-16 text-gray-500 dark:text-gray-400">
              <p className="text-lg">No exercises found</p>
              <p className="text-sm mt-2">Try adjusting your filters or create a custom exercise</p>
            </div>
          )}
        </div>
      )}

      {/* Exercise Detail Modal */}
      {selectedExercise && (
        <ExerciseDetail
          exercise={selectedExercise}
          onClose={() => setSelectedExercise(null)}
        />
      )}

      {/* Create Exercise Modal */}
      {showCreateModal && (
        <CreateExerciseModal
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            fetchExercises();
          }}
        />
      )}
    </div>
  );
};

// Create Exercise Modal Component
const CreateExerciseModal: React.FC<{ onClose: () => void; onCreated: () => void }> = ({
  onClose,
  onCreated,
}) => {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    lift_type: 'accessory',
    difficulty: 'intermediate',
    primary_muscles: '',
    secondary_muscles: '',
    equipment_needed: '',
    demo_video_url: '',
    instructions: '',
    form_cues: '',
  });

  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const exerciseData = {
        name: formData.name,
        description: formData.description || null,
        lift_type: formData.lift_type,
        difficulty: formData.difficulty || null,
        primary_muscles: formData.primary_muscles.split(',').map(m => m.trim()).filter(Boolean),
        secondary_muscles: formData.secondary_muscles.split(',').map(m => m.trim()).filter(Boolean),
        equipment_needed: formData.equipment_needed.split(',').map(e => e.trim()).filter(Boolean),
        demo_video_url: formData.demo_video_url || null,
        instructions: formData.instructions || null,
        form_cues: formData.form_cues.split('\n').map(c => c.trim()).filter(Boolean),
      };

      await apiClient.createCustomExercise(exerciseData);
      onCreated();
    } catch (error) {
      console.error('Failed to create exercise:', error);
      alert('Failed to create exercise. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Create Custom Exercise</h2>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Exercise Name *</label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
              placeholder="e.g., Bulgarian Split Squat"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
              rows={2}
              placeholder="Brief description of the exercise..."
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Lift Type *</label>
              <select
                required
                value={formData.lift_type}
                onChange={(e) => setFormData({ ...formData, lift_type: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
              >
                <option value="squat">Squat</option>
                <option value="bench">Bench Press</option>
                <option value="deadlift">Deadlift</option>
                <option value="accessory">Accessory</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Difficulty</label>
              <select
                value={formData.difficulty}
                onChange={(e) => setFormData({ ...formData, difficulty: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
              >
                <option value="beginner">Beginner</option>
                <option value="intermediate">Intermediate</option>
                <option value="advanced">Advanced</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Primary Muscles (comma-separated)
            </label>
            <input
              type="text"
              value={formData.primary_muscles}
              onChange={(e) => setFormData({ ...formData, primary_muscles: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              placeholder="e.g., quadriceps, glutes"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Secondary Muscles (comma-separated)
            </label>
            <input
              type="text"
              value={formData.secondary_muscles}
              onChange={(e) => setFormData({ ...formData, secondary_muscles: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              placeholder="e.g., hamstrings, core"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Equipment Needed (comma-separated)
            </label>
            <input
              type="text"
              value={formData.equipment_needed}
              onChange={(e) => setFormData({ ...formData, equipment_needed: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              placeholder="e.g., barbell, bench, dumbbells"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Demo Video URL</label>
            <input
              type="url"
              value={formData.demo_video_url}
              onChange={(e) => setFormData({ ...formData, demo_video_url: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              placeholder="https://youtube.com/..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Instructions</label>
            <textarea
              value={formData.instructions}
              onChange={(e) => setFormData({ ...formData, instructions: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              rows={3}
              placeholder="Step-by-step instructions..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Form Cues (one per line)
            </label>
            <textarea
              value={formData.form_cues}
              onChange={(e) => setFormData({ ...formData, form_cues: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              rows={4}
              placeholder="Brace core hard&#10;Keep chest up&#10;Drive through heels"
            />
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Create Exercise'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
