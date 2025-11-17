import React, { useState, useEffect } from 'react';
import Calendar from 'react-calendar';
import { apiClient } from '../../utils/api';
import { format } from 'date-fns';
import 'react-calendar/dist/Calendar.css';

interface WorkoutSession {
  id: string;
  session_name: string;
  scheduled_date: string;
  completed_at: string;
  notes?: string;
  rpe_rating?: number;
  duration_minutes?: number;
  exercises: any[];
}

export const WorkoutHistory: React.FC = () => {
  const [viewMode, setViewMode] = useState<'calendar' | 'list'>('list');
  const [sessions, setSessions] = useState<WorkoutSession[]>([]);
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [loading, setLoading] = useState(true);
  const [selectedSession, setSelectedSession] = useState<WorkoutSession | null>(null);

  useEffect(() => {
    fetchSessionHistory();
  }, []);

  const fetchSessionHistory = async () => {
    setLoading(true);
    try {
      const endDate = new Date();
      const startDate = new Date();
      startDate.setMonth(startDate.getMonth() - 3); // Last 3 months

      const response = await apiClient.get(`/sessions/history?start_date=${startDate.toISOString()}&end_date=${endDate.toISOString()}`);
      setSessions(response.data.sessions || []);
    } catch (error) {
      console.error('Failed to fetch session history:', error);
    } finally {
      setLoading(false);
    }
  };

  const getSessionsForDate = (date: Date) => {
    return sessions.filter(session => {
      const sessionDate = new Date(session.completed_at);
      return sessionDate.toDateString() === date.toDateString();
    });
  };

  const tileContent = ({ date, view }: any) => {
    if (view === 'month') {
      const daySessions = getSessionsForDate(date);
      if (daySessions.length > 0) {
        return (
          <div className="flex justify-center mt-1">
            <div className="w-2 h-2 rounded-full bg-blue-600"></div>
          </div>
        );
      }
    }
    return null;
  };

  const calculateVolume = (session: WorkoutSession) => {
    let totalVolume = 0;
    session.exercises?.forEach(exercise => {
      exercise.completed_sets?.forEach((set: any) => {
        totalVolume += set.weight_kg * set.reps_completed;
      });
    });
    return totalVolume.toFixed(0);
  };

  const calculateSets = (session: WorkoutSession) => {
    return session.exercises?.reduce((total, exercise) => total + (exercise.completed_sets?.length || 0), 0) || 0;
  };

  const handleDeleteSession = async (sessionId: string) => {
    if (!window.confirm('Are you sure you want to delete this workout? This action cannot be undone.')) {
      return;
    }

    try {
      await apiClient.delete(`/sessions/${sessionId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      setSessions(sessions.filter(s => s.id !== sessionId));
    } catch (error) {
      console.error('Failed to delete session:', error);
      alert('Failed to delete workout. Please try again.');
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Workout History</h1>

        <div className="flex gap-2">
          <button
            onClick={() => setViewMode('list')}
            className={`px-4 py-2 rounded ${viewMode === 'list' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
          >
            List View
          </button>
          <button
            onClick={() => setViewMode('calendar')}
            className={`px-4 py-2 rounded ${viewMode === 'calendar' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
          >
            Calendar View
          </button>
        </div>
      </div>

      {viewMode === 'calendar' ? (
        <div className="bg-white rounded-lg shadow p-6">
          <Calendar
            onChange={(date: any) => setSelectedDate(date)}
            value={selectedDate}
            tileContent={tileContent}
            className="w-full"
          />

          {selectedDate && (
            <div className="mt-6">
              <h3 className="text-lg font-bold mb-4">
                Sessions on {format(selectedDate, 'MMMM d, yyyy')}
              </h3>

              {getSessionsForDate(selectedDate).length > 0 ? (
                <div className="space-y-4">
                  {getSessionsForDate(selectedDate).map(session => (
                    <div
                      key={session.id}
                      className="border border-gray-200 rounded-lg p-4 hover:shadow-lg cursor-pointer transition"
                      onClick={() => setSelectedSession(session)}
                    >
                      <div className="flex justify-between items-start">
                        <div>
                          <h4 className="font-semibold text-lg">{session.session_name}</h4>
                          <p className="text-sm text-gray-600">
                            {session.exercises?.length || 0} exercises • {calculateSets(session)} sets • {calculateVolume(session)} kg volume
                          </p>
                        </div>
                        {session.rpe_rating && (
                          <div className="text-right">
                            <div className="text-sm text-gray-600">RPE</div>
                            <div className="text-2xl font-bold">{session.rpe_rating}</div>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-8">No workouts on this date</p>
              )}
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {sessions.map(session => (
            <div
              key={session.id}
              className="bg-white rounded-lg shadow p-6 hover:shadow-lg cursor-pointer transition"
              onClick={() => setSelectedSession(session)}
            >
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-xl font-bold">{session.session_name}</h3>
                  <p className="text-sm text-gray-600">
                    {format(new Date(session.completed_at), 'MMMM d, yyyy • h:mm a')}
                  </p>
                </div>
                {session.rpe_rating && (
                  <div className="text-right">
                    <div className="text-sm text-gray-600">RPE</div>
                    <div className="text-2xl font-bold">{session.rpe_rating}</div>
                  </div>
                )}
              </div>

              <div className="grid grid-cols-3 gap-4 mb-4">
                <div>
                  <div className="text-sm text-gray-600">Exercises</div>
                  <div className="text-lg font-semibold">{session.exercises?.length || 0}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-600">Total Sets</div>
                  <div className="text-lg font-semibold">{calculateSets(session)}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-600">Volume</div>
                  <div className="text-lg font-semibold">{calculateVolume(session)} kg</div>
                </div>
              </div>

              {session.notes && (
                <div className="text-sm text-gray-600 italic border-t pt-3 mt-3">
                  {session.notes}
                </div>
              )}

              <div className="mt-4 flex gap-2">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelectedSession(session);
                  }}
                  className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                  View Details
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDeleteSession(session.id);
                  }}
                  className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}

          {sessions.length === 0 && (
            <div className="text-center py-16 text-gray-500">
              <p className="text-lg">No workout history yet</p>
              <p className="text-sm mt-2">Complete some workouts to see them here!</p>
            </div>
          )}
        </div>
      )}

      {/* Session Detail Modal */}
      {selectedSession && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
            <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
              <div>
                <h2 className="text-2xl font-bold">{selectedSession.session_name}</h2>
                <p className="text-sm text-gray-600">
                  {format(new Date(selectedSession.completed_at), 'MMMM d, yyyy • h:mm a')}
                </p>
              </div>
              <button
                onClick={() => setSelectedSession(null)}
                className="text-gray-400 hover:text-gray-600"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="flex-1 overflow-y-auto px-6 py-4">
              <div className="space-y-6">
                {selectedSession.exercises?.map((exercise, idx) => (
                  <div key={idx} className="border border-gray-200 rounded-lg p-4">
                    <h3 className="font-bold text-lg mb-3">{exercise.exercise_name}</h3>

                    <div className="space-y-2">
                      {exercise.completed_sets?.map((set: any, setIdx: number) => (
                        <div key={setIdx} className="flex items-center gap-4 p-3 bg-gray-50 rounded">
                          <div className="w-16 text-center font-semibold">Set {set.set_number}</div>
                          <div className="flex-1 grid grid-cols-4 gap-3 text-sm">
                            <div>
                              <span className="text-gray-600">Type:</span> {set.set_type || 'working'}
                            </div>
                            <div>
                              <span className="text-gray-600">Weight:</span> {set.weight_kg} kg
                            </div>
                            <div>
                              <span className="text-gray-600">Reps:</span> {set.reps_completed}
                            </div>
                            <div>
                              <span className="text-gray-600">RPE:</span> {set.rpe_actual || 'N/A'}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>

                    {exercise.notes && (
                      <div className="mt-3 text-sm text-gray-600 italic">
                        Notes: {exercise.notes}
                      </div>
                    )}
                  </div>
                ))}
              </div>

              {selectedSession.notes && (
                <div className="mt-6 p-4 bg-blue-50 rounded-lg">
                  <h4 className="font-semibold mb-2">Workout Notes</h4>
                  <p className="text-sm text-gray-700">{selectedSession.notes}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
