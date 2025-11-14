import React from 'react';
import { format } from 'date-fns';

interface CompPrepHeaderProps {
  competitionName: string;
  competitionDate: string;
  weeksUntil: number | null;
  daysUntil: number | null;
}

export const CompPrepHeader: React.FC<CompPrepHeaderProps> = ({
  competitionName,
  competitionDate,
  weeksUntil,
  daysUntil,
}) => {
  return (
    <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg shadow-lg p-8 mb-8 text-white">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-2">{competitionName}</h1>
        <p className="text-xl mb-4">{format(new Date(competitionDate), 'MMMM dd, yyyy')}</p>
        <div className="flex justify-center gap-8">
          <div>
            <div className="text-5xl font-bold">{weeksUntil}</div>
            <div className="text-sm uppercase tracking-wide">Weeks Out</div>
          </div>
          <div>
            <div className="text-5xl font-bold">{daysUntil}</div>
            <div className="text-sm uppercase tracking-wide">Days Out</div>
          </div>
        </div>
      </div>
    </div>
  );
};
