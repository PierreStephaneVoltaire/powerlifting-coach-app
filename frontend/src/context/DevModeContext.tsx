import React, { createContext, useContext, useState, useEffect } from 'react';

interface DevModeContextType {
  devMode: boolean;
  toggleDevMode: () => void;
}

const DevModeContext = createContext<DevModeContextType | undefined>(undefined);

export const DevModeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [devMode, setDevMode] = useState(() => {
    const saved = localStorage.getItem('powercoach_dev_mode');
    return saved === 'true';
  });

  useEffect(() => {
    localStorage.setItem('powercoach_dev_mode', String(devMode));
    // Why: Reload app when dev mode changes to ensure clean state
    if (devMode !== (localStorage.getItem('powercoach_dev_mode') === 'true')) {
      window.location.reload();
    }
  }, [devMode]);

  const toggleDevMode = () => {
    setDevMode(prev => !prev);
  };

  return (
    <DevModeContext.Provider value={{ devMode, toggleDevMode }}>
      {children}
    </DevModeContext.Provider>
  );
};

export const useDevMode = () => {
  const context = useContext(DevModeContext);
  if (!context) {
    throw new Error('useDevMode must be used within DevModeProvider');
  }
  return context;
};
