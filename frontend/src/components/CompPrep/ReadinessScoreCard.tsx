import React from 'react';

interface ReadinessScoreCardProps {
  score: number;
}

export const ReadinessScoreCard: React.FC<ReadinessScoreCardProps> = ({ score }) => {
  const getScoreColor = () => {
    if (score >= 85) return 'text-green-600';
    if (score >= 70) return 'text-yellow-600';
    return 'text-orange-600';
  };

  const getScoreLabel = () => {
    if (score >= 85) return 'Excellent';
    if (score >= 70) return 'Good';
    if (score >= 50) return 'Fair';
    return 'Needs Work';
  };

  const circumference = 2 * Math.PI * 56;
  const strokeDashoffset = circumference - (score / 100) * circumference;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Readiness Score</h3>
      <div className="relative w-32 h-32 mx-auto">
        <svg className="transform -rotate-90 w-32 h-32">
          <circle
            cx="64"
            cy="64"
            r="56"
            stroke="currentColor"
            strokeWidth="8"
            fill="transparent"
            className="text-gray-200 dark:text-gray-700"
          />
          <circle
            cx="64"
            cy="64"
            r="56"
            stroke="currentColor"
            strokeWidth="8"
            fill="transparent"
            strokeDasharray={circumference}
            strokeDashoffset={strokeDashoffset}
            className={getScoreColor()}
            style={{ transition: 'stroke-dashoffset 0.5s ease' }}
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center">
            <div className={`text-3xl font-bold ${getScoreColor()}`}>{score}</div>
            <div className="text-xs text-gray-600 dark:text-gray-400">/100</div>
          </div>
        </div>
      </div>
      <p className="text-center mt-4 font-medium text-gray-900 dark:text-white">{getScoreLabel()}</p>
    </div>
  );
};
