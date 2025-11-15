import React, { useState, useEffect } from 'react';
import { apiClient } from '@/utils/api';

interface WorkoutTemplate {
  id: string;
  name: string;
  description?: string;
  exercises: Array<{
    exercise_name: string;
    sets: number;
    reps_target: number;
    rest_seconds: number;
  }>;
  is_public: boolean;
  created_at: string;
}

interface WorkoutTemplateLibraryProps {
  onSelectTemplate?: (template: WorkoutTemplate) => void;
}

export const WorkoutTemplateLibrary: React.FC<WorkoutTemplateLibraryProps> = ({ onSelectTemplate }) => {
  const [templates, setTemplates] = useState<WorkoutTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<WorkoutTemplate | null>(null);

  useEffect(() => {
    loadTemplates();
  }, []);

  const loadTemplates = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getWorkoutTemplates();
      setTemplates(response.data.templates || []);
    } catch (err) {
      setError('Failed to load workout templates');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectTemplate = (template: WorkoutTemplate) => {
    if (onSelectTemplate) {
      onSelectTemplate(template);
    } else {
      setSelectedTemplate(template);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Workout Templates</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Save and reuse your favorite workout structures
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
        >
          + Create Template
        </button>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <p className="text-red-600 dark:text-red-400">{error}</p>
        </div>
      )}

      {templates.length === 0 ? (
        <div className="text-center py-16 bg-white dark:bg-gray-800 rounded-lg shadow">
          <svg className="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p className="text-lg text-gray-600 dark:text-gray-400 mb-2">No templates yet</p>
          <p className="text-sm text-gray-500 dark:text-gray-500 mb-4">
            Create your first workout template to save time on future sessions
          </p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
          >
            Create Template
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {templates.map((template) => (
            <div
              key={template.id}
              className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 hover:shadow-lg transition cursor-pointer"
              onClick={() => handleSelectTemplate(template)}
            >
              <div className="flex justify-between items-start mb-4">
                <h3 className="text-xl font-bold text-gray-900 dark:text-white">{template.name}</h3>
                {template.is_public && (
                  <span className="px-2 py-1 text-xs bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded">
                    Public
                  </span>
                )}
              </div>

              {template.description && (
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">{template.description}</p>
              )}

              <div className="space-y-2">
                <div className="text-sm text-gray-700 dark:text-gray-300">
                  <strong>{template.exercises?.length || 0}</strong> exercises
                </div>
                <div className="flex flex-wrap gap-2">
                  {template.exercises?.slice(0, 3).map((exercise, idx) => (
                    <span
                      key={idx}
                      className="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded"
                    >
                      {exercise.exercise_name}
                    </span>
                  ))}
                  {template.exercises?.length > 3 && (
                    <span className="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded">
                      +{template.exercises.length - 3} more
                    </span>
                  )}
                </div>
              </div>

              <button
                onClick={(e) => {
                  e.stopPropagation();
                  handleSelectTemplate(template);
                }}
                className="mt-4 w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
              >
                Use Template
              </button>
            </div>
          ))}
        </div>
      )}

      {showCreateModal && (
        <CreateTemplateModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={() => {
            setShowCreateModal(false);
            loadTemplates();
          }}
        />
      )}

      {selectedTemplate && !onSelectTemplate && (
        <TemplateDetailModal
          template={selectedTemplate}
          onClose={() => setSelectedTemplate(null)}
          onUse={() => {
            console.log('Using template:', selectedTemplate);
            setSelectedTemplate(null);
          }}
        />
      )}
    </div>
  );
};

interface CreateTemplateModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

const CreateTemplateModal: React.FC<CreateTemplateModalProps> = ({ onClose, onSuccess }) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!name.trim()) {
      setError('Template name is required');
      return;
    }

    try {
      setSaving(true);
      setError(null);

      await apiClient.createWorkoutTemplate({
        name: name.trim(),
        description: description.trim() || undefined,
        is_public: isPublic,
        exercises: [],
      });

      onSuccess();
    } catch (err) {
      setError('Failed to create template');
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full p-6">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">Create Workout Template</h2>

        {error && (
          <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
            <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          </div>
        )}

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Template Name *
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="e.g., Upper Body Day"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Optional description..."
            />
          </div>

          <div className="flex items-center">
            <input
              type="checkbox"
              id="is_public"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
            />
            <label htmlFor="is_public" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
              Make template public (visible to others)
            </label>
          </div>
        </div>

        <div className="mt-6 flex gap-3">
          <button
            onClick={onClose}
            disabled={saving}
            className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving || !name.trim()}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {saving ? 'Creating...' : 'Create'}
          </button>
        </div>
      </div>
    </div>
  );
};

interface TemplateDetailModalProps {
  template: WorkoutTemplate;
  onClose: () => void;
  onUse: () => void;
}

const TemplateDetailModal: React.FC<TemplateDetailModalProps> = ({ template, onClose, onUse }) => {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">{template.name}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4">
          {template.description && (
            <p className="text-gray-600 dark:text-gray-400 mb-6">{template.description}</p>
          )}

          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Exercises</h3>
          <div className="space-y-3">
            {template.exercises?.map((exercise, idx) => (
              <div key={idx} className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                <div className="font-semibold text-gray-900 dark:text-white mb-2">{exercise.exercise_name}</div>
                <div className="text-sm text-gray-600 dark:text-gray-400 grid grid-cols-3 gap-4">
                  <div>
                    <span className="font-medium">Sets:</span> {exercise.sets}
                  </div>
                  <div>
                    <span className="font-medium">Reps:</span> {exercise.reps_target}
                  </div>
                  <div>
                    <span className="font-medium">Rest:</span> {exercise.rest_seconds}s
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition"
          >
            Close
          </button>
          <button
            onClick={onUse}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
          >
            Use This Template
          </button>
        </div>
      </div>
    </div>
  );
};
