import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { ChatInterface } from '@/components/Chat/ChatInterface';

export const ChatPage: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();

  if (!user) {
    navigate('/login');
    return null;
  }

  const handleProgramGenerated = (program: any) => {
    // Navigate to program page or handle the generated program
    console.log('Program generated:', program);
    if (program?.id) {
      navigate(`/program/${program.id}`);
    }
  };

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

      <div className="flex-1 overflow-hidden">
        <ChatInterface
          onProgramGenerated={handleProgramGenerated}
        />
      </div>
    </div>
  );
};
