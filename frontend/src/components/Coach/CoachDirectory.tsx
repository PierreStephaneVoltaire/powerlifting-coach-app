import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from '@/components/UI/Toast';

interface Coach {
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
  }>;
  success_stories: Array<{
    id: string;
    achievement: string;
    total_kg?: number;
  }>;
  created_at: string;
}

export const CoachDirectory: React.FC = () => {
  const [coaches, setCoaches] = useState<Coach[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    loadCoaches();
  }, []);

  const loadCoaches = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/coaches/directory', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        if (response.status === 404) {
          toast.warning('Coach directory feature is not yet implemented', 8000);
          setError('Coach directory is coming soon. This feature will allow you to browse and connect with certified powerlifting coaches.');
        } else {
          throw new Error('Failed to load coaches');
        }
        return;
      }

      const data = await response.json();
      setCoaches(data.coaches || []);
    } catch (err) {
      toast.warning('Coach directory feature is not yet implemented', 8000);
      setError('Coach directory is coming soon. This feature will allow you to browse and connect with certified powerlifting coaches.');
    } finally {
      setLoading(false);
    }
  };

  const filteredCoaches = coaches.filter(coach =>
    coach.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    coach.bio?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600 dark:text-gray-400">Loading coaches...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Find a Coach
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Connect with experienced powerlifting coaches to take your training to the next level
        </p>
      </div>

      <div className="mb-6">
        <input
          type="text"
          placeholder="Search coaches by name or specialty..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg
                   bg-white dark:bg-gray-800 text-gray-900 dark:text-white
                   focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <p className="text-red-600 dark:text-red-400">{error}</p>
        </div>
      )}

      {filteredCoaches.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-600 dark:text-gray-400">
            {searchTerm ? 'No coaches found matching your search' : 'No coaches available yet'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredCoaches.map((coach) => (
            <div
              key={coach.id}
              onClick={() => navigate(`/coaches/${coach.user_id}`)}
              className="bg-white dark:bg-gray-800 rounded-lg shadow-md hover:shadow-lg
                       transition-all cursor-pointer border border-gray-200 dark:border-gray-700
                       hover:border-blue-500 dark:hover:border-blue-400"
            >
              <div className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-1">
                      {coach.name}
                    </h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {coach.total_athletes} athlete{coach.total_athletes !== 1 ? 's' : ''}
                    </p>
                  </div>
                  <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600
                                rounded-full flex items-center justify-center text-white text-2xl font-bold">
                    {coach.name.charAt(0).toUpperCase()}
                  </div>
                </div>

                {coach.bio && (
                  <p className="text-gray-600 dark:text-gray-300 text-sm mb-4 line-clamp-3">
                    {coach.bio}
                  </p>
                )}

                {coach.certifications.length > 0 && (
                  <div className="mb-4">
                    <h4 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase mb-2">
                      Certifications
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {coach.certifications.slice(0, 2).map((cert) => (
                        <span
                          key={cert.id}
                          className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300
                                   rounded text-xs font-medium"
                        >
                          {cert.certification_name}
                        </span>
                      ))}
                      {coach.certifications.length > 2 && (
                        <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300
                                       rounded text-xs">
                          +{coach.certifications.length - 2} more
                        </span>
                      )}
                    </div>
                  </div>
                )}

                {coach.success_stories.length > 0 && (
                  <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {coach.success_stories.length} success stor{coach.success_stories.length !== 1 ? 'ies' : 'y'}
                    </p>
                  </div>
                )}

                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(`/coaches/${coach.user_id}`);
                  }}
                  className="mt-4 w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white
                           rounded-lg font-medium transition-colors"
                >
                  View Profile
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
