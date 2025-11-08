import React from 'react';

export const ProgramPage: React.FC = () => {
  return (
    <div className="max-w-6xl mx-auto p-4">
      <h2 className="text-2xl font-bold mb-6">My Program</h2>
      <div className="bg-white shadow rounded-lg p-8">
        <div className="text-center text-gray-500">
          <h3 className="text-xl font-semibold mb-4">Program Overview</h3>
          <p className="mb-4">
            Your training program will be generated through conversation with your AI coach.
          </p>
          <p className="text-sm">
            Features coming soon:
          </p>
          <ul className="text-sm text-left max-w-md mx-auto mt-2 space-y-2">
            <li>• View current program summary</li>
            <li>• See upcoming workouts</li>
            <li>• Track program progress</li>
            <li>• View workout history</li>
          </ul>
        </div>
      </div>
    </div>
  );
};
