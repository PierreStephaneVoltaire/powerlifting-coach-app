import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { apiClient } from '@/utils/api';

export const ChatPage: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [chatUrl, setChatUrl] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const initializeChat = async () => {
      if (!user) {
        navigate('/login');
        return;
      }

      try {
        // Get the OpenWebUI URL from environment or config
        const openWebUIUrl = process.env.REACT_APP_OPENWEBUI_URL || 'http://localhost:3000';

        // TODO: Implement JWT token passing to OpenWebUI for authentication
        // For now, we'll just load the iframe
        setChatUrl(openWebUIUrl);
        setLoading(false);
      } catch (error) {
        console.error('Failed to initialize chat:', error);
        setLoading(false);
      }
    };

    initializeChat();
  }, [user, navigate]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading AI Coach...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col">
      <div className="bg-white shadow-sm border-b px-6 py-3 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">AI Powerlifting Coach</h1>
          <p className="text-sm text-gray-600">
            Design your personalized training program
          </p>
        </div>
        <button
          onClick={() => navigate('/feed')}
          className="px-4 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
        >
          Back to Feed
        </button>
      </div>

      <div className="flex-1 relative">
        {chatUrl ? (
          <iframe
            src={chatUrl}
            className="w-full h-full border-0"
            title="AI Coach Chat"
            sandbox="allow-same-origin allow-scripts allow-forms allow-popups"
          />
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center max-w-md">
              <div className="text-6xl mb-4">🤖</div>
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                AI Coach Unavailable
              </h2>
              <p className="text-gray-600 mb-6">
                The AI coaching interface is currently unavailable. Please try again later or contact support.
              </p>
              <button
                onClick={() => navigate('/feed')}
                className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Go to Feed
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
