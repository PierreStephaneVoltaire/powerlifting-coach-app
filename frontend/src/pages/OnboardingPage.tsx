import React from 'react';
import { OnboardingForm } from '@/components/Onboarding/OnboardingForm';

export const OnboardingPage: React.FC = () => {
  return (
    <div className="min-h-screen bg-gray-100 py-8">
      <div className="max-w-4xl mx-auto px-4">
        <h1 className="text-3xl font-bold text-center mb-8">Welcome! Let's Set Up Your Profile</h1>
        <OnboardingForm />
      </div>
    </div>
  );
};
