import React, { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { apiClient } from '@/utils/api';

export const OnboardingCheck: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, onboarded } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();
  const [isCheckingProgram, setIsCheckingProgram] = useState(true);
  const [hasPendingProgram, setHasPendingProgram] = useState(false);

  useEffect(() => {
    const checkOnboardingAndProgram = async () => {
      if (!isAuthenticated) {
        setIsCheckingProgram(false);
        return;
      }

      // First check: Is user onboarded?
      if (!onboarded) {
        navigate('/onboarding');
        setIsCheckingProgram(false);
        return;
      }

      // Second check: Does user have a program?
      try {
        // Check for active approved program
        const { has_program } = await apiClient.getActiveProgram();

        if (has_program) {
          // User has an approved program, let them access the app
          setIsCheckingProgram(false);
          return;
        }

        // Check for pending program
        const { has_pending } = await apiClient.getPendingProgram();

        if (has_pending) {
          // User has a pending program awaiting approval
          setHasPendingProgram(true);

          // If not already on program page, redirect there to show approval UI
          if (!location.pathname.startsWith('/program')) {
            navigate('/program');
          }
        } else {
          // No program at all - redirect to chat to create one
          if (location.pathname !== '/chat') {
            navigate('/chat');
          }
        }
      } catch (error) {
        console.error('Failed to check program status:', error);
        // On error, allow access but log the issue
      } finally {
        setIsCheckingProgram(false);
      }
    };

    checkOnboardingAndProgram();
  }, [isAuthenticated, onboarded, navigate, location.pathname]);

  if (isCheckingProgram) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};
