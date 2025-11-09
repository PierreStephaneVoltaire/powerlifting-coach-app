import React, { useState } from 'react';
import { useAuthStore } from '@/store/authStore';
import { offlineQueue } from '@/utils/offlineQueue';

import { generateUUID } from '@/utils/uuid';
interface VideoUploadProps {
  onUploadComplete?: () => void;
}

const VideoUpload: React.FC<VideoUploadProps> = ({ onUploadComplete }) => {
  const { user } = useAuthStore();

  const [file, setFile] = useState<File | null>(null);
  const [movementLabel, setMovementLabel] = useState('');
  const [weight, setWeight] = useState('');
  const [unit, setUnit] = useState<'kg' | 'lb'>('lb');
  const [effortType, setEffortType] = useState<'rpe' | 'rir'>('rpe');
  const [effortValue, setEffortValue] = useState('');
  const [commentText, setCommentText] = useState('');
  const [visibility, setVisibility] = useState<'public' | 'private'>('public');
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!file || !user) return;

    setUploading(true);
    setUploadProgress(0);

    const rpeValue = effortValue
      ? effortType === 'rpe'
        ? parseFloat(effortValue)
        : 10 - parseFloat(effortValue)
      : null;

    const mediaUploadRequestedEvent = {
      schema_version: '1.0.0',
      event_type: 'media.upload_requested',
      client_generated_id: generateUUID(),
      user_id: user.id,
      timestamp: new Date().toISOString(),
      source_service: 'frontend',
      data: {
        filename: file.name,
        content_type: file.type,
        file_size: file.size,
        movement_label: movementLabel,
        weight: weight ? `${parseFloat(weight)}${unit}` : null,
        rpe: rpeValue,
        comment_text: commentText,
        visibility,
      },
    };

    console.log('Uploading video with metadata:', {
      filename: file.name,
      movement: movementLabel,
      weight: weight ? `${weight}${unit}` : 'N/A',
      effort: effortValue ? `${effortValue} ${effortType.toUpperCase()}` : 'N/A',
      rpe: rpeValue,
      visibility,
    });

    try {
      const interval = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev >= 90) {
            clearInterval(interval);
            return 90;
          }
          return prev + 10;
        });
      }, 200);

      console.log('Enqueueing media upload event...');
      await offlineQueue.enqueue(mediaUploadRequestedEvent);
      console.log('Event enqueued successfully');

      clearInterval(interval);
      setUploadProgress(100);

      setTimeout(() => {
        setFile(null);
        setMovementLabel('');
        setWeight('');
        setEffortValue('');
        setCommentText('');
        setVisibility('public');
        setUploading(false);
        setUploadProgress(0);

        if (onUploadComplete) {
          onUploadComplete();
        }
      }, 500);
    } catch (error) {
      console.error('Upload failed:', error);
      setUploading(false);
      setUploadProgress(0);
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h2 className="text-2xl font-bold mb-4">Upload Video</h2>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-2">Video File</label>
          <input
            type="file"
            accept="video/*"
            onChange={handleFileChange}
            className="w-full border rounded p-2"
            disabled={uploading}
          />
          {file && (
            <p className="text-sm text-gray-600 mt-1">
              Selected: {file.name} ({(file.size / 1024 / 1024).toFixed(2)} MB)
            </p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Movement Label *</label>
          <select
            value={movementLabel}
            onChange={(e) => setMovementLabel(e.target.value)}
            className="w-full border rounded p-2"
            required
            disabled={uploading}
          >
            <option value="">Select movement</option>
            <option value="squat">Squat</option>
            <option value="bench">Bench Press</option>
            <option value="deadlift">Deadlift</option>
            <option value="accessory">Accessory</option>
          </select>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">Weight</label>
            <div className="flex gap-2">
              <input
                type="number"
                step="0.1"
                value={weight}
                onChange={(e) => setWeight(e.target.value)}
                className="flex-1 border rounded p-2"
                placeholder="0"
                disabled={uploading}
              />
              <select
                value={unit}
                onChange={(e) => setUnit(e.target.value as 'kg' | 'lb')}
                className="border rounded p-2 bg-white"
                disabled={uploading}
              >
                <option value="lb">lb</option>
                <option value="kg">kg</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Effort</label>
            <div className="flex gap-2">
              <input
                type="number"
                step="0.5"
                min={effortType === 'rpe' ? '1' : '0'}
                max={effortType === 'rpe' ? '10' : '9'}
                value={effortValue}
                onChange={(e) => setEffortValue(e.target.value)}
                className="flex-1 border rounded p-2"
                placeholder={effortType === 'rpe' ? '1-10' : '0-9'}
                disabled={uploading}
              />
              <select
                value={effortType}
                onChange={(e) => setEffortType(e.target.value as 'rpe' | 'rir')}
                className="border rounded p-2 bg-white"
                disabled={uploading}
              >
                <option value="rpe">RPE</option>
                <option value="rir">RIR</option>
              </select>
            </div>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Comment</label>
          <textarea
            value={commentText}
            onChange={(e) => setCommentText(e.target.value)}
            className="w-full border rounded p-2"
            rows={3}
            placeholder="Optional notes about this lift..."
            disabled={uploading}
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Visibility</label>
          <div className="flex gap-4">
            <label className="flex items-center">
              <input
                type="radio"
                value="public"
                checked={visibility === 'public'}
                onChange={(e) => setVisibility(e.target.value as 'public')}
                className="mr-2"
                disabled={uploading}
              />
              Public
            </label>
            <label className="flex items-center">
              <input
                type="radio"
                value="private"
                checked={visibility === 'private'}
                onChange={(e) => setVisibility(e.target.value as 'private')}
                className="mr-2"
                disabled={uploading}
              />
              Private
            </label>
          </div>
        </div>

        {uploading && (
          <div className="w-full bg-gray-200 rounded-full h-4">
            <div
              className="bg-blue-500 h-4 rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
            <p className="text-sm text-center mt-2">{uploadProgress}% uploaded</p>
          </div>
        )}

        <button
          type="submit"
          disabled={!file || !movementLabel || uploading}
          className="w-full bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          {uploading ? 'Uploading...' : 'Upload Video'}
        </button>
      </form>
    </div>
  );
};

export default VideoUpload;
