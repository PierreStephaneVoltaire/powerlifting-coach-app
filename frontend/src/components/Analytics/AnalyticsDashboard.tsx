import React, { useState, useEffect } from 'react';
import { VolumeChart } from './VolumeChart';
import { E1RMChart } from './E1RMChart';
import { api } from '../../utils/apiWrapper';

export const AnalyticsDashboard: React.FC = () => {
  const [volumeData, setVolumeData] = useState<any[]>([]);
  const [e1rmData, setE1rmData] = useState<any[]>([]);
  const [timeRange, setTimeRange] = useState(30); // days
  const [selectedLift, setSelectedLift] = useState<'squat' | 'bench' | 'deadlift' | 'all'>('all');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAnalyticsData();
  }, [timeRange, selectedLift]);

  const fetchAnalyticsData = async () => {
    setLoading(true);
    try {
      const endDate = new Date();
      const startDate = new Date();
      startDate.setDate(startDate.getDate() - timeRange);

      // Fetch volume data
      const volumeResponse = await api.post('/analytics/volume', {
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString(),
        lift_type: selectedLift !== 'all' ? selectedLift : null,
      });
      setVolumeData(volumeResponse.data.volume_data || []);

      // Fetch e1RM data
      const e1rmResponse = await api.post('/analytics/e1rm', {
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString(),
        lift_type: selectedLift !== 'all' ? selectedLift : null,
      });
      setE1rmData(e1rmResponse.data.e1rm_data || []);
    } catch (error) {
      console.error('Failed to fetch analytics data:', error);
    } finally {
      setLoading(false);
    }
  };

  const calculateTotalVolume = () => {
    return volumeData.reduce((sum, item) => sum + item.total_volume, 0).toFixed(0);
  };

  const calculateMaxE1RM = () => {
    if (e1rmData.length === 0) return 0;
    return Math.max(...e1rmData.map(item => item.estimated_1rm)).toFixed(1);
  };

  const calculateAverageRPE = () => {
    const validRPE = volumeData.filter(item => item.average_rpe);
    if (validRPE.length === 0) return 0;
    return (validRPE.reduce((sum, item) => sum + (item.average_rpe || 0), 0) / validRPE.length).toFixed(1);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-8 text-gray-900 dark:text-white">Training Analytics</h1>

      {/* Filters */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Time Range</label>
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(parseInt(e.target.value))}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
            >
              <option value={7}>Last 7 days</option>
              <option value={30}>Last 30 days</option>
              <option value={90}>Last 90 days</option>
              <option value={180}>Last 6 months</option>
              <option value={365}>Last year</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Lift Type</label>
            <select
              value={selectedLift}
              onChange={(e) => setSelectedLift(e.target.value as any)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md"
            >
              <option value="all">All Lifts</option>
              <option value="squat">Squat</option>
              <option value="bench">Bench Press</option>
              <option value="deadlift">Deadlift</option>
            </select>
          </div>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      ) : (
        <>
          {/* Summary Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">Total Volume</h3>
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{calculateTotalVolume()} kg</p>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Last {timeRange} days</p>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">Max Estimated 1RM</h3>
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{calculateMaxE1RM()} kg</p>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Based on working sets</p>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">Average RPE</h3>
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{calculateAverageRPE()}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Across all sessions</p>
            </div>
          </div>

          {/* Charts */}
          <div className="space-y-8">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Volume Over Time</h2>
              <VolumeChart data={volumeData} />
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Estimated 1RM Progression</h2>
              <E1RMChart data={e1rmData} />
            </div>
          </div>

          {/* Exercise Breakdown */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mt-8">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Exercise Breakdown</h2>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Exercise
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Total Sets
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Total Reps
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Volume (kg)
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Avg Weight
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Avg RPE
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  {volumeData
                    .reduce((acc: any[], curr) => {
                      const existing = acc.find(item => item.exercise_name === curr.exercise_name);
                      if (existing) {
                        existing.total_sets += curr.total_sets;
                        existing.total_reps += curr.total_reps;
                        existing.total_volume += curr.total_volume;
                        existing.count += 1;
                        existing.avg_weight = (existing.avg_weight * (existing.count - 1) + curr.average_weight) / existing.count;
                        if (curr.average_rpe) {
                          existing.total_rpe = (existing.total_rpe || 0) + curr.average_rpe;
                          existing.rpe_count = (existing.rpe_count || 0) + 1;
                        }
                      } else {
                        acc.push({
                          exercise_name: curr.exercise_name,
                          total_sets: curr.total_sets,
                          total_reps: curr.total_reps,
                          total_volume: curr.total_volume,
                          avg_weight: curr.average_weight,
                          total_rpe: curr.average_rpe || 0,
                          rpe_count: curr.average_rpe ? 1 : 0,
                          count: 1,
                        });
                      }
                      return acc;
                    }, [])
                    .map((exercise: any) => (
                      <tr key={exercise.exercise_name}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                          {exercise.exercise_name}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                          {exercise.total_sets}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                          {exercise.total_reps}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                          {exercise.total_volume.toFixed(0)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                          {exercise.avg_weight.toFixed(1)} kg
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                          {exercise.rpe_count > 0 ? (exercise.total_rpe / exercise.rpe_count).toFixed(1) : 'N/A'}
                        </td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
