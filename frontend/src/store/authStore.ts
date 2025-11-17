import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { AuthTokens, User, UserResponse } from '@/types';

interface AuthState {
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  onboarded: boolean;
  login: (tokens: AuthTokens, userResponse: UserResponse) => void;
  logout: () => void;
  updateUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
  setOnboarded: (onboarded: boolean) => void;
  refreshTokens: (tokens: AuthTokens) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      tokens: null,
      isAuthenticated: false,
      isLoading: false,
      onboarded: false,

      login: (tokens: AuthTokens, userResponse: UserResponse) => {
        set({
          user: userResponse.user,
          tokens,
          isAuthenticated: true,
          isLoading: false,
        });
      },

      logout: () => {
        // Clear persisted auth storage immediately
        localStorage.removeItem('auth-storage');
        // Clear any other auth-related items
        localStorage.removeItem('token');
        // Clear cookies
        document.cookie.split(";").forEach((c) => {
          document.cookie = c
            .replace(/^ +/, "")
            .replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
        });
        set({
          user: null,
          tokens: null,
          isAuthenticated: false,
          isLoading: false,
          onboarded: false,
        });
      },

      updateUser: (user: User) => {
        set({ user });
      },

      setLoading: (isLoading: boolean) => {
        set({ isLoading });
      },

      setOnboarded: (onboarded: boolean) => {
        set({ onboarded });
      },

      refreshTokens: (tokens: AuthTokens) => {
        set({ tokens });
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        tokens: state.tokens,
        isAuthenticated: state.isAuthenticated,
        onboarded: state.onboarded,
      }),
    }
  )
);