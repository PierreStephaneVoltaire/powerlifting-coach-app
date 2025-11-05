import React, { useState } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';
import { OnboardingSettings } from '@/types';

interface FormData extends OnboardingSettings {}

export const OnboardingForm: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [step, setStep] = useState(1);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { control, handleSubmit, watch, formState: { errors } } = useForm<FormData>({
    defaultValues: {
      weight: { value: 0, unit: 'kg' },
      age: 0,
      training_days_per_week: 4,
      weight_plan: 'maintain',
      deadlift_style: 'conventional',
      squat_stance: 'medium',
      volume_preference: 'medium',
      feed_visibility: 'public',
    },
  });

  const weightUnit = watch('weight.unit');

  const onSubmit = async (data: FormData) => {
    if (!user) {
      setError('User not authenticated');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      await apiClient.submitOnboardingSettings(user.id, data);
      navigate('/feed');
    } catch (err: any) {
      console.error('Failed to submit onboarding settings:', err);
      setError(err.response?.data?.error || 'Failed to save settings. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const nextStep = () => setStep((prev) => Math.min(prev + 1, 4));
  const prevStep = () => setStep((prev) => Math.max(prev - 1, 1));

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg p-8">
        <h2 className="text-2xl font-bold mb-2">Welcome! Let's Get Started</h2>
        <p className="text-gray-600 mb-6">Help us customize your training experience</p>

        <div className="mb-8">
          <div className="flex justify-between items-center">
            {[1, 2, 3, 4].map((s) => (
              <div key={s} className="flex items-center">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center ${
                    step >= s ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-600'
                  }`}
                >
                  {s}
                </div>
                {s < 4 && (
                  <div
                    className={`h-1 w-16 ${
                      step > s ? 'bg-blue-600' : 'bg-gray-200'
                    }`}
                  />
                )}
              </div>
            ))}
          </div>
          <div className="flex justify-between mt-2">
            <span className="text-xs text-gray-600">Basics</span>
            <span className="text-xs text-gray-600">Goals</span>
            <span className="text-xs text-gray-600">Training</span>
            <span className="text-xs text-gray-600">Details</span>
          </div>
        </div>

        <form onSubmit={handleSubmit(onSubmit)}>
          {step === 1 && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold mb-4">Basic Information</h3>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Age *
                  </label>
                  <Controller
                    name="age"
                    control={control}
                    rules={{ required: true, min: 13, max: 120 }}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="number"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        onChange={(e) => field.onChange(parseInt(e.target.value))}
                      />
                    )}
                  />
                  {errors.age && (
                    <span className="text-sm text-red-600">Age must be between 13 and 120</span>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Weight *
                  </label>
                  <div className="flex gap-2">
                    <Controller
                      name="weight.value"
                      control={control}
                      rules={{ required: true, min: 1 }}
                      render={({ field }) => (
                        <input
                          {...field}
                          type="number"
                          step="0.1"
                          className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
                          onChange={(e) => field.onChange(parseFloat(e.target.value))}
                        />
                      )}
                    />
                    <Controller
                      name="weight.unit"
                      control={control}
                      render={({ field }) => (
                        <select
                          {...field}
                          className="px-3 py-2 border border-gray-300 rounded-md"
                        >
                          <option value="kg">kg</option>
                          <option value="lb">lb</option>
                        </select>
                      )}
                    />
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Height
                  </label>
                  <div className="flex gap-2">
                    <Controller
                      name="height.value"
                      control={control}
                      render={({ field }) => (
                        <input
                          {...field}
                          type="number"
                          step="0.1"
                          className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
                          onChange={(e) => field.onChange(parseFloat(e.target.value))}
                        />
                      )}
                    />
                    <Controller
                      name="height.unit"
                      control={control}
                      render={({ field }) => (
                        <select
                          {...field}
                          className="px-3 py-2 border border-gray-300 rounded-md"
                        >
                          <option value="cm">cm</option>
                          <option value="in">in</option>
                        </select>
                      )}
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Federation
                  </label>
                  <Controller
                    name="federation"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="text"
                        placeholder="e.g., IPF, USPA"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      />
                    )}
                  />
                </div>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold mb-4">Goals & Competition</h3>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Target Weight Class
                  </label>
                  <Controller
                    name="target_weight_class"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="text"
                        placeholder="e.g., 93kg, 183lb"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      />
                    )}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Weeks Until Competition
                  </label>
                  <Controller
                    name="weeks_until_comp"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="number"
                        min="0"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        onChange={(e) => field.onChange(parseInt(e.target.value))}
                      />
                    )}
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Squat Goal ({weightUnit})
                </label>
                <Controller
                  name="squat_goal.value"
                  control={control}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      step="0.5"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      onChange={(e) => field.onChange(parseFloat(e.target.value))}
                    />
                  )}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Bench Goal ({weightUnit})
                </label>
                <Controller
                  name="bench_goal.value"
                  control={control}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      step="0.5"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      onChange={(e) => field.onChange(parseFloat(e.target.value))}
                    />
                  )}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Deadlift Goal ({weightUnit})
                </label>
                <Controller
                  name="dead_goal.value"
                  control={control}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      step="0.5"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      onChange={(e) => field.onChange(parseFloat(e.target.value))}
                    />
                  )}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Most Important Lift
                  </label>
                  <Controller
                    name="most_important_lift"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      >
                        <option value="">Select...</option>
                        <option value="squat">Squat</option>
                        <option value="bench">Bench</option>
                        <option value="deadlift">Deadlift</option>
                      </select>
                    )}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Least Important Lift
                  </label>
                  <Controller
                    name="least_important_lift"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      >
                        <option value="">Select...</option>
                        <option value="squat">Squat</option>
                        <option value="bench">Bench</option>
                        <option value="deadlift">Deadlift</option>
                      </select>
                    )}
                  />
                </div>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold mb-4">Training Preferences</h3>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Training Days Per Week *
                </label>
                <Controller
                  name="training_days_per_week"
                  control={control}
                  rules={{ required: true, min: 1, max: 7 }}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      min="1"
                      max="7"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      onChange={(e) => field.onChange(parseInt(e.target.value))}
                    />
                  )}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Session Length (minutes)
                </label>
                <Controller
                  name="session_length_minutes"
                  control={control}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      min="15"
                      max="300"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      onChange={(e) => field.onChange(parseInt(e.target.value))}
                    />
                  )}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Volume Preference
                </label>
                <Controller
                  name="volume_preference"
                  control={control}
                  render={({ field }) => (
                    <select
                      {...field}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    >
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                    </select>
                  )}
                />
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Squat Recovery (1-5)
                  </label>
                  <Controller
                    name="recovery_rating_squat"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="number"
                        min="1"
                        max="5"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        onChange={(e) => field.onChange(parseInt(e.target.value))}
                      />
                    )}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Bench Recovery (1-5)
                  </label>
                  <Controller
                    name="recovery_rating_bench"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="number"
                        min="1"
                        max="5"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        onChange={(e) => field.onChange(parseInt(e.target.value))}
                      />
                    )}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Deadlift Recovery (1-5)
                  </label>
                  <Controller
                    name="recovery_rating_dead"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="number"
                        min="1"
                        max="5"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        onChange={(e) => field.onChange(parseInt(e.target.value))}
                      />
                    )}
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Squat Stance
                  </label>
                  <Controller
                    name="squat_stance"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      >
                        <option value="narrow">Narrow</option>
                        <option value="medium">Medium</option>
                        <option value="wide">Wide</option>
                      </select>
                    )}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Deadlift Style
                  </label>
                  <Controller
                    name="deadlift_style"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      >
                        <option value="conventional">Conventional</option>
                        <option value="sumo">Sumo</option>
                      </select>
                    )}
                  />
                </div>
              </div>
            </div>
          )}

          {step === 4 && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold mb-4">Additional Details</h3>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Weight Plan
                </label>
                <Controller
                  name="weight_plan"
                  control={control}
                  render={({ field }) => (
                    <select
                      {...field}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    >
                      <option value="maintain">Maintain</option>
                      <option value="gain">Gain</option>
                      <option value="lose">Lose</option>
                    </select>
                  )}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Injuries or Limitations
                </label>
                <Controller
                  name="injuries"
                  control={control}
                  render={({ field }) => (
                    <textarea
                      {...field}
                      rows={3}
                      placeholder="Describe any injuries, pain, or limitations..."
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    />
                  )}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Feed Visibility
                </label>
                <Controller
                  name="feed_visibility"
                  control={control}
                  render={({ field }) => (
                    <select
                      {...field}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    >
                      <option value="public">Public</option>
                      <option value="passcode">Passcode Protected</option>
                    </select>
                  )}
                />
              </div>

              {watch('feed_visibility') === 'passcode' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Feed Passcode *
                  </label>
                  <Controller
                    name="passcode"
                    control={control}
                    rules={{ minLength: 4 }}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="password"
                        minLength={4}
                        placeholder="Minimum 4 characters"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md"
                      />
                    )}
                  />
                  {errors.passcode && (
                    <span className="text-sm text-red-600">
                      Passcode must be at least 4 characters
                    </span>
                  )}
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Knee Sleeves Brand
                </label>
                <Controller
                  name="knee_sleeve"
                  control={control}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="text"
                      placeholder="e.g., SBD, Inzer"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    />
                  )}
                />
              </div>
            </div>
          )}

          {error && (
            <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <div className="mt-8 flex justify-between">
            {step > 1 && (
              <button
                type="button"
                onClick={prevStep}
                className="px-6 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
              >
                Back
              </button>
            )}

            {step < 4 ? (
              <button
                type="button"
                onClick={nextStep}
                className="ml-auto px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Next
              </button>
            ) : (
              <button
                type="submit"
                disabled={isSubmitting}
                className="ml-auto px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {isSubmitting ? 'Saving...' : 'Complete Setup'}
              </button>
            )}
          </div>
        </form>
      </div>
    </div>
  );
};
