import React, { useState } from 'react';
import { FeedList } from '@/components/Feed/FeedList';
import VideoUpload from '@/components/Video/VideoUpload';

export const FeedPage: React.FC = () => {
  const [showUpload, setShowUpload] = useState(false);

  return (
    <div className="max-w-4xl mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Feed</h2>
        <button
          onClick={() => setShowUpload(!showUpload)}
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          {showUpload ? 'Cancel' : 'Upload Video'}
        </button>
      </div>

      {showUpload && (
        <div className="mb-6">
          <VideoUpload onUploadComplete={() => setShowUpload(false)} />
        </div>
      )}

      <FeedList />
    </div>
  );
};
