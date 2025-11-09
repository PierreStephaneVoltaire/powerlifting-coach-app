import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';

export const OnboardingCheck: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, onboarded } = useAuthStore();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated && !onboarded) {
      navigate('/onboarding');
    }
  }, [isAuthenticated, onboarded, navigate]);

  return <>{children}</>;
};
