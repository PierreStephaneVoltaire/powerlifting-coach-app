import React from 'react';
import { Controller } from 'react-hook-form';
import { StepProps } from './OnboardingFormTypes';
import { getWeightClassKg, weightPlanValidation } from './OnboardingFormValidation';

export const OnboardingStep4Details: React.FC<StepProps> = ({
  control,
  errors,
  watch,
}) => {
  const currentWeightKg = watch('weight.unit') === 'lb'
    ? (watch('weight.value') || 0) / 2.20462
    : (watch('weight.value') || 0);

  const targetWeightClass = watch('target_weight_class');
  const targetKg = getWeightClassKg(targetWeightClass);

  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold mb-6">Additional Details</h3>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Weight Plan *
        </label>
        <Controller
          name="weight_plan"
          control={control}
          rules={{
            required: "Weight plan is required",
            validate: (value) => weightPlanValidation(value, currentWeightKg, targetKg)
          }}
          render={({ field }) => (
            <select
              {...field}
              className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
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
        <p className="text-xs text-gray-500 mt-1">
          Current weight: {currentWeightKg.toFixed(1)} kg. Target class: {targetKg} kg.
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Injuries or Limitations *
        </label>
        <Controller
          name="injuries"
          control={control}
          rules={{ required: "This field is required" }}
          render={({ field }) => (
            <textarea
              {...field}
              rows={3}
              placeholder="Describe any injuries, pain, or limitations..."
              className={`w-full px-3 py-2 border rounded-md ${
                errors.injuries ? 'border-red-300' : 'border-gray-300'
              }`}
            />
          )}
        />
        {errors.injuries && (
          <span className="text-sm text-red-600">{errors.injuries.message}</span>
        )}
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
              className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
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
          Knee Sleeves Brand *
        </label>
        <Controller
          name="knee_sleeve"
          control={control}
          rules={{ required: "Knee sleeve brand is required" }}
          render={({ field }) => (
            <input
              {...field}
              type="text"
              placeholder="e.g., SBD, Inzer"
              className={`w-full px-3 py-2 border rounded-md ${
                errors.knee_sleeve ? 'border-red-300' : 'border-gray-300'
              }`}
            />
          )}
        />
        {errors.knee_sleeve && (
          <span className="text-sm text-red-600">{errors.knee_sleeve.message}</span>
        )}
      </div>
    </div>
  );
};
