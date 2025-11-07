import React, { useState } from 'react';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';

import { generateUUID } from '@/utils/uuid';
interface MachineNote {
  note_id: string;
  brand: string;
  model: string;
  machine_type: 'barbell' | 'hack_squat' | 'leg_press' | 'hex_bar' | 'cable' | 'other';
  settings: string;
  visibility: 'public' | 'private';
  created_at: string;
}

export const MachineNotes: React.FC = () => {
  const { user } = useAuthStore();
  const [notes, setNotes] = useState<MachineNote[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    brand: '',
    model: '',
    machine_type: 'barbell' as MachineNote['machine_type'],
    settings: '',
    visibility: 'private' as MachineNote['visibility'],
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!user) {
      setError('User not authenticated');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const noteId = generateUUID();
      const event = {
        schema_version: '1.0.0',
        event_type: 'machine.notes.submitted',
        client_generated_id: noteId,
        user_id: user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          note_id: noteId,
          brand: formData.brand,
          model: formData.model,
          machine_type: formData.machine_type,
          settings: formData.settings,
          visibility: formData.visibility,
        },
      };

      await apiClient.submitEvent(event);
      console.info('Machine note submitted', { note_id: noteId });

      const newNote: MachineNote = {
        note_id: noteId,
        ...formData,
        created_at: new Date().toISOString(),
      };
      setNotes([newNote, ...notes]);

      setFormData({
        brand: '',
        model: '',
        machine_type: 'barbell',
        settings: '',
        visibility: 'private',
      });
      setShowForm(false);
    } catch (err: any) {
      console.error('Failed to submit machine note', err);
      if (!err.queued) {
        setError(err.response?.data?.error || 'Failed to save note. Please try again.');
      } else {
        setShowForm(false);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg">
        <div className="p-6 border-b border-gray-200 flex items-center justify-between">
          <h2 className="text-2xl font-bold text-gray-900">Machine Notes</h2>
          <button
            onClick={() => setShowForm(!showForm)}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            {showForm ? 'Cancel' : 'Add Note'}
          </button>
        </div>

        {showForm && (
          <div className="p-6 bg-gray-50 border-b border-gray-200">
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Brand *
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.brand}
                    onChange={(e) => setFormData({ ...formData, brand: e.target.value })}
                    placeholder="e.g., Rogue, Eleiko"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Model
                  </label>
                  <input
                    type="text"
                    value={formData.model}
                    onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                    placeholder="e.g., Ohio Power Bar"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Machine Type *
                </label>
                <select
                  required
                  value={formData.machine_type}
                  onChange={(e) =>
                    setFormData({ ...formData, machine_type: e.target.value as MachineNote['machine_type'] })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="barbell">Barbell</option>
                  <option value="hack_squat">Hack Squat</option>
                  <option value="leg_press">Leg Press</option>
                  <option value="hex_bar">Hex Bar</option>
                  <option value="cable">Cable Machine</option>
                  <option value="other">Other</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Settings & Notes *
                </label>
                <textarea
                  required
                  rows={4}
                  value={formData.settings}
                  onChange={(e) => setFormData({ ...formData, settings: e.target.value })}
                  placeholder="e.g., Seat position 5, footplate angle 45°, belt hooks on 3rd hole..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Visibility
                </label>
                <select
                  value={formData.visibility}
                  onChange={(e) =>
                    setFormData({ ...formData, visibility: e.target.value as MachineNote['visibility'] })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="private">Private (Only Me)</option>
                  <option value="public">Public (Share with Community)</option>
                </select>
              </div>

              {error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                  <p className="text-sm text-red-600">{error}</p>
                </div>
              )}

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {isSubmitting ? 'Saving...' : 'Save Note'}
              </button>
            </form>
          </div>
        )}

        <div className="divide-y divide-gray-200">
          {notes.length === 0 ? (
            <div className="p-12 text-center">
              <svg
                className="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
              <h3 className="mt-2 text-lg font-medium text-gray-900">No machine notes</h3>
              <p className="mt-1 text-sm text-gray-500">
                Save your equipment settings and preferences
              </p>
            </div>
          ) : (
            notes.map((note) => (
              <div key={note.note_id} className="p-6 hover:bg-gray-50">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">
                      {note.brand} {note.model && `- ${note.model}`}
                    </h3>
                    <p className="text-sm text-gray-600 mt-1">
                      <span className="capitalize">{note.machine_type.replace('_', ' ')}</span>
                      {' • '}
                      <span className="capitalize">{note.visibility}</span>
                    </p>
                  </div>
                  <span className="text-xs text-gray-500">
                    {new Date(note.created_at).toLocaleDateString()}
                  </span>
                </div>
                <p className="text-gray-700 whitespace-pre-wrap">{note.settings}</p>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};
