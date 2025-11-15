import React, { useState, useEffect } from 'react';
import { apiClient } from '@/utils/api';

interface ProgramChange {
  id: string;
  program_id: string;
  change_type: string;
  changes_json: any;
  description?: string;
  status: 'pending' | 'approved' | 'rejected';
  proposed_by: string;
  proposed_at: string;
  approved_by?: string;
  resolved_at?: string;
}

interface ProgramChangeManagerProps {
  programId: string;
}

export const ProgramChangeManager: React.FC<ProgramChangeManagerProps> = ({ programId }) => {
  const [changes, setChanges] = useState<ProgramChange[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedChange, setSelectedChange] = useState<ProgramChange | null>(null);

  useEffect(() => {
    if (programId) {
      loadPendingChanges();
    }
  }, [programId]);

  const loadPendingChanges = async () => {
    try {
      setLoading(true);
      const response = await apiClient.get(`/programs/${programId}/changes/pending`);
      setChanges(response.data.changes || []);
    } catch (err) {
      setError('Failed to load pending changes');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (changeId: string) => {
    try {
      await apiClient.post(`/programs/changes/${changeId}/apply`);
      loadPendingChanges();
    } catch (err) {
      console.error('Failed to approve change:', err);
      alert('Failed to approve change');
    }
  };

  const handleReject = async (changeId: string) => {
    try {
      await apiClient.post(`/programs/changes/${changeId}/reject`);
      loadPendingChanges();
    } catch (err) {
      console.error('Failed to reject change:', err);
      alert('Failed to reject change');
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
        <p className="text-red-600 dark:text-red-400">{error}</p>
      </div>
    );
  }

  if (changes.length === 0) {
    return (
      <div className="text-center py-12 bg-white dark:bg-gray-800 rounded-lg shadow">
        <svg className="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p className="text-lg text-gray-600 dark:text-gray-400">No pending changes</p>
        <p className="text-sm text-gray-500 dark:text-gray-500 mt-2">
          All program changes have been reviewed
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Pending Program Changes</h2>
        <span className="px-3 py-1 bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200 rounded-full text-sm font-medium">
          {changes.length} pending
        </span>
      </div>

      <div className="space-y-4">
        {changes.map((change) => (
          <div
            key={change.id}
            className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border-l-4 border-yellow-500"
          >
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                  {change.change_type.replace(/_/g, ' ').toUpperCase()}
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  Proposed {new Date(change.proposed_at).toLocaleDateString()}
                </p>
              </div>
              <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                change.status === 'pending' ? 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200' :
                change.status === 'approved' ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200' :
                'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200'
              }`}>
                {change.status}
              </span>
            </div>

            {change.description && (
              <p className="text-gray-700 dark:text-gray-300 mb-4">{change.description}</p>
            )}

            <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 mb-4">
              <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Changes:</h4>
              <pre className="text-xs text-gray-600 dark:text-gray-400 overflow-x-auto">
                {JSON.stringify(change.changes_json, null, 2)}
              </pre>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => setSelectedChange(change)}
                className="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition"
              >
                View Details
              </button>
              <button
                onClick={() => handleApprove(change.id)}
                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition"
              >
                ✓ Approve
              </button>
              <button
                onClick={() => handleReject(change.id)}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition"
              >
                ✗ Reject
              </button>
            </div>
          </div>
        ))}
      </div>

      {selectedChange && (
        <ChangeDetailModal
          change={selectedChange}
          onClose={() => setSelectedChange(null)}
          onApprove={() => {
            handleApprove(selectedChange.id);
            setSelectedChange(null);
          }}
          onReject={() => {
            handleReject(selectedChange.id);
            setSelectedChange(null);
          }}
        />
      )}
    </div>
  );
};

interface ChangeDetailModalProps {
  change: ProgramChange;
  onClose: () => void;
  onApprove: () => void;
  onReject: () => void;
}

const ChangeDetailModal: React.FC<ChangeDetailModalProps> = ({ change, onClose, onApprove, onReject }) => {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
              {change.change_type.replace(/_/g, ' ').toUpperCase()}
            </h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              Proposed {new Date(change.proposed_at).toLocaleString()}
            </p>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4">
          {change.description && (
            <div className="mb-6">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Description:</h3>
              <p className="text-gray-900 dark:text-white">{change.description}</p>
            </div>
          )}

          <div className="mb-6">
            <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Proposed Changes:</h3>
            <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
              <pre className="text-sm text-gray-900 dark:text-white overflow-x-auto whitespace-pre-wrap">
                {JSON.stringify(change.changes_json, null, 2)}
              </pre>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-gray-600 dark:text-gray-400">Change Type:</span>
              <p className="text-gray-900 dark:text-white font-medium">{change.change_type}</p>
            </div>
            <div>
              <span className="text-gray-600 dark:text-gray-400">Status:</span>
              <p className={`font-medium ${
                change.status === 'pending' ? 'text-yellow-600 dark:text-yellow-400' :
                change.status === 'approved' ? 'text-green-600 dark:text-green-400' :
                'text-red-600 dark:text-red-400'
              }`}>
                {change.status.toUpperCase()}
              </p>
            </div>
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
            onClick={onReject}
            className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition"
          >
            ✗ Reject
          </button>
          <button
            onClick={onApprove}
            className="flex-1 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition"
          >
            ✓ Approve
          </button>
        </div>
      </div>
    </div>
  );
};
