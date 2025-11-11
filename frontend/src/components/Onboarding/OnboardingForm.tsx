import React, { useState } from 'react';
import { useForm, SubmitHandler, FieldPath } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';
import { FormData, SESSION_MIN_LENGTH } from './OnboardingFormTypes';
import { getStepFields, prepareApiPayload } from './OnboardingFormValidation';
import { OnboardingStep1BasicInfo } from './OnboardingStep1BasicInfo';
import { OnboardingStep2Goals } from './OnboardingStep2Goals';
import { OnboardingStep3Training } from './OnboardingStep3Training';
import { OnboardingStep4Details } from './OnboardingStep4Details';

export const OnboardingForm: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [step, setStep] = useState(1);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { control, handleSubmit, watch, trigger, formState: { errors }, getValues } = useForm<FormData>({
    defaultValues: {
      weight: { value: 0, unit: 'kg' },
      age: 0,
      training_days_per_week: 4,
      weight_plan: 'maintain',
      deadlift_style: 'conventional',
      squat_stance: 'medium',
      volume_preference: 'medium',
      feed_visibility: 'public',
      height: { value: 0, unit: 'cm', feet: 0, inches: 0 },
      squat_goal: { value: 0, unit: 'kg' },
      bench_goal: { value: 0, unit: 'kg' },
      dead_goal: { value: 0, unit: 'kg' },
      session_length_minutes: SESSION_MIN_LENGTH,
      injuries: '',
      knee_sleeve: '',
      has_competed: false,
      best_squat_kg: 0,
      best_bench_kg: 0,
      best_dead_kg: 0,
      best_total_kg: 0,
      comp_pr_date: '',
      comp_federation: '',
      squat_bar_position: 'medium',
    },
  });

  const onSubmit: SubmitHandler<FormData> = async (data) => {
    if (!user) {
      setError('User not authenticated');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    const apiPayload = prepareApiPayload(data);

    try {
      await apiClient.submitOnboardingSettings(user.id, apiPayload);
      navigate('/chat');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save settings. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const hasCompeted = watch('has_competed');
  const heightUnit = watch('height.unit');
  const feedVisibility = watch('feed_visibility');

  const nextStep = async () => {
    const currentStepFields = getStepFields(step, hasCompeted, heightUnit || 'cm', feedVisibility || 'public');

    const fieldsToValidate: FieldPath<FormData>[] = currentStepFields.filter(field => {
      if (!hasCompeted && (field === 'best_total_kg' || field === 'comp_pr_date' || field === 'comp_federation')) {
        return false;
      }
      return true;
    });

    const isValid = await trigger(fieldsToValidate);

    if (isValid) {
      setStep((prev) => Math.min(prev + 1, 4));
    }
  };

  const prevStep = () => setStep((prev) => Math.max(prev - 1, 1));

  const stepProps = { control, errors, watch, getValues };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg p-8">
        <h2 className="text-2xl font-bold mb-2">Welcome! Let's Get Started</h2>
        <p className="text-gray-600 mb-6">Help us customize your training experience</p>

        <div className="mb-8">
          <div className="flex justify-between items-center px-4">
            <div className="flex items-center">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 ${step >= 1 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-600'}`}>
                1
              </div>
              <span className="ml-2 text-sm text-gray-600">Basics</span>
            </div>

            <div className="flex items-center">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 ${step >= 2 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-600'}`}>
                2
              </div>
              <span className="ml-2 text-sm text-gray-600">Goals</span>
            </div>

            <div className="flex items-center">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 ${step >= 3 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-600'}`}>
                3
              </div>
              <span className="ml-2 text-sm text-gray-600">Training</span>
            </div>

            <div className="flex items-center">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 ${step >= 4 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-600'}`}>
                4
              </div>
              <span className="ml-2 text-sm text-gray-600">Details</span>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit(onSubmit)}>
          {step === 1 && <OnboardingStep1BasicInfo {...stepProps} />}
          {step === 2 && <OnboardingStep2Goals {...stepProps} />}
          {step === 3 && <OnboardingStep3Training {...stepProps} />}
          {step === 4 && <OnboardingStep4Details {...stepProps} />}

          {error && (
            <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <div className="mt-10 flex flex-col sm:flex-row justify-between gap-4">
            {step > 1 && (
              <button
                type="button"
                onClick={prevStep}
                className="px-6 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
              >
                Back
              </button>
            )}

            <div className="flex gap-4 ml-auto">
              {step < 4 ? (
                <button
                  type="button"
                  onClick={nextStep}
                  className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  Next
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  {isSubmitting ? 'Saving...' : 'Complete Setup'}
                </button>
              )}
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};
