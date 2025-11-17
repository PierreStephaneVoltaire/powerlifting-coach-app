import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

interface AthleteStats {
  total_workouts: number;
  total_volume_kg: number;
  current_squat_max: number;
  current_bench_max: number;
  current_deadlift_max: number;
  total: number;
}

interface FeedPost {
  id: string;
  content: string;
  created_at: string;
  workout_data?: any;
}

interface AthleteProfileData {
  id: string;
  name: string;
  email: string;
  bio?: string;
  weight_class?: string;
  federation?: string;
  stats: AthleteStats;
  recent_posts: FeedPost[];
  joined_at: string;
}

export const AthleteProfile: React.FC = () => {
  const { athleteId } = useParams<{ athleteId: string }>();
  const [athlete, setAthlete] = useState<AthleteProfileData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'feed' | 'stats'>('feed');

  useEffect(() => {
    if (athleteId) {
      loadAthleteProfile();
    }
  }, [athleteId]);

  const loadAthleteProfile = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/athletes/profile/${athleteId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        if (response.status === 403) {
          throw new Error('This profile is private');
        }
        throw new Error('Failed to load athlete profile');
      }

      const data = await response.json();
      setAthlete(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load profile');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600 dark:text-gray-400">Loading profile...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen px-4">
        <p className="text-red-600 dark:text-red-400 mb-4">{error}</p>
      </div>
    );
  }

  if (!athlete) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-600 dark:text-gray-400">Athlete not found</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 md:p-8 mb-6">
        <div className="flex flex-col md:flex-row items-start gap-6">
          <div className="w-20 h-20 md:w-24 md:h-24 bg-gradient-to-br from-purple-500 to-pink-600
                        rounded-full flex items-center justify-center text-white text-3xl md:text-4xl font-bold">
            {athlete.name.charAt(0).toUpperCase()}
          </div>
          <div className="flex-1">
            <h1 className="text-2xl md:text-3xl font-bold text-gray-900 dark:text-white mb-2">
              {athlete.name}
            </h1>
            {athlete.bio && (
              <p className="text-gray-600 dark:text-gray-300 mb-4">{athlete.bio}</p>
            )}
            <div className="flex flex-wrap gap-3 text-sm">
              {athlete.weight_class && (
                <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full">
                  {athlete.weight_class}
                </span>
              )}
              {athlete.federation && (
                <span className="px-3 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded-full">
                  {athlete.federation}
                </span>
              )}
              <span className="px-3 py-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-full">
                Joined {new Date(athlete.joined_at).toLocaleDateString()}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Squat</div>
          <div className="text-2xl font-bold text-gray-900 dark:text-white">
            {athlete.stats.current_squat_max}kg
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Bench</div>
          <div className="text-2xl font-bold text-gray-900 dark:text-white">
            {athlete.stats.current_bench_max}kg
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Deadlift</div>
          <div className="text-2xl font-bold text-gray-900 dark:text-white">
            {athlete.stats.current_deadlift_max}kg
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Total</div>
          <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
            {athlete.stats.total}kg
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-6 border-b border-gray-200 dark:border-gray-700">
        {[
          { key: 'feed', label: 'Recent Activity' },
          { key: 'stats', label: 'Training Stats' },
        ].map((tab) => (
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
          </button>
        ))}
      </div>

      {/* Content */}
      {activeTab === 'feed' && (
        <div className="space-y-4">
          {athlete.recent_posts.length === 0 ? (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8 text-center">
              <p className="text-gray-600 dark:text-gray-400">No recent activity</p>
            </div>
          ) : (
            athlete.recent_posts.map((post) => (
              <div
                key={post.id}
                className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6"
              >
                <p className="text-gray-900 dark:text-white mb-2">{post.content}</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {new Date(post.created_at).toLocaleDateString()}
                </p>
              </div>
            ))
          )}
        </div>
      )}

      {activeTab === 'stats' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Training Volume
              </h3>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Total Workouts:</span>
                  <span className="font-medium text-gray-900 dark:text-white">
                    {athlete.stats.total_workouts}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Total Volume:</span>
                  <span className="font-medium text-gray-900 dark:text-white">
                    {athlete.stats.total_volume_kg.toLocaleString()}kg
                  </span>
                </div>
              </div>
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Current Maxes
              </h3>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Squat:</span>
                  <span className="font-medium text-gray-900 dark:text-white">
                    {athlete.stats.current_squat_max}kg
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Bench:</span>
                  <span className="font-medium text-gray-900 dark:text-white">
                    {athlete.stats.current_bench_max}kg
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Deadlift:</span>
                  <span className="font-medium text-gray-900 dark:text-white">
                    {athlete.stats.current_deadlift_max}kg
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
