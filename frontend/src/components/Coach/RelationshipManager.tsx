import React, { useEffect, useState } from 'react';

interface Relationship {
  id: string;
  coach_id: string;
  athlete_id: string;
  status: 'pending' | 'active' | 'terminated';
  request_message?: string;
  requested_at: string;
  accepted_at?: string;
  terminated_at?: string;
  termination_reason?: string;
}

interface User {
  id: string;
  name: string;
  email: string;
}

export const RelationshipManager: React.FC = () => {
  const [relationships, setRelationships] = useState<Relationship[]>([]);
  const [users, setUsers] = useState<Map<string, User>>(new Map());
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'pending' | 'active' | 'all'>('pending');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadRelationships();
  }, []);

  const loadRelationships = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/relationships', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load relationships');
      }

      const data = await response.json();
      setRelationships(data.relationships || []);

      const userIds = new Set<string>();
      data.relationships?.forEach((rel: Relationship) => {
        userIds.add(rel.coach_id);
        userIds.add(rel.athlete_id);
      });

      await loadUsers(Array.from(userIds));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load relationships');
    } finally {
      setLoading(false);
    }
  };

  const loadUsers = async (userIds: string[]) => {
    const userMap = new Map<string, User>();

    for (const userId of userIds) {
      try {
        const response = await fetch(`/api/v1/users/${userId}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          },
        });

        if (response.ok) {
          const user = await response.json();
          userMap.set(userId, user);
        }
      } catch (err) {
        console.error(`Failed to load user ${userId}:`, err);
      }
    }

    setUsers(userMap);
  };

  const acceptRequest = async (relationshipId: string) => {
    try {
      const response = await fetch(`/api/v1/relationships/${relationshipId}/accept`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to accept request');
      }

      await loadRelationships();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to accept request');
    }
  };

  const terminateRelationship = async (relationshipId: string, reason?: string) => {
    if (!window.confirm('Are you sure you want to terminate this coaching relationship?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/relationships/${relationshipId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          termination_reason: reason,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to terminate relationship');
      }

      await loadRelationships();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to terminate relationship');
    }
  };

  const filteredRelationships = relationships.filter(rel => {
    if (activeTab === 'all') return true;
    return rel.status === activeTab;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600 dark:text-gray-400">Loading relationships...</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-6">
        Coaching Relationships
      </h1>

      {/* Tabs */}
      <div className="flex gap-2 mb-6 border-b border-gray-200 dark:border-gray-700">
        {[
          { key: 'pending', label: 'Pending Requests' },
          { key: 'active', label: 'Active Coaching' },
          { key: 'all', label: 'All' },
        ].map(tab => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key as any)}
            className={`px-4 py-2 font-medium transition-colors border-b-2 ${
              activeTab === tab.key
                ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
            }`}
          >
            {tab.label}
            {tab.key !== 'all' && (
              <span className="ml-2 px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded-full text-xs">
                {relationships.filter(r => r.status === tab.key).length}
              </span>
            )}
          </button>
        ))}
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <p className="text-red-600 dark:text-red-400">{error}</p>
        </div>
      )}

      {/* Relationships List */}
      {filteredRelationships.length === 0 ? (
        <div className="text-center py-12 bg-white dark:bg-gray-800 rounded-lg">
          <p className="text-gray-600 dark:text-gray-400">
            {activeTab === 'pending'
              ? 'No pending requests'
              : activeTab === 'active'
              ? 'No active coaching relationships'
              : 'No relationships yet'}
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredRelationships.map((rel) => {
            const otherUser = users.get(
              localStorage.getItem('userType') === 'coach' ? rel.athlete_id : rel.coach_id
            );

            return (
              <div
                key={rel.id}
                className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                        {otherUser?.name || 'Loading...'}
                      </h3>
                      <span
                        className={`px-2 py-1 rounded text-xs font-medium ${
                          rel.status === 'pending'
                            ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-300'
                            : rel.status === 'active'
                            ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300'
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                        }`}
                      >
                        {rel.status.charAt(0).toUpperCase() + rel.status.slice(1)}
                      </span>
                    </div>

                    {rel.request_message && (
                      <p className="text-gray-600 dark:text-gray-300 text-sm mb-3 italic">
                        "{rel.request_message}"
                      </p>
                    )}

                    <div className="flex flex-wrap gap-4 text-sm text-gray-500 dark:text-gray-400">
                      <span>Requested: {new Date(rel.requested_at).toLocaleDateString()}</span>
                      {rel.accepted_at && (
                        <span>Accepted: {new Date(rel.accepted_at).toLocaleDateString()}</span>
                      )}
                      {rel.terminated_at && (
                        <span>Ended: {new Date(rel.terminated_at).toLocaleDateString()}</span>
                      )}
                    </div>

                    {rel.termination_reason && (
                      <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                        Reason: {rel.termination_reason}
                      </p>
                    )}
                  </div>

                  <div className="flex gap-2">
                    {rel.status === 'pending' && localStorage.getItem('userType') === 'coach' && (
                      <>
                        <button
                          onClick={() => acceptRequest(rel.id)}
                          className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg
                                   font-medium transition-colors text-sm"
                        >
                          Accept
                        </button>
                        <button
                          onClick={() => terminateRelationship(rel.id, 'Request declined')}
                          className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg
                                   font-medium transition-colors text-sm"
                        >
                          Decline
                        </button>
                      </>
                    )}
                    {rel.status === 'active' && (
                      <button
                        onClick={() => terminateRelationship(rel.id)}
                        className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg
                                 font-medium transition-colors text-sm"
                      >
                        End Coaching
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
