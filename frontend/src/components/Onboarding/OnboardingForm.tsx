import React, { useState } from 'react';
import { useForm, Controller, SubmitHandler } from 'react-hook-form';
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

  const { control, handleSubmit, watch, trigger, formState: { errors } } = useForm<FormData>({
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
      session_length_minutes: 0,
    },
  });

  const onSubmit: SubmitHandler<FormData> = async (data) => {
    console.log("submitting...")
    if (!user) {
      setError('User not authenticated');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    const apiPayload = JSON.parse(JSON.stringify(data));

    if (apiPayload.height && apiPayload.height.unit === 'in') {
      const feet = apiPayload.height.feet || 0;
      const inches = apiPayload.height.inches || 0;
      apiPayload.height.value = Math.round((feet * 12 + inches) * 2.54);
      apiPayload.height.unit = 'cm'; 
    }

    if (apiPayload.height) {
      delete apiPayload.height.feet;
      delete apiPayload.height.inches;
    }

    if (apiPayload.squat_goal) apiPayload.squat_goal.unit = 'kg';
    if (apiPayload.bench_goal) apiPayload.bench_goal.unit = 'kg';
    if (apiPayload.dead_goal) apiPayload.dead_goal.unit = 'kg';

    try {
      await apiClient.submitOnboardingSettings(user.id, apiPayload);
      navigate('/feed');
    } catch (err: any) {
      console.error('Failed to submit onboarding settings:', err);
      setError(err.response?.data?.error || 'Failed to save settings. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const getStepFields = (step: number): (keyof FormData)[] => {
    switch (step) {
      case 1:
        return ['age', 'weight.value', 'weight.unit', 'height.feet', 'height.inches', 'height.value'];
      case 2:
        return ['target_weight_class', 'competition_date', 'squat_goal.value', 'bench_goal.value', 'dead_goal.value', 'most_important_lift', 'least_important_lift'];
      case 3:
        return ['training_days_per_week', 'session_length_minutes', 'volume_preference', 'recovery_rating_squat', 'recovery_rating_bench', 'recovery_rating_dead', 'squat_stance', 'deadlift_style'];
      case 4:
        const step4Fields: (keyof FormData)[] = ['weight_plan', 'feed_visibility'];
        if (watch('feed_visibility') === 'passcode') {
            step4Fields.push('passcode');
        }
        return step4Fields;
      default:
        return [];
    }
  };

  const nextStep = async () => {
    const currentStepFields = getStepFields(step);
    let isValid = true;

    for (const field of currentStepFields) {
      const fieldValid = await trigger(field);
      if (!fieldValid) {
        isValid = false;
        break;
      }
    }

    if (isValid) {
      setStep((prev) => Math.min(prev + 1, 4));
    }
  };
  const prevStep = () => setStep((prev) => Math.max(prev - 1, 1));

  const heightUnit = watch('height.unit');
  const watchedFeet = watch('height.feet') as number | undefined;
  const watchedInches = watch('height.inches') as number | undefined;
  const watchedHeightValue = watch('height.value') as number | undefined;

  const watchedSquatGoal = watch('squat_goal.value') as number | undefined;
  const watchedBenchGoal = watch('bench_goal.value') as number | undefined;
  const watchedDeadGoal = watch('dead_goal.value') as number | undefined;


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
          {step === 1 && (
            <div className="space-y-6">
              <h3 className="text-xl font-semibold mb-6">Basic Information</h3>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Age *
                  </label>
                  <Controller
                    name="age"
                    control={control}
                    rules={{
                      required: "Age is required",
                      min: { value: 8, message: "Must be at least 8 years old" },
                      max: { value: 120, message: "Must be no more than 120 years old" },
                      validate: (value) => value > 0 || "Age must be greater than 0"
                    }}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="number"
                        min="8"
                        max="120"
                        className={`w-full px-3 py-2 border rounded-md ${
                          errors.age ? 'border-red-300' : 'border-gray-300'
                        }`}
                        onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                      />
                    )}
                  />
                  {errors.age && (
                    <span className="text-sm text-red-600">{errors.age.message}</span>
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
                      rules={{
                        required: "Weight is required",
                        min: { value: 1, message: "Weight must be at least 1" },
                        validate: (value) => value > 0 || "Weight must be greater than 0"
                      }}
                      render={({ field }) => (
                        <input
                          {...field}
                          type="number"
                          step="0.1"
                          min="1"
                          className={`flex-1 px-3 py-2 border rounded-md ${
                            errors.weight?.value ? 'border-red-300' : 'border-gray-300'
                          }`}
                          onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
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
                  {errors.weight?.value && (
                    <span className="text-sm text-red-600">{errors.weight.value.message}</span>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Height
                  </label>
                  <Controller
                    name="height.unit"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        className="px-3 py-2 border border-gray-300 rounded-md mb-2"
                      >
                        <option value="cm">cm</option>
                        <option value="in">ft/in</option>
                      </select>
                    )}
                  />
                  {heightUnit === 'in' ? (
                    <div className="flex items-center gap-2">
                      <Controller
                        name="height.feet"
                        control={control}
                        rules={{
                          min: { value: 0, message: "Feet cannot be negative" },
                          max: { value: 10, message: "Feet must be reasonable (max 10')" },
                          validate: (value) => value !== undefined || "Feet is required"
                        }}
                        render={({ field }) => (
                          <input
                            {...field}
                            type="number"
                            min="0"
                            max="10"
                            placeholder="Feet"
                            className={`w-20 px-3 py-2 border rounded-md ${
                              errors.height?.feet ? 'border-red-300' : 'border-gray-300'
                            }`}
                            onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                          />
                        )}
                      />
                      <span className="text-gray-500">'</span>
                      <Controller
                        name="height.inches"
                        control={control}
                        rules={{
                          min: { value: 0, message: "Inches cannot be negative" },
                          max: { value: 11, message: "Inches must be less than 12" },
                          validate: (value) => value !== undefined || "Inches is required"
                        }}
                        render={({ field }) => (
                          <input
                            {...field}
                            type="number"
                            min="0"
                            max="11"
                            placeholder="Inches"
                            className={`w-20 px-3 py-2 border rounded-md ${
                              errors.height?.inches ? 'border-red-300' : 'border-gray-300'
                            }`}
                            onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                          />
                        )}
                      />
                      <span className="text-gray-500">"</span>
                    </div>
                  ) : (
                    <Controller
                      name="height.value"
                      control={control}
                      rules={{
                        min: { value: 0, message: "Height cannot be negative" },
                        validate: (value) => !value || value >= 0 || "Height cannot be negative"
                      }}
                      render={({ field }) => (
                        <input
                          {...field}
                          type="number"
                          step="0.1"
                          min="0"
                          className={`w-full px-3 py-2 border rounded-md ${
                            errors.height?.value ? 'border-red-300' : 'border-gray-300'
                          }`}
                          onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                          placeholder="e.g., 175"
                        />
                      )}
                    />
                  )}
                  {errors.height?.feet && (
                    <span className="text-sm text-red-600">{errors.height.feet.message}</span>
                  )}
                  {errors.height?.inches && (
                    <span className="text-sm text-red-600">{errors.height.inches.message}</span>
                  )}
                  {errors.height?.value && (
                    <span className="text-sm text-red-600">{errors.height.value.message}</span>
                  )}
                  {(() => {
                    if (heightUnit === 'in') {
                      const feet = watchedFeet || 0;
                      const inches = watchedInches || 0;
                      const totalInches = feet * 12 + inches;
                      return totalInches > 0 ? (
                        <p className="text-xs text-gray-500 mt-1">
                          {feet}'{inches}" = {Math.round(totalInches * 2.54)} cm
                        </p>
                      ) : null;
                    } else {
                      const heightValue = watchedHeightValue || 0;
                      const totalInches = heightValue / 2.54;
                      const feet = Math.floor(totalInches / 12);
                      const inches = Math.round(totalInches % 12);
                      return heightValue > 0 ? (
                        <p className="text-xs text-gray-500 mt-1">
                          {heightValue} cm = {feet}'{inches}"
                        </p>
                      ) : null;
                    }
                  })()}
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
            <div className="space-y-6">
              <h3 className="text-xl font-semibold mb-6">Goals & Competition</h3>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Target Weight Class *
                  </label>
                  <Controller
                    name="target_weight_class"
                    control={control}
                    rules={{ required: "Target weight class is required" }}
                    render={({ field }) => (
                      <select
                        {...field}
                        className={`w-full px-3 py-2 border rounded-md ${
                          errors.target_weight_class ? 'border-red-300' : 'border-gray-300'
                        }`}
                      >
                        <option value="">Select Weight Class</option>
                        <optgroup label="Men - kg">
                          <option value="59kg">59 kg</option>
                          <option value="66kg">66 kg</option>
                          <option value="74kg">74 kg</option>
                          <option value="83kg">83 kg</option>
                          <option value="93kg">93 kg</option>
                          <option value="105kg">105 kg</option>
                          <option value="120kg">120 kg</option>
                          <option value="120+kg">120+ kg</option>
                        </optgroup>
                        <optgroup label="Women - kg">
                          <option value="47kg">47 kg</option>
                          <option value="52kg">52 kg</option>
                          <option value="57kg">57 kg</option>
                          <option value="63kg">63 kg</option>
                          <option value="72kg">72 kg</option>
                          <option value="84kg">84 kg</option>
                          <option value="84+kg">84+ kg</option>
                        </optgroup>
                      </select>
                    )}
                  />
                  {errors.target_weight_class && (
                    <span className="text-sm text-red-600">{errors.target_weight_class.message}</span>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Competition Date *
                  </label>
                  <Controller
                    name="competition_date"
                    control={control}
                    rules={{ required: "Competition date is required" }}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="date"
                        className={`w-full px-3 py-2 border rounded-md ${
                          errors.competition_date ? 'border-red-300' : 'border-gray-300'
                        }`}
                      />
                    )}
                  />
                  {errors.competition_date && (
                    <span className="text-sm text-red-600">{errors.competition_date.message}</span>
                  )}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Squat Goal (kg)
                </label>
                <Controller
                  name="squat_goal.value"
                  control={control}
                  rules={{
                    min: { value: 0, message: "Squat goal cannot be negative" },
                    validate: (value) => !value || value >= 0 || "Squat goal cannot be negative"
                  }}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      step="0.5"
                      min="0"
                      className={`w-full px-3 py-2 border rounded-md ${
                        errors.squat_goal?.value ? 'border-red-300' : 'border-gray-300'
                      }`}
                      onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                    />
                  )}
                />
                {errors.squat_goal?.value && (
                  <span className="text-sm text-red-600">{errors.squat_goal.value.message}</span>
                )}
                {(() => {
                    const squatGoalValue = watchedSquatGoal || 0;
                    return squatGoalValue > 0 ? (
                      <p className="text-xs text-gray-500 mt-1">
                        {squatGoalValue} kg = {Math.round(squatGoalValue * 2.20462)} lbs (helper conversion)
                      </p>
                    ) : null;
                  })()}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Bench Goal (kg)
                </label>
                <Controller
                  name="bench_goal.value"
                  control={control}
                  rules={{
                    min: { value: 0, message: "Bench goal cannot be negative" },
                    validate: (value) => !value || value >= 0 || "Bench goal cannot be negative"
                  }}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      step="0.5"
                      min="0"
                      className={`w-full px-3 py-2 border rounded-md ${
                        errors.bench_goal?.value ? 'border-red-300' : 'border-gray-300'
                      }`}
                      onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                    />
                  )}
                />
                {errors.bench_goal?.value && (
                  <span className="text-sm text-red-600">{errors.bench_goal.value.message}</span>
                )}
                {(() => {
                    const benchGoalValue = watchedBenchGoal || 0;
                    return benchGoalValue > 0 ? (
                      <p className="text-xs text-gray-500 mt-1">
                        {benchGoalValue} kg = {Math.round(benchGoalValue * 2.20462)} lbs (helper conversion)
                      </p>
                    ) : null;
                  })()}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Deadlift Goal (kg)
                </label>
                <Controller
                  name="dead_goal.value"
                  control={control}
                  rules={{
                    min: { value: 0, message: "Deadlift goal cannot be negative" },
                    validate: (value) => !value || value >= 0 || "Deadlift goal cannot be negative"
                  }}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      step="0.5"
                      min="0"
                      className={`w-full px-3 py-2 border rounded-md ${
                        errors.dead_goal?.value ? 'border-red-300' : 'border-gray-300'
                      }`}
                      onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                    />
                  )}
                />
                {errors.dead_goal?.value && (
                  <span className="text-sm text-red-600">{errors.dead_goal.value.message}</span>
                )}
                {(() => {
                    const deadGoalValue = watchedDeadGoal || 0;
                    return deadGoalValue > 0 ? (
                      <p className="text-xs text-gray-500 mt-1">
                        {deadGoalValue} kg = {Math.round(deadGoalValue * 2.20462)} lbs (helper conversion)
                      </p>
                    ) : null;
                  })()}
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Most Important Lift *
                  </label>
                  <Controller
                    name="most_important_lift"
                    control={control}
                    rules={{ required: "Most important lift is required" }}
                    render={({ field }) => (
                      <select
                        {...field}
                        className={`w-full px-3 py-2 border rounded-md ${
                          errors.most_important_lift ? 'border-red-300' : 'border-gray-300'
                        }`}
                      >
                        <option value="">Select...</option>
                        <option value="squat">Squat</option>
                        <option value="bench">Bench</option>
                        <option value="deadlift">Deadlift</option>
                      </select>
                    )}
                  />
                  {errors.most_important_lift && (
                    <span className="text-sm text-red-600">{errors.most_important_lift.message}</span>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Least Important Lift *
                  </label>
                  <Controller
                    name="least_important_lift"
                    control={control}
                    rules={{
                      required: "Least important lift is required",
                      validate: (value) =>
                        watch('most_important_lift') !== value ||
                        "Most and least important lifts cannot be the same"
                    }}
                    render={({ field }) => (
                      <select
                        {...field}
                        className={`w-full px-3 py-2 border rounded-md ${
                          errors.least_important_lift
                            ? 'border-red-300'
                            : 'border-gray-300'
                        }`}
                      >
                        <option value="">Select...</option>
                        <option value="squat">Squat</option>
                        <option value="bench">Bench</option>
                        <option value="deadlift">Deadlift</option>
                      </select>
                    )}
                  />
                   {errors.least_important_lift && (
                    <span className="text-sm text-red-600">{errors.least_important_lift.message}</span>
                  )}
                </div>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-6">
              <h3 className="text-xl font-semibold mb-6">Training Preferences</h3>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Training Days Per Week *
                </label>
                <Controller
                  name="training_days_per_week"
                  control={control}
                  rules={{
                    required: "Training days per week is required",
                    min: { value: 1, message: "Training days must be at least 1" },
                    max: { value: 7, message: "Training days must be no more than 7" },
                    validate: (value) => value > 0 || "Training days must be greater than 0"
                  }}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      min="1"
                      max="7"
                      className={`w-full px-3 py-2 border rounded-md ${
                        errors.training_days_per_week
                          ? 'border-red-300'
                          : 'border-gray-300'
                      }`}
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                    />
                  )}
                />
                {errors.training_days_per_week && (
                  <span className="text-sm text-red-600">
                    {errors.training_days_per_week.message}
                  </span>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Session Length (minutes)
                </label>
                <Controller
                  name="session_length_minutes"
                  control={control}
                  rules={{
                    min: { value: 15, message: "Session length must be at least 15 minutes" },
                    max: { value: 360, message: "Session length must be no more than 360 minutes (6 hours)" },
                    validate: (value) =>
                      !value || (value >= 15 && value <= 360) ||
                      "Session length must be between 15 and 360 minutes (6 hours)"
                  }}
                  render={({ field }) => (
                    <input
                      {...field}
                      type="number"
                      min="15"
                      max="360"
                      className={`w-full px-3 py-2 border rounded-md ${
                        errors.session_length_minutes
                          ? 'border-red-300'
                          : 'border-gray-300'
                      }`}
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                      placeholder="15-360 minutes"
                    />
                  )}
                />
                {errors.session_length_minutes && (
                  <span className="text-sm text-red-600">
                    {errors.session_length_minutes.message}
                  </span>
                )}
                {(() => {
                    const sessionLengthMinutes = watch('session_length_minutes') || 0;
                    return sessionLengthMinutes > 0 ? (
                      <p className="text-xs text-gray-500 mt-1">
                        {Math.floor(sessionLengthMinutes / 60)}h {sessionLengthMinutes % 60}m
                      </p>
                    ) : null;
                  })()}
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

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Recovery Ratings
                </label>
                <p className="text-xs text-gray-500 mb-4">
                  How well do you recover from heavy lifting sessions?
                </p>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="text-center">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Squat Recovery
                    </label>
                    <Controller
                      name="recovery_rating_squat"
                      control={control}
                      render={({ field }) => (
                        <div className="flex justify-center space-x-4">
                          <button
                            type="button"
                            onClick={() => field.onChange(1)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_squat') === 1
                                ? 'bg-red-100 border-2 border-red-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Sad - Need more than a week to recover"
                          >
                            😞
                          </button>
                          <button
                            type="button"
                            onClick={() => field.onChange(2)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_squat') === 2
                                ? 'bg-yellow-100 border-2 border-yellow-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Neutral - Can handle 1-2 heavy sessions per week"
                          >
                            😐
                          </button>
                          <button
                            type="button"
                            onClick={() => field.onChange(3)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_squat') === 3
                                ? 'bg-green-100 border-2 border-green-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Happy - Can handle more than 2 heavy sessions per week"
                          >
                            😊
                          </button>
                        </div>
                      )}
                    />
                    <p className="text-xs text-gray-500 mt-2">
                      {watch('recovery_rating_squat') === 1 && "😞 Need more than a week to recover"}
                      {watch('recovery_rating_squat') === 2 && "😐 Can handle 1-2 heavy sessions per week"}
                      {watch('recovery_rating_squat') === 3 && "😊 Can handle 2+ heavy sessions per week"}
                    </p>
                  </div>

                  <div className="text-center">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Bench Recovery
                    </label>
                    <Controller
                      name="recovery_rating_bench"
                      control={control}
                      render={({ field }) => (
                        <div className="flex justify-center space-x-4">
                          <button
                            type="button"
                            onClick={() => field.onChange(1)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_bench') === 1
                                ? 'bg-red-100 border-2 border-red-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Sad - Need more than a week to recover"
                          >
                            😞
                          </button>
                          <button
                            type="button"
                            onClick={() => field.onChange(2)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_bench') === 2
                                ? 'bg-yellow-100 border-2 border-yellow-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Neutral - Can handle 1-2 heavy sessions per week"
                          >
                            😐
                          </button>
                          <button
                            type="button"
                            onClick={() => field.onChange(3)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_bench') === 3
                                ? 'bg-green-100 border-2 border-green-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Happy - Can handle more than 2 heavy sessions per week"
                          >
                            😊
                          </button>
                        </div>
                      )}
                    />
                    <p className="text-xs text-gray-500 mt-2">
                      {watch('recovery_rating_bench') === 1 && "😞 Need more than a week to recover"}
                      {watch('recovery_rating_bench') === 2 && "😐 Can handle 1-2 heavy sessions per week"}
                      {watch('recovery_rating_bench') === 3 && "😊 Can handle 2+ heavy sessions per week"}
                    </p>
                  </div>

                  <div className="text-center">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Deadlift Recovery
                    </label>
                    <Controller
                      name="recovery_rating_dead"
                      control={control}
                      render={({ field }) => (
                        <div className="flex justify-center space-x-4">
                          <button
                            type="button"
                            onClick={() => field.onChange(1)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_dead') === 1
                                ? 'bg-red-100 border-2 border-red-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Sad - Need more than a week to recover"
                          >
                            😞
                          </button>
                          <button
                            type="button"
                            onClick={() => field.onChange(2)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_dead') === 2
                                ? 'bg-yellow-100 border-2 border-yellow-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Neutral - Can handle 1-2 heavy sessions per week"
                          >
                            😐
                          </button>
                          <button
                            type="button"
                            onClick={() => field.onChange(3)}
                            className={`text-2xl p-2 rounded-full ${
                              watch('recovery_rating_dead') === 3
                                ? 'bg-green-100 border-2 border-green-500'
                                : 'border-2 border-gray-200'
                            }`}
                            title="Happy - Can handle more than 2 heavy sessions per week"
                          >
                            😊
                          </button>
                        </div>
                      )}
                    />
                    <p className="text-xs text-gray-500 mt-2">
                      {watch('recovery_rating_dead') === 1 && "😞 Need more than a week to recover"}
                      {watch('recovery_rating_dead') === 2 && "😐 Can handle 1-2 heavy sessions per week"}
                      {watch('recovery_rating_dead') === 3 && "😊 Can handle 2+ heavy sessions per week"}
                    </p>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
            <div className="space-y-6">
              <h3 className="text-xl font-semibold mb-6">Additional Details</h3>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Weight Plan *
                </label>
                <Controller
                  name="weight_plan"
                  control={control}
                  rules={{ required: "Weight plan is required" }}
                  render={({ field }) => (
                    <select
                      {...field}
                      className={`w-full px-3 py-2 border rounded-md ${
                        errors.weight_plan ? 'border-red-300' : 'border-gray-300'
                      }`}
                    >
                      <option value="maintain">Maintain</option>
                      <option value="gain">Gain</option>
                      <option value="lose">Lose</option>
                    </select>
                  )}
                />
                {errors.weight_plan && (
                  <span className="text-sm text-red-600">{errors.weight_plan.message}</span>
                )}
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
                  Feed Visibility *
                </label>
                <Controller
                  name="feed_visibility"
                  control={control}
                  rules={{ required: "Feed visibility is required" }}
                  render={({ field }) => (
                    <select
                      {...field}
                      className={`w-full px-3 py-2 border rounded-md ${
                        errors.feed_visibility ? 'border-red-300' : 'border-gray-300'
                      }`}
                    >
                      <option value="public">Public</option>
                      <option value="passcode">Passcode Protected</option>
                    </select>
                  )}
                />
                {errors.feed_visibility && (
                  <span className="text-sm text-red-600">{errors.feed_visibility.message}</span>
                )}
              </div>

              {watch('feed_visibility') === 'passcode' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Feed Passcode *
                  </label>
                  <Controller
                    name="passcode"
                    control={control}
                    rules={{
                      required: "Passcode is required for protected feed",
                      minLength: { value: 4, message: "Passcode must be at least 4 characters" }
                    }}
                    render={({ field }) => (
                      <input
                        {...field}
                        type="password"
                        minLength={4}
                        placeholder="Minimum 4 characters"
                        className={`w-full px-3 py-2 border rounded-md ${
                          errors.passcode ? 'border-red-300' : 'border-gray-300'
                        }`}
                      />
                    )}
                  />
                  {errors.passcode && (
                    <span className="text-sm text-red-600">
                      {errors.passcode.message}
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