import React from 'react';
import { Controller } from 'react-hook-form';
import {
  StepProps,
  MAX_SQUAT_KG,
  MAX_BENCH_KG,
  MAX_DEADLIFT_KG,
  MIN_LIFT_KG,
} from './OnboardingFormTypes';
import { getMinCompetitionDate } from './OnboardingFormValidation';

export const OnboardingStep2Goals: React.FC<StepProps> = ({
  control,
  errors,
  watch,
}) => {
  const watchedSquatGoal = watch('squat_goal.value') as number;
  const watchedBenchGoal = watch('bench_goal.value') as number;
  const watchedDeadGoal = watch('dead_goal.value') as number;

  return (
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
                className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
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
            rules={{
              required: "Competition date is required",
              validate: (value) => (value && value >= getMinCompetitionDate()) || `Date must be at least two weeks from today (${getMinCompetitionDate()})`
            }}
            render={({ field }) => (
              <input
                {...field}
                type="date"
                min={getMinCompetitionDate()}
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

      <h4 className="text-lg font-medium pt-4">Goal Lifts (in kg)</h4>
      <p className="text-sm text-gray-500 mb-4">
        Your goals must be at least {MIN_LIFT_KG} kg. Max limits: S/{MAX_SQUAT_KG}kg, B/{MAX_BENCH_KG}kg, D/{MAX_DEADLIFT_KG}kg.
      </p>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Squat Goal (kg) *
        </label>
        <Controller
          name="squat_goal.value"
          control={control}
          rules={{
            required: "Squat goal is required",
            min: { value: MIN_LIFT_KG, message: `Squat goal must be at least ${MIN_LIFT_KG} kg` },
            max: { value: MAX_SQUAT_KG, message: `Max squat goal is ${MAX_SQUAT_KG} kg` },
          }}
          render={({ field }) => (
            <input
              {...field}
              type="number"
              step="0.5"
              min={MIN_LIFT_KG}
              max={MAX_SQUAT_KG}
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
        {watchedSquatGoal > 0 && (
          <p className="text-xs text-gray-500 mt-1">
            {watchedSquatGoal} kg = {Math.round(watchedSquatGoal * 2.20462)} lbs
          </p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Bench Goal (kg) *
        </label>
        <Controller
          name="bench_goal.value"
          control={control}
          rules={{
            required: "Bench goal is required",
            min: { value: MIN_LIFT_KG, message: `Bench goal must be at least ${MIN_LIFT_KG} kg` },
            max: { value: MAX_BENCH_KG, message: `Max bench goal is ${MAX_BENCH_KG} kg` },
          }}
          render={({ field }) => (
            <input
              {...field}
              type="number"
              step="0.5"
              min={MIN_LIFT_KG}
              max={MAX_BENCH_KG}
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
        {watchedBenchGoal > 0 && (
          <p className="text-xs text-gray-500 mt-1">
            {watchedBenchGoal} kg = {Math.round(watchedBenchGoal * 2.20462)} lbs
          </p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Deadlift Goal (kg) *
        </label>
        <Controller
          name="dead_goal.value"
          control={control}
          rules={{
            required: "Deadlift goal is required",
            min: { value: MIN_LIFT_KG, message: `Deadlift goal must be at least ${MIN_LIFT_KG} kg` },
            max: { value: MAX_DEADLIFT_KG, message: `Max deadlift goal is ${MAX_DEADLIFT_KG} kg` },
          }}
          render={({ field }) => (
            <input
              {...field}
              type="number"
              step="0.5"
              min={MIN_LIFT_KG}
              max={MAX_DEADLIFT_KG}
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
        {watchedDeadGoal > 0 && (
          <p className="text-xs text-gray-500 mt-1">
            {watchedDeadGoal} kg = {Math.round(watchedDeadGoal * 2.20462)} lbs
          </p>
        )}
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
                className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
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
                className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
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
  );
};
