import React, { useState } from 'react';

interface ProgramArtifactProps {
  artifact: any;
  onClose: () => void;
  onApprove?: (program: any) => void;
}

export const ProgramArtifact: React.FC<ProgramArtifactProps> = ({
  artifact,
  onClose,
  onApprove,
}) => {
  const [selectedTab, setSelectedTab] = useState<'overview' | 'progression' | 'timeline'>('overview');

  const handleApprove = () => {
    if (onApprove) {
      onApprove(artifact.program);
    }
    onClose();
  };

  const renderWeeklyProgression = () => {
    if (!artifact.program?.weeklyWorkouts) return null;

    const weeks = Object.keys(artifact.program.weeklyWorkouts).sort((a, b) => {
      const weekA = parseInt(a.replace('week', ''));
      const weekB = parseInt(b.replace('week', ''));
      return weekA - weekB;
    });

    return (
      <div className="space-y-4">
        {weeks.map((weekKey) => {
          const week = artifact.program.weeklyWorkouts[weekKey];
          return (
            <div key={weekKey} className="border border-gray-200 rounded-lg p-4">
              <h4 className="font-bold text-lg mb-2 capitalize">{weekKey.replace(/([A-Z])/g, ' $1')}</h4>

              {week.days && week.days.map((day: any, idx: number) => (
                <div key={idx} className="mt-3 p-3 bg-gray-50 rounded">
                  <div className="font-semibold text-sm mb-2">{day.name || `Day ${idx + 1}`}</div>
                  <div className="space-y-1 text-sm">
                    {day.exercises?.map((exercise: any, exIdx: number) => (
                      <div key={exIdx} className="flex justify-between items-center">
                        <span className="text-gray-700">{exercise.name}</span>
                        <span className="text-gray-600">
                          {exercise.sets}×{exercise.reps} @ {exercise.intensity || exercise.rpe}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          );
        })}
      </div>
    );
  };

  const renderPeakingTimeline = () => {
    const phases = [
      { name: 'Volume', weeks: '1-4', color: 'bg-blue-500', description: 'Build work capacity' },
      { name: 'Strength', weeks: '5-8', color: 'bg-purple-500', description: 'Increase intensity' },
      { name: 'Peaking', weeks: '9-11', color: 'bg-orange-500', description: 'Competition prep' },
      { name: 'Taper', weeks: '12', color: 'bg-green-500', description: 'Deload and recovery' },
    ];

    return (
      <div className="space-y-4">
        <h4 className="font-bold text-lg">Training Phases</h4>
        <div className="relative">
          {phases.map((phase, idx) => (
            <div key={idx} className="flex items-center gap-4 mb-4">
              <div className={`w-16 h-16 ${phase.color} rounded-full flex items-center justify-center text-white font-bold shrink-0`}>
                {idx + 1}
              </div>
              <div className="flex-1">
                <div className="font-semibold">{phase.name}</div>
                <div className="text-sm text-gray-600">Weeks {phase.weeks}</div>
                <div className="text-sm text-gray-500">{phase.description}</div>
              </div>
            </div>
          ))}
        </div>

        {artifact.program?.overview && (
          <div className="mt-6 p-4 bg-blue-50 rounded-lg">
            <h5 className="font-semibold mb-2">Program Notes</h5>
            <p className="text-sm text-gray-700">{artifact.program.overview}</p>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-blue-600 to-purple-600 text-white">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-2xl font-bold">{artifact.name || 'Program Preview'}</h2>
              <p className="text-sm mt-1 opacity-90">{artifact.description || 'AI-generated competition prep program'}</p>
            </div>
            <button onClick={onClose} className="text-white hover:text-gray-200">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-200">
          <button
            onClick={() => setSelectedTab('overview')}
            className={`px-6 py-3 font-medium ${
              selectedTab === 'overview'
                ? 'border-b-2 border-blue-600 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Overview
          </button>
          <button
            onClick={() => setSelectedTab('progression')}
            className={`px-6 py-3 font-medium ${
              selectedTab === 'progression'
                ? 'border-b-2 border-blue-600 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Weekly Progression
          </button>
          <button
            onClick={() => setSelectedTab('timeline')}
            className={`px-6 py-3 font-medium ${
              selectedTab === 'timeline'
                ? 'border-b-2 border-blue-600 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Peaking Timeline
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {selectedTab === 'overview' && (
            <div className="space-y-6">
              {artifact.program?.name && (
                <div>
                  <h3 className="text-xl font-bold mb-2">{artifact.program.name}</h3>
                  {artifact.program.description && (
                    <p className="text-gray-700">{artifact.program.description}</p>
                  )}
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                {artifact.program?.duration && (
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <div className="text-sm text-gray-600">Duration</div>
                    <div className="text-2xl font-bold">{artifact.program.duration} weeks</div>
                  </div>
                )}
                {artifact.program?.daysPerWeek && (
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <div className="text-sm text-gray-600">Training Days</div>
                    <div className="text-2xl font-bold">{artifact.program.daysPerWeek} per week</div>
                  </div>
                )}
                {artifact.program?.phase && (
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <div className="text-sm text-gray-600">Current Phase</div>
                    <div className="text-2xl font-bold capitalize">{artifact.program.phase}</div>
                  </div>
                )}
                {artifact.program?.experience && (
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <div className="text-sm text-gray-600">Experience Level</div>
                    <div className="text-2xl font-bold capitalize">{artifact.program.experience}</div>
                  </div>
                )}
              </div>

              {artifact.program?.keyPoints && (
                <div>
                  <h4 className="font-bold mb-3">Key Program Features</h4>
                  <ul className="space-y-2">
                    {artifact.program.keyPoints.map((point: string, idx: number) => (
                      <li key={idx} className="flex items-start gap-2">
                        <svg className="w-5 h-5 text-green-600 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                        <span className="text-gray-700">{point}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {selectedTab === 'progression' && renderWeeklyProgression()}
          {selectedTab === 'timeline' && renderPeakingTimeline()}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 bg-gray-50 flex justify-between items-center">
          <button
            onClick={onClose}
            className="px-4 py-2 border border-gray-300 rounded-md hover:bg-white"
          >
            Close
          </button>

          <div className="flex gap-3">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-white"
            >
              Request Changes
            </button>
            <button
              onClick={handleApprove}
              className="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
            >
              Approve Program
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
