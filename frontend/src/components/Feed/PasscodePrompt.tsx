import React, { useState } from 'react';
import { useAuthStore } from '@/store/authStore';
import { offlineQueue } from '@/utils/offlineQueue';

import { generateUUID } from '@/utils/uuid';
interface PasscodePromptProps {
  targetUserId: string;
  onAccessGranted: (token: string) => void;
  onCancel: () => void;
}

const PasscodePrompt: React.FC<PasscodePromptProps> = ({
  targetUserId,
  onAccessGranted,
  onCancel,
}) => {
  const { user } = useAuthStore();

  const [passcode, setPasscode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!passcode || !user) {
      setError('Please enter a passcode');
      return;
    }

    setLoading(true);

    const feedAccessAttemptEvent = {
      schema_version: '1.0.0',
      event_type: 'feed.access.attempt',
      client_generated_id: generateUUID(),
      user_id: user.id,
      timestamp: new Date().toISOString(),
      source_service: 'frontend',
      data: {
        target_user_id: targetUserId,
        passcode_provided: passcode,
      },
    };

    try {
      await offlineQueue.enqueue(feedAccessAttemptEvent);

      const mockToken = `access_token_${Date.now()}`;
      localStorage.setItem(`feed_access_${targetUserId}`, mockToken);

      setTimeout(() => {
        setLoading(false);
        onAccessGranted(mockToken);
      }, 500);
    } catch (error) {
      console.error('Access attempt failed:', error);
      setError('Failed to verify passcode. Please try again.');
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold">Enter Passcode</h2>
          <button
            onClick={onCancel}
            className="text-gray-400 hover:text-gray-600"
            disabled={loading}
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <p className="text-gray-600 mb-4">
          This user's feed is protected. Please enter the passcode to view their content.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Passcode</label>
            <input
              type="password"
              value={passcode}
              onChange={(e) => setPasscode(e.target.value)}
              className="w-full border rounded p-2"
              placeholder="Enter passcode"
              autoFocus
              disabled={loading}
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="flex gap-3">
            <button
              type="button"
              onClick={onCancel}
              className="flex-1 bg-gray-200 text-gray-700 py-2 px-4 rounded hover:bg-gray-300 disabled:opacity-50"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed"
              disabled={!passcode || loading}
            >
              {loading ? 'Verifying...' : 'Submit'}
            </button>
          </div>
        </form>

        <div className="mt-4 text-sm text-gray-500">
          <p>Storage option:</p>
          <div className="flex gap-4 mt-2">
            <label className="flex items-center">
              <input type="radio" name="storage" value="cookie" className="mr-2" defaultChecked />
              Cookie (secure)
            </label>
            <label className="flex items-center">
              <input type="radio" name="storage" value="localStorage" className="mr-2" />
              Local Storage
            </label>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PasscodePrompt;
