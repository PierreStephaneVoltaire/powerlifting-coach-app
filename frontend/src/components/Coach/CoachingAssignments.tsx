import React, { useEffect, useState } from 'react';

interface CoachingAssignment {
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

export const CoachingAssignments: React.FC = () => {
  const [assignments, setAssignments] = useState<CoachingAssignment[]>([]);
  const [users, setUsers] = useState<Map<string, User>>(new Map());
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'pending' | 'active' | 'all'>('pending');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAssignments();
  }, []);

  const loadAssignments = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/coaches/assignments', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load coaching assignments');
      }

      const data = await response.json();
      setAssignments(data.assignments || []);

      const userIds = new Set<string>();
      data.assignments?.forEach((assignment: CoachingAssignment) => {
        userIds.add(assignment.coach_id);
        userIds.add(assignment.athlete_id);
      });

      await loadUsers(Array.from(userIds));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load coaching assignments');
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

  const acceptRequest = async (assignmentId: string) => {
    try {
      const response = await fetch(`/api/v1/coaches/assignments/${assignmentId}/accept`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to accept request');
      }

      await loadAssignments();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to accept request');
    }
  };

  const terminateAssignment = async (assignmentId: string, reason?: string) => {
    if (!window.confirm('Are you sure you want to end this coaching assignment?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/coaches/assignments/${assignmentId}`, {
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
        throw new Error('Failed to end coaching assignment');
      }

      await loadAssignments();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to end coaching assignment');
    }
  };

  const filteredAssignments = assignments.filter(assignment => {
    if (activeTab === 'all') return true;
    return assignment.status === activeTab;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600 dark:text-gray-400">Loading coaching assignments...</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-6">
        Coaching Assignments
      </h1>

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
                {assignments.filter(a => a.status === tab.key).length}
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

      {filteredAssignments.length === 0 ? (
        <div className="text-center py-12 bg-white dark:bg-gray-800 rounded-lg">
          <p className="text-gray-600 dark:text-gray-400">
            {activeTab === 'pending'
              ? 'No pending requests'
              : activeTab === 'active'
              ? 'No active coaching assignments'
              : 'No coaching assignments yet'}
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredAssignments.map((assignment) => {
            const otherUser = users.get(
              localStorage.getItem('userType') === 'coach' ? assignment.athlete_id : assignment.coach_id
            );

            return (
              <div
                key={assignment.id}
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
                          assignment.status === 'pending'
                            ? 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-300'
                            : assignment.status === 'active'
                            ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300'
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                        }`}
                      >
                        {assignment.status.charAt(0).toUpperCase() + assignment.status.slice(1)}
                      </span>
                    </div>

                    {assignment.request_message && (
                      <p className="text-gray-600 dark:text-gray-300 text-sm mb-3 italic">
                        "{assignment.request_message}"
                      </p>
                    )}

                    <div className="flex flex-wrap gap-4 text-sm text-gray-500 dark:text-gray-400">
                      <span>Requested: {new Date(assignment.requested_at).toLocaleDateString()}</span>
                      {assignment.accepted_at && (
                        <span>Accepted: {new Date(assignment.accepted_at).toLocaleDateString()}</span>
                      )}
                      {assignment.terminated_at && (
                        <span>Ended: {new Date(assignment.terminated_at).toLocaleDateString()}</span>
                      )}
                    </div>

                    {assignment.termination_reason && (
                      <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                        Reason: {assignment.termination_reason}
                      </p>
                    )}
                  </div>

                  <div className="flex gap-2">
                    {assignment.status === 'pending' && localStorage.getItem('userType') === 'coach' && (
                      <>
                        <button
                          onClick={() => acceptRequest(assignment.id)}
                          className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg
                                   font-medium transition-colors text-sm"
                        >
                          Accept
                        </button>
                        <button
                          onClick={() => terminateAssignment(assignment.id, 'Request declined')}
                          className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg
                                   font-medium transition-colors text-sm"
                        >
                          Decline
                        </button>
                      </>
                    )}
                    {assignment.status === 'active' && (
                      <button
                        onClick={() => terminateAssignment(assignment.id)}
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
