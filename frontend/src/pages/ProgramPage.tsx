import React from 'react';
import { ProgramPlanner } from '@/components/Program/ProgramPlanner';

export const ProgramPage: React.FC = () => {
  return (
    <div className="max-w-6xl mx-auto p-4">
      <h2 className="text-2xl font-bold mb-6">Program Planner</h2>
      <ProgramPlanner />
    </div>
  );
};
