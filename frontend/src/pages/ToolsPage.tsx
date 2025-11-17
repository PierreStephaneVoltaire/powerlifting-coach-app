import React, { useState } from 'react';
import { PlateCalculator } from '@/components/Tools/PlateCalculator';
import { MachineNotes } from '@/components/Tools/MachineNotes';

export const ToolsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'calculator' | 'notes'>('calculator');

  return (
    <div className="max-w-4xl mx-auto p-4">
      <h2 className="text-2xl font-bold mb-6">Tools</h2>

      <div className="flex gap-4 mb-6 border-b">
        <button
          onClick={() => setActiveTab('calculator')}
          className={`pb-2 px-4 ${
            activeTab === 'calculator'
              ? 'border-b-2 border-blue-500 text-blue-600 font-semibold'
              : 'text-gray-600'
          }`}
        >
          Plate Calculator
        </button>
        <button
          onClick={() => setActiveTab('notes')}
          className={`pb-2 px-4 ${
            activeTab === 'notes'
              ? 'border-b-2 border-blue-500 text-blue-600 font-semibold'
              : 'text-gray-600'
          }`}
        >
          Machine Notes
        </button>
      </div>

      {activeTab === 'calculator' && <PlateCalculator />}
      {activeTab === 'notes' && <MachineNotes />}
    </div>
  );
};
