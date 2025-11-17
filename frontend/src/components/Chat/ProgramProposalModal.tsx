import React, { useState } from 'react';
import { format, addDays } from 'date-fns';

interface ProgramProposal {
  phases?: Array<{
    name: string;
    weeks: number[];
    focus: string;
    characteristics: string;
  }>;
  weeklyWorkouts?: Array<{
    week: number;
    workouts: Array<{
      day: number;
      name: string;
      exercises: Array<{
        name: string;
        liftType: string;
        sets: number;
        reps: string;
        intensity: string;
        rpe?: number;
        notes?: string;
      }>;
    }>;
  }>;
  summary?: {
    totalWeeks: number;
    trainingDaysPerWeek: number;
    peakWeek: number;
    competitionWeek: number;
  };
}

interface Props {
  proposal: ProgramProposal;
  competitionDate?: string;
  onApprove: (programData: any) => void;
  onClose: () => void;
  isLoading: boolean;
  isChangeProposal?: boolean;
}

export const ProgramProposalModal: React.FC<Props> = ({
  proposal,
  competitionDate,
  onApprove,
  onClose,
  isLoading,
  isChangeProposal = false,
}) => {
  const [programName, setProgramName] = useState('Competition Prep Program');
  const [startDate, setStartDate] = useState(format(new Date(), 'yyyy-MM-dd'));

  const totalWeeks = proposal.summary?.totalWeeks || 12;
  const daysPerWeek = proposal.summary?.trainingDaysPerWeek || 4;

  const handleApprove = () => {
    const programData = {
      name: programName,
      description: `AI-generated ${totalWeeks}-week program`,
      program_data: proposal,
      start_date: new Date(startDate).toISOString(),
      weeks_total: totalWeeks,
      days_per_week: daysPerWeek,
    };
    onApprove(programData);
  };

  const endDate = addDays(new Date(startDate), totalWeeks * 7);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
          <h2 className="text-xl font-bold text-gray-900">
            {isChangeProposal ? 'Review Program Changes' : 'Review Program Proposal'}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            disabled={isLoading}
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Program Name
              </label>
              <input
                type="text"
                value={programName}
                onChange={(e) => setProgramName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Start Date
              </label>
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h3 className="font-semibold text-blue-900 mb-2">Program Summary</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div>
                <span className="text-blue-700 font-medium">Duration:</span>
                <p className="text-blue-900">{totalWeeks} weeks</p>
              </div>
              <div>
                <span className="text-blue-700 font-medium">Training Days:</span>
                <p className="text-blue-900">{daysPerWeek} per week</p>
              </div>
              <div>
                <span className="text-blue-700 font-medium">End Date:</span>
                <p className="text-blue-900">{format(endDate, 'MMM d, yyyy')}</p>
              </div>
              {competitionDate && (
                <div>
                  <span className="text-blue-700 font-medium">Competition:</span>
                  <p className="text-blue-900">{format(new Date(competitionDate), 'MMM d, yyyy')}</p>
                </div>
              )}
            </div>
          </div>

          {proposal.phases && proposal.phases.length > 0 && (
            <div>
              <h3 className="font-semibold text-gray-900 mb-3">Training Phases</h3>
              <div className="space-y-3">
                {proposal.phases.map((phase, idx) => (
                  <div key={idx} className="bg-gray-50 rounded-lg p-4">
                    <div className="flex justify-between items-start">
                      <div>
                        <h4 className="font-medium text-gray-900">{phase.name}</h4>
                        <p className="text-sm text-gray-600 mt-1">{phase.focus}</p>
                      </div>
                      <span className="text-sm bg-gray-200 px-2 py-1 rounded">
                        Weeks {phase.weeks[0]}-{phase.weeks[phase.weeks.length - 1]}
                      </span>
                    </div>
                    <p className="text-xs text-gray-500 mt-2">{phase.characteristics}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {proposal.weeklyWorkouts && proposal.weeklyWorkouts.length > 0 && (
            <div>
              <h3 className="font-semibold text-gray-900 mb-3">
                Sample Week Structure (Week 1)
              </h3>
              <div className="grid gap-3">
                {proposal.weeklyWorkouts[0]?.workouts.map((workout, idx) => (
                  <div key={idx} className="bg-gray-50 rounded-lg p-3">
                    <h4 className="font-medium text-gray-900 mb-2">
                      Day {workout.day}: {workout.name}
                    </h4>
                    <div className="space-y-1">
                      {workout.exercises.slice(0, 3).map((exercise, exIdx) => (
                        <div key={exIdx} className="text-sm flex justify-between">
                          <span className="text-gray-700">{exercise.name}</span>
                          <span className="text-gray-500">
                            {exercise.sets}x{exercise.reps} @ {exercise.intensity}
                            {exercise.rpe && ` RPE ${exercise.rpe}`}
                          </span>
                        </div>
                      ))}
                      {workout.exercises.length > 3 && (
                        <p className="text-xs text-gray-400">
                          +{workout.exercises.length - 3} more exercises
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-gray-200 bg-gray-50 flex justify-end space-x-3">
          <button
            onClick={onClose}
            disabled={isLoading}
            className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
          >
            Continue Editing
          </button>
          <button
            onClick={handleApprove}
            disabled={isLoading || !programName.trim()}
            className="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 flex items-center"
          >
            {isLoading ? (
              <>
                <svg className="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                {isChangeProposal ? 'Proposing...' : 'Creating...'}
              </>
            ) : isChangeProposal ? (
              'Propose Changes'
            ) : (
              'Approve & Create Program'
            )}
          </button>
        </div>
      </div>
    </div>
  );
};
