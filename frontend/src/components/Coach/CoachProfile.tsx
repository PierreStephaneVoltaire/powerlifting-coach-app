import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

interface CoachProfileData {
  id: string;
  user_id: string;
  name: string;
  email: string;
  bio?: string;
  total_athletes: number;
  certifications: Array<{
    id: string;
    certification_name: string;
    issuing_organization?: string;
    issue_date?: string;
    expiry_date?: string;
    verification_status: string;
  }>;
  success_stories: Array<{
    id: string;
    athlete_name?: string;
    achievement: string;
    competition_name?: string;
    competition_date?: string;
    total_kg?: number;
    weight_class?: string;
    federation?: string;
    placement?: number;
  }>;
  created_at: string;
}

interface Relationship {
  id: string;
  status: 'pending' | 'active' | 'terminated';
}

export const CoachProfile: React.FC = () => {
  const { coachId } = useParams<{ coachId: string }>();
  const navigate = useNavigate();
  const [coach, setCoach] = useState<CoachProfileData | null>(null);
  const [relationship, setRelationship] = useState<Relationship | null>(null);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [requestMessage, setRequestMessage] = useState('');
  const [showRequestModal, setShowRequestModal] = useState(false);

  useEffect(() => {
    if (coachId) {
      loadCoachProfile();
      checkExistingRelationship();
    }
  }, [coachId]);

  const loadCoachProfile = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/coaches/profile/${coachId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load coach profile');
      }

      const data = await response.json();
      setCoach(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load coach profile');
    } finally {
      setLoading(false);
    }
  };

  const checkExistingRelationship = async () => {
    try {
      const response = await fetch('/api/v1/relationships', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        const existing = data.relationships?.find(
          (rel: any) => rel.coach_id === coachId
        );
        if (existing) {
          setRelationship(existing);
        }
      }
    } catch (err) {
      console.error('Failed to check existing relationship:', err);
    }
  };

  const sendRelationshipRequest = async () => {
    try {
      setSending(true);
      const response = await fetch('/api/v1/relationships', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          coach_id: coachId,
          request_message: requestMessage || undefined,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to send relationship request');
      }

      const data = await response.json();
      setRelationship(data);
      setShowRequestModal(false);
      setRequestMessage('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send request');
    } finally {
      setSending(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600 dark:text-gray-400">Loading coach profile...</div>
      </div>
    );
  }

  if (!coach) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen">
        <p className="text-gray-600 dark:text-gray-400 mb-4">Coach not found</p>
        <button
          onClick={() => navigate('/coaches')}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg"
        >
          Back to Directory
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8 mb-6">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-6">
            <div className="w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600
                          rounded-full flex items-center justify-center text-white text-4xl font-bold">
              {coach.name.charAt(0).toUpperCase()}
            </div>
            <div>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
                {coach.name}
              </h1>
              <p className="text-gray-600 dark:text-gray-400">
                Coaching {coach.total_athletes} athlete{coach.total_athletes !== 1 ? 's' : ''}
              </p>
            </div>
          </div>

          <div>
            {!relationship && (
              <button
                onClick={() => setShowRequestModal(true)}
                className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg
                         font-medium transition-colors"
              >
                Request Coaching
              </button>
            )}
            {relationship?.status === 'pending' && (
              <div className="px-6 py-3 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-300
                            rounded-lg font-medium">
                Request Pending
              </div>
            )}
            {relationship?.status === 'active' && (
              <div className="px-6 py-3 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300
                            rounded-lg font-medium">
                Active Coaching
              </div>
            )}
          </div>
        </div>

        {coach.bio && (
          <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">About</h2>
            <p className="text-gray-600 dark:text-gray-300 whitespace-pre-wrap">{coach.bio}</p>
          </div>
        )}
      </div>

      {/* Certifications */}
      {coach.certifications.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8 mb-6">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Certifications</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {coach.certifications.map((cert) => (
              <div
                key={cert.id}
                className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-semibold text-gray-900 dark:text-white">
                    {cert.certification_name}
                  </h3>
                  {cert.verification_status === 'verified' && (
                    <span className="px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300
                                   rounded text-xs font-medium">
                      Verified
                    </span>
                  )}
                </div>
                {cert.issuing_organization && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">
                    {cert.issuing_organization}
                  </p>
                )}
                {cert.issue_date && (
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Issued: {new Date(cert.issue_date).toLocaleDateString()}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Success Stories */}
      {coach.success_stories.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Success Stories</h2>
          <div className="space-y-4">
            {coach.success_stories.map((story) => (
              <div
                key={story.id}
                className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <p className="text-gray-900 dark:text-white font-medium mb-2">
                  {story.achievement}
                </p>
                <div className="flex flex-wrap gap-4 text-sm text-gray-600 dark:text-gray-400">
                  {story.athlete_name && <span>Athlete: {story.athlete_name}</span>}
                  {story.total_kg && <span className="font-semibold">{story.total_kg}kg Total</span>}
                  {story.weight_class && <span>{story.weight_class}</span>}
                  {story.placement && <span>#{story.placement} Place</span>}
                  {story.federation && <span>{story.federation}</span>}
                </div>
                {story.competition_name && (
                  <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                    {story.competition_name}
                    {story.competition_date && ` - ${new Date(story.competition_date).toLocaleDateString()}`}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Request Modal */}
      {showRequestModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full p-6">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              Request Coaching
            </h2>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Send a coaching request to {coach.name}. Include a message about your goals and experience.
            </p>
            <textarea
              value={requestMessage}
              onChange={(e) => setRequestMessage(e.target.value)}
              placeholder="Tell the coach about your training goals, experience, and why you'd like to work with them..."
              rows={6}
              className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent mb-4"
            />
            {error && (
              <p className="text-red-600 dark:text-red-400 text-sm mb-4">{error}</p>
            )}
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowRequestModal(false);
                  setRequestMessage('');
                  setError(null);
                }}
                disabled={sending}
                className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600
                         text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700
                         disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={sendRelationshipRequest}
                disabled={sending}
                className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg
                         disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {sending ? 'Sending...' : 'Send Request'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
