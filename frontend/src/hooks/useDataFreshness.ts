import { useState, useEffect } from 'react';

const STALE_THRESHOLD_MS = 5 * 60 * 1000; // 5 minutes

export const useDataFreshness = (lastFetchTimestamp: number | null) => {
  const [isStale, setIsStale] = useState(false);

  useEffect(() => {
    if (!lastFetchTimestamp) {
      setIsStale(false);
      return;
    }

    const checkStaleness = () => {
      const age = Date.now() - lastFetchTimestamp;
      setIsStale(age > STALE_THRESHOLD_MS);
    };

    checkStaleness();

    const interval = setInterval(checkStaleness, 30000); // Check every 30 seconds

    return () => clearInterval(interval);
  }, [lastFetchTimestamp]);

  return isStale;
};
