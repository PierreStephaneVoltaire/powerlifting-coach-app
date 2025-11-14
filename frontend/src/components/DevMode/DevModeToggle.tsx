import React from 'react';
import { useDevMode } from '@/context/DevModeContext';

export const DevModeToggle: React.FC = () => {
  const { devMode, toggleDevMode } = useDevMode();

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <button
        onClick={toggleDevMode}
        className={`px-4 py-2 rounded-full shadow-lg font-medium transition-all ${
          devMode
            ? 'bg-yellow-500 text-black hover:bg-yellow-600'
            : 'bg-gray-700 text-white hover:bg-gray-800'
        }`}
        title={devMode ? 'Dev Mode ON - Using fake data' : 'Dev Mode OFF - Using real backend'}
      >
        {devMode ? '🔧 DEV' : '🌐 PROD'}
      </button>
    </div>
  );
};
