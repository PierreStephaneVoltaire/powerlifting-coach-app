import { create } from 'zustand';
import { apiClient } from '@/utils/api';

interface SyncState {
  pendingCount: number;
  isOnline: boolean;
  updatePendingCount: () => Promise<void>;
  setOnlineStatus: (status: boolean) => void;
}

export const useSyncStore = create<SyncState>((set) => ({
  pendingCount: 0,
  isOnline: navigator.onLine,

  updatePendingCount: async () => {
    const count = await apiClient.getPendingEventsCount();
    set({ pendingCount: count });
  },

  setOnlineStatus: (status: boolean) => {
    set({ isOnline: status });
  },
}));

window.addEventListener('online', () => {
  useSyncStore.getState().setOnlineStatus(true);
});

window.addEventListener('offline', () => {
  useSyncStore.getState().setOnlineStatus(false);
});

setInterval(() => {
  useSyncStore.getState().updatePendingCount();
}, 10000);
