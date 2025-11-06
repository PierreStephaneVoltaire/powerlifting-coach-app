import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { ConversationList } from '@/components/DM/ConversationList';
import { ChatView } from '@/components/DM/ChatView';

export const DMPage: React.FC = () => {
  return (
    <div className="max-w-6xl mx-auto p-4">
      <Routes>
        <Route index element={<ConversationList />} />
        <Route path=":conversationId" element={<ChatView />} />
      </Routes>
    </div>
  );
};
