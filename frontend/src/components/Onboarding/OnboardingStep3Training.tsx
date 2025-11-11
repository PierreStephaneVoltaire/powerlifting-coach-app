import React from 'react';
import { Controller } from 'react-hook-form';
import { StepProps, SESSION_MIN_LENGTH } from './OnboardingFormTypes';

interface RecoveryRatingButtonsProps {
  field: any;
  currentValue: number;
}

const RecoveryRatingButtons: React.FC<RecoveryRatingButtonsProps> = ({ field, currentValue }) => (
  <div className="flex justify-center space-x-4">
    <button
      type="button"
      onClick={() => field.onChange(1)}
      className={`text-2xl p-2 rounded-full cursor-pointer ${
        currentValue === 1
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
      className={`text-2xl p-2 rounded-full cursor-pointer ${
        currentValue === 2
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
      className={`text-2xl p-2 rounded-full cursor-pointer ${
        currentValue === 3
          ? 'bg-green-100 border-2 border-green-500'
          : 'border-2 border-gray-200'
      }`}
      title="Happy - Can handle more than 2 heavy sessions per week"
    >
      😊
    </button>
  </div>
);

const getRecoveryLabel = (rating: number) => {
  if (rating === 1) return "😞 Need more than a week to recover";
  if (rating === 2) return "😐 Can handle 1-2 heavy sessions per week";
  if (rating === 3) return "😊 Can handle 2+ heavy sessions per week";
  return "";
};

export const OnboardingStep3Training: React.FC<StepProps> = ({
  control,
  errors,
  watch,
}) => {
  return (
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
            validate: (value) => (value as number) > 0 || "Training days must be greater than 0"
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
          Session Length (minutes) *
        </label>
        <Controller
          name="session_length_minutes"
          control={control}
          rules={{
            required: "Session length is required",
            min: { value: SESSION_MIN_LENGTH, message: `Session length must be at least ${SESSION_MIN_LENGTH} minutes` },
            max: { value: 360, message: "Session length must be no more than 360 minutes (6 hours)" },
          }}
          render={({ field }) => (
            <input
              {...field}
              type="number"
              min={SESSION_MIN_LENGTH}
              max="360"
              className={`w-full px-3 py-2 border rounded-md ${
                errors.session_length_minutes
                  ? 'border-red-300'
                  : 'border-gray-300'
              }`}
              onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
              placeholder={`${SESSION_MIN_LENGTH}-360 minutes`}
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
          Volume Preference *
        </label>
        <Controller
          name="volume_preference"
          control={control}
          rules={{ required: "Volume preference is required" }}
          render={({ field }) => (
            <select
              {...field}
              className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
                errors.volume_preference ? 'border-red-300' : 'border-gray-300'
              }`}
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
            </select>
          )}
        />
        {errors.volume_preference && (
          <span className="text-sm text-red-600">{errors.volume_preference.message}</span>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Recovery Ratings *
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
              rules={{ required: "Squat recovery rating is required" }}
              render={({ field }) => (
                <RecoveryRatingButtons
                  field={field}
                  currentValue={watch('recovery_rating_squat')}
                />
              )}
            />
            {errors.recovery_rating_squat && (
              <span className="text-sm text-red-600">{errors.recovery_rating_squat.message}</span>
            )}
            <p className="text-xs text-gray-500 mt-2">
              {getRecoveryLabel(watch('recovery_rating_squat'))}
            </p>
          </div>

          <div className="text-center">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Bench Recovery
            </label>
            <Controller
              name="recovery_rating_bench"
              control={control}
              rules={{ required: "Bench recovery rating is required" }}
              render={({ field }) => (
                <RecoveryRatingButtons
                  field={field}
                  currentValue={watch('recovery_rating_bench')}
                />
              )}
            />
            {errors.recovery_rating_bench && (
              <span className="text-sm text-red-600">{errors.recovery_rating_bench.message}</span>
            )}
            <p className="text-xs text-gray-500 mt-2">
              {getRecoveryLabel(watch('recovery_rating_bench'))}
            </p>
          </div>

          <div className="text-center">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Deadlift Recovery
            </label>
            <Controller
              name="recovery_rating_dead"
              control={control}
              rules={{ required: "Deadlift recovery rating is required" }}
              render={({ field }) => (
                <RecoveryRatingButtons
                  field={field}
                  currentValue={watch('recovery_rating_dead')}
                />
              )}
            />
            {errors.recovery_rating_dead && (
              <span className="text-sm text-red-600">{errors.recovery_rating_dead.message}</span>
            )}
            <p className="text-xs text-gray-500 mt-2">
              {getRecoveryLabel(watch('recovery_rating_dead'))}
            </p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Squat Foot Stance *
          </label>
          <Controller
            name="squat_stance"
            control={control}
            rules={{ required: "Squat foot stance is required" }}
            render={({ field }) => (
              <select
                {...field}
                className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
                  errors.squat_stance ? 'border-red-300' : 'border-gray-300'
                }`}
              >
                <option value="narrow">Narrow</option>
                <option value="medium">Medium</option>
                <option value="wide">Wide</option>
              </select>
            )}
          />
          {errors.squat_stance && (
            <span className="text-sm text-red-600">{errors.squat_stance.message}</span>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Squat Bar Position *
          </label>
          <Controller
            name="squat_bar_position"
            control={control}
            rules={{ required: "Squat bar position is required" }}
            render={({ field }) => (
              <select
                {...field}
                className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
                  errors.squat_bar_position ? 'border-red-300' : 'border-gray-300'
                }`}
              >
                <option value="high">High Bar</option>
                <option value="medium">Medium Bar</option>
                <option value="low">Low Bar</option>
                <option value="french">French</option>
              </select>
            )}
          />
          {errors.squat_bar_position && (
            <span className="text-sm text-red-600">{errors.squat_bar_position.message}</span>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Deadlift Style *
          </label>
          <Controller
            name="deadlift_style"
            control={control}
            rules={{ required: "Deadlift style is required" }}
            render={({ field }) => (
              <select
                {...field}
                className={`w-full px-3 py-2 border rounded-md appearance-none cursor-pointer ${
                  errors.deadlift_style ? 'border-red-300' : 'border-gray-300'
                }`}
              >
                <option value="conventional">Conventional</option>
                <option value="sumo">Sumo</option>
              </select>
            )}
          />
          {errors.deadlift_style && (
            <span className="text-sm text-red-600">{errors.deadlift_style.message}</span>
          )}
        </div>
      </div>
    </div>
  );
};
