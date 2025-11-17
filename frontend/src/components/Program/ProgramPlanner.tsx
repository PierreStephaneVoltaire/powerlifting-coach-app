import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';

import { generateUUID } from '@/utils/uuid';
interface ProgramFormData {
  name: string;
  start_date: string;
  comp_date: string;
  training_days_per_week: number;
  notes?: string;
}

export const ProgramPlanner: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [formData, setFormData] = useState<ProgramFormData>({
    name: '',
    start_date: new Date().toISOString().split('T')[0],
    comp_date: '',
    training_days_per_week: 4,
    notes: '',
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const calculateWeeksUntilComp = () => {
    if (!formData.comp_date) return 0;
    const today = new Date();
    const compDate = new Date(formData.comp_date);
    const diffTime = compDate.getTime() - today.getTime();
    const diffWeeks = Math.ceil(diffTime / (1000 * 60 * 60 * 24 * 7));
    return diffWeeks;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!user) {
      setError('User not authenticated');
      return;
    }

    if (!formData.comp_date) {
      setError('Competition date is required');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const event = {
        schema_version: '1.0.0',
        event_type: 'program.plan.created',
        client_generated_id: generateUUID(),
        user_id: user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          name: formData.name,
          start_date: formData.start_date,
          comp_date: formData.comp_date,
          training_days_per_week: formData.training_days_per_week,
          notes: formData.notes,
        },
      };

      await apiClient.submitEvent(event);
      console.info('Program plan created', { event_type: event.event_type });
      navigate('/program/list');
    } catch (err: any) {
      console.error('Failed to create program plan', err);
      if (err.queued) {
        navigate('/program/list');
      } else {
        setError(err.response?.data?.error || 'Failed to create program. Please try again.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const weeksUntilComp = calculateWeeksUntilComp();

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg p-8">
        <h2 className="text-2xl font-bold mb-6">Create Training Plan</h2>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Plan Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g., 12-Week Comp Prep"
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Start Date *
              </label>
              <input
                type="date"
                required
                value={formData.start_date}
                onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Competition Date *
              </label>
              <input
                type="date"
                required
                value={formData.comp_date}
                onChange={(e) => setFormData({ ...formData, comp_date: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
              />
            </div>
          </div>

          {weeksUntilComp > 0 && (
            <div className="p-3 bg-blue-50 border border-blue-200 rounded-md">
              <p className="text-sm text-blue-800">
                <span className="font-semibold">{weeksUntilComp} weeks</span> until competition
              </p>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Training Days Per Week *
            </label>
            <input
              type="number"
              required
              min="1"
              max="7"
              value={formData.training_days_per_week}
              onChange={(e) => setFormData({ ...formData, training_days_per_week: parseInt(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Notes
            </label>
            <textarea
              rows={4}
              value={formData.notes}
              onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              placeholder="Add any notes about this training plan..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            />
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={() => navigate('/program/list')}
              className="px-6 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {isSubmitting ? 'Creating...' : 'Create Plan'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
