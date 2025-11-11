import React from 'react';
import { Controller } from 'react-hook-form';
import {
  StepProps,
  MAX_WEIGHT_KG,
  MAX_SQUAT_KG,
  MAX_BENCH_KG,
  MAX_DEADLIFT_KG,
  MAX_TOTAL_KG,
  MIN_LIFT_KG,
  MAX_HEIGHT_CM,
  MIN_HEIGHT_CM,
  CANADIAN_FEDS,
} from './OnboardingFormTypes';

export const OnboardingStep1BasicInfo: React.FC<StepProps> = ({
  control,
  errors,
  watch,
  getValues,
}) => {
  const hasCompeted = watch('has_competed');
  const heightUnit = watch('height.unit');
  const watchedFeet = watch('height.feet') as number | undefined;
  const watchedInches = watch('height.inches') as number | undefined;
  const watchedHeightValue = watch('height.value') as number | undefined;

  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold mb-6">Basic Information & Current Lifts</h3>

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
              validate: (value) => (value as number) > 0 || "Age must be greater than 0"
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
            Current Weight *
          </label>
          <div className="flex gap-2">
            <Controller
              name="weight.value"
              control={control}
              rules={{
                required: "Weight is required",
                min: { value: 1, message: "Weight must be at least 1" },
                max: { value: MAX_WEIGHT_KG * 2.20462 + 1, message: `Max weight is ${MAX_WEIGHT_KG} kg / ${(MAX_WEIGHT_KG * 2.20462).toFixed(0)} lb` },
                validate: (value) => {
                  const val = value as number | undefined;
                  if (val === undefined || val === null) return "Weight is required";
                  if (watch('weight.unit') === 'kg' && val > MAX_WEIGHT_KG) {
                    return `Max weight is ${MAX_WEIGHT_KG} kg`;
                  }
                  return val > 0 || "Weight must be greater than 0"
                }
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
              rules={{ required: "Unit is required" }}
              render={({ field }) => (
                <select
                  {...field}
                  className="px-3 py-2 border border-gray-300 rounded-md appearance-none cursor-pointer"
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

      <div className="grid grid-cols-1 gap-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Height *
          </label>
          <Controller
            name="height.unit"
            control={control}
            rules={{ required: "Height unit is required" }}
            render={({ field }) => (
              <select
                {...field}
                className="px-3 py-2 border border-gray-300 rounded-md mb-2 appearance-none cursor-pointer"
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
                  required: "Feet is required",
                  min: { value: 3, message: "Min height is 3'0\"" },
                  max: { value: 8, message: "Max height is 8'11\"" },
                  validate: (value) => (value !== undefined && value !== null && (value as number) >= 3) || "Feet is required"
                }}
                render={({ field }) => (
                  <input
                    {...field}
                    type="number"
                    min="3"
                    max="8"
                    placeholder="Feet (3-8)"
                    className={`w-24 px-3 py-2 border rounded-md ${
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
                  required: "Inches is required",
                  min: { value: 0, message: "Inches cannot be negative" },
                  max: { value: 11, message: "Inches must be less than 12" },
                  validate: (value: number | undefined) => {
                    if (value === undefined || value === null) {
                      return "Inches is required";
                    }
                    const feet = getValues('height.feet') || 0;
                    if (feet === 8 && value > 11) {
                      return "Max height is 8'11\"";
                    }
                    if (feet === 3 && value < 0) {
                      return "Min height is 3'0\"";
                    }
                    return true;
                  }
                }}
                render={({ field }) => (
                  <input
                    {...field}
                    type="number"
                    min="0"
                    max="11"
                    placeholder="Inches (0-11)"
                    className={`w-24 px-3 py-2 border rounded-md ${
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
                required: "Height is required",
                min: { value: MIN_HEIGHT_CM, message: `Min height is ${MIN_HEIGHT_CM} cm (3'0")` },
                max: { value: MAX_HEIGHT_CM, message: `Max height is ${MAX_HEIGHT_CM} cm (8'11")` },
                validate: (value: number | undefined) => (value !== undefined && value > 0) || "Height must be greater than 0"
              }}
              render={({ field }) => (
                <input
                  {...field}
                  type="number"
                  step="0.1"
                  min={MIN_HEIGHT_CM}
                  max={MAX_HEIGHT_CM}
                  className={`w-full px-3 py-2 border rounded-md ${
                    errors.height?.value ? 'border-red-300' : 'border-gray-300'
                  }`}
                  onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                  placeholder={`e.g., 175 (${MIN_HEIGHT_CM}-${MAX_HEIGHT_CM})`}
                />
              )}
            />
          )}
          {(errors.height?.feet || errors.height?.inches || errors.height?.value) && (
            <span className="text-sm text-red-600">
              {errors.height?.feet?.message || errors.height?.inches?.message || errors.height?.value?.message}
            </span>
          )}
          {(() => {
            const feet = watchedFeet || 0;
            const inches = watchedInches || 0;
            const totalInches = feet * 12 + inches;
            const heightValue = watchedHeightValue || 0;

            if (heightUnit === 'in' && totalInches > 0) {
              return (
                <p className="text-xs text-gray-500 mt-1">
                  {feet}'{inches}" = {Math.round(totalInches * 2.54)} cm
                </p>
              );
            } else if (heightUnit === 'cm' && heightValue > 0) {
              const totalInchesCalc = heightValue / 2.54;
              const feetCalc = Math.floor(totalInchesCalc / 12);
              const inchesCalc = Math.round(totalInchesCalc % 12);
              return (
                <p className="text-xs text-gray-500 mt-1">
                  {heightValue} cm ≈ {feetCalc}'{inchesCalc}"
                </p>
              );
            }
            return null;
          })()}
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Have you competed in powerlifting before? *
        </label>
        <Controller
          name="has_competed"
          control={control}
          rules={{ required: "This field is required" }}
          render={({ field: { onChange, onBlur, name, ref, value } }) => (
            <div className="flex gap-4">
              <label className="inline-flex items-center">
                <input
                  type="radio"
                  onBlur={onBlur}
                  name={name}
                  ref={ref}
                  checked={value === true}
                  onChange={() => onChange(true)}
                  className="form-radio text-blue-600"
                />
                <span className="ml-2">Yes</span>
              </label>
              <label className="inline-flex items-center">
                <input
                  type="radio"
                  onBlur={onBlur}
                  name={name}
                  ref={ref}
                  checked={value === false}
                  onChange={() => onChange(false)}
                  className="form-radio text-blue-600"
                />
                <span className="ml-2">No</span>
              </label>
            </div>
          )}
        />
      </div>

      <h4 className="text-lg font-medium pt-4">Current Best Lifts (in kg)</h4>
      <p className="text-sm text-gray-500 mb-4">
        Enter your best gym or competition 1RM. Must be at least {MIN_LIFT_KG} kg. Max limits: S/{MAX_SQUAT_KG}kg, B/{MAX_BENCH_KG}kg, D/{MAX_DEADLIFT_KG}kg.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Best Squat (kg) *
          </label>
          <Controller
            name="best_squat_kg"
            control={control}
            rules={{
              required: "Squat is required",
              min: { value: MIN_LIFT_KG, message: `Squat must be at least ${MIN_LIFT_KG} kg` },
              max: { value: MAX_SQUAT_KG, message: `Max squat is ${MAX_SQUAT_KG} kg` },
            }}
            render={({ field }) => (
              <input
                {...field}
                type="number"
                step="0.5"
                min={MIN_LIFT_KG}
                max={MAX_SQUAT_KG}
                className={`w-full px-3 py-2 border rounded-md ${
                  errors.best_squat_kg ? 'border-red-300' : 'border-gray-300'
                }`}
                onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
              />
            )}
          />
          {errors.best_squat_kg && (
            <span className="text-sm text-red-600">{errors.best_squat_kg.message}</span>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Best Bench (kg) *
          </label>
          <Controller
            name="best_bench_kg"
            control={control}
            rules={{
              required: "Bench is required",
              min: { value: MIN_LIFT_KG, message: `Bench must be at least ${MIN_LIFT_KG} kg` },
              max: { value: MAX_BENCH_KG, message: `Max bench is ${MAX_BENCH_KG} kg` },
            }}
            render={({ field }) => (
              <input
                {...field}
                type="number"
                step="0.5"
                min={MIN_LIFT_KG}
                max={MAX_BENCH_KG}
                className={`w-full px-3 py-2 border rounded-md ${
                  errors.best_bench_kg ? 'border-red-300' : 'border-gray-300'
                }`}
                onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
              />
            )}
          />
          {errors.best_bench_kg && (
            <span className="text-sm text-red-600">{errors.best_bench_kg.message}</span>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Best Deadlift (kg) *
          </label>
          <Controller
            name="best_dead_kg"
            control={control}
            rules={{
              required: "Deadlift is required",
              min: { value: MIN_LIFT_KG, message: `Deadlift must be at least ${MIN_LIFT_KG} kg` },
              max: { value: MAX_DEADLIFT_KG, message: `Max deadlift is ${MAX_DEADLIFT_KG} kg` },
            }}
            render={({ field }) => (
              <input
                {...field}
                type="number"
                step="0.5"
                min={MIN_LIFT_KG}
                max={MAX_DEADLIFT_KG}
                className={`w-full px-3 py-2 border rounded-md ${
                  errors.best_dead_kg ? 'border-red-300' : 'border-gray-300'
                }`}
                onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
              />
            )}
          />
          {errors.best_dead_kg && (
            <span className="text-sm text-red-600">{errors.best_dead_kg.message}</span>
          )}
        </div>
      </div>

      {hasCompeted && (
        <div className="space-y-6 pt-6 border-t border-gray-200">
          <h4 className="text-lg font-medium">Last Competition PRs</h4>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Best Total (kg) *
            </label>
            <Controller
              name="best_total_kg"
              control={control}
              rules={{
                required: "Total is required",
                min: { value: MIN_LIFT_KG * 3, message: `Total must be at least ${MIN_LIFT_KG * 3} kg` },
                max: { value: MAX_TOTAL_KG, message: `Max total is ${MAX_TOTAL_KG} kg` },
                validate: (value) => {
                  const s = getValues('best_squat_kg') || 0;
                  const b = getValues('best_bench_kg') || 0;
                  const d = getValues('best_dead_kg') || 0;
                  const val = value as number || 0;
                  if (val > 0 && val < s + b + d) {
                    return "Total should typically be equal to or greater than the sum of the best lifts.";
                  }
                  return true;
                }
              }}
              render={({ field }) => (
                <input
                  {...field}
                  type="number"
                  step="0.5"
                  min={MIN_LIFT_KG * 3}
                  max={MAX_TOTAL_KG}
                  className={`w-full px-3 py-2 border rounded-md ${
                    errors.best_total_kg ? 'border-red-300' : 'border-gray-300'
                  }`}
                  onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                />
              )}
            />
            {errors.best_total_kg && (
              <span className="text-sm text-red-600">{errors.best_total_kg.message}</span>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Competition Federation *
              </label>
              <Controller
                name="comp_federation"
                control={control}
                rules={{ required: "Federation is required" }}
                render={({ field }) => (
                  <select
                    {...field}
                    className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
                      errors.comp_federation ? 'border-red-300' : 'border-gray-300'
                    }`}
                  >
                    <option value="">Select Federation</option>
                    {CANADIAN_FEDS.map(fed => (
                      <option key={fed.value} value={fed.value}>{fed.label}</option>
                    ))}
                  </select>
                )}
              />
              {errors.comp_federation && (
                <span className="text-sm text-red-600">{errors.comp_federation.message}</span>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Date PRs were set *
              </label>
              <Controller
                name="comp_pr_date"
                control={control}
                rules={{
                  required: "Date is required",
                  validate: (value) => (value && value <= new Date().toISOString().split('T')[0]) || "Date cannot be in the future"
                }}
                render={({ field }) => (
                  <input
                    {...field}
                    type="date"
                    max={new Date().toISOString().split('T')[0]}
                    className={`w-full px-3 py-2 border rounded-md ${
                      errors.comp_pr_date ? 'border-red-300' : 'border-gray-300'
                    }`}
                  />
                )}
              />
              {errors.comp_pr_date && (
                <span className="text-sm text-red-600">{errors.comp_pr_date.message}</span>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
