import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface Conversation {
  conversation_id: string;
  participant_id: string;
  participant_name: string;
  last_message: string;
  last_message_timestamp: string;
  unread_count: number;
  is_ai_coach?: boolean;
}

export const ConversationList: React.FC = () => {
  const navigate = useNavigate();
  const [conversations] = useState<Conversation[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  const filteredConversations = conversations.filter(
    (conv) =>
      conv.participant_name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg">
        <div className="p-6 border-b border-gray-200">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Messages</h2>
          <div className="relative">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search conversations..."
              className="w-full px-4 py-2 pl-10 border border-gray-300 rounded-lg"
            />
            <svg
              className="absolute left-3 top-3 w-5 h-5 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>
        </div>

        {filteredConversations.length === 0 ? (
          <div className="p-12 text-center">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
            <h3 className="mt-2 text-lg font-medium text-gray-900">No conversations</h3>
            <p className="mt-1 text-sm text-gray-500">
              Start chatting with your coach or the AI assistant
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {filteredConversations.map((conversation) => (
              <div
                key={conversation.conversation_id}
                onClick={() => navigate(`/dm/${conversation.conversation_id}`)}
                className="p-4 hover:bg-gray-50 cursor-pointer transition-colors"
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start flex-1">
                    <div className="flex-shrink-0">
                      <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center">
                        {conversation.is_ai_coach ? (
                          <svg
                            className="w-6 h-6 text-blue-600"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                            />
                          </svg>
                        ) : (
                          <span className="text-blue-600 font-semibold text-lg">
                            {conversation.participant_name.charAt(0).toUpperCase()}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="ml-3 flex-1">
                      <div className="flex items-center justify-between">
                        <h3 className="text-sm font-semibold text-gray-900">
                          {conversation.participant_name}
                          {conversation.is_ai_coach && (
                            <span className="ml-2 text-xs font-normal text-blue-600">
                              AI Coach
                            </span>
                          )}
                        </h3>
                        <span className="text-xs text-gray-500">
                          {new Date(conversation.last_message_timestamp).toLocaleDateString()}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 truncate mt-1">
                        {conversation.last_message}
                      </p>
                    </div>
                  </div>
                  {conversation.unread_count > 0 && (
                    <div className="ml-2 flex-shrink-0">
                      <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-blue-600 text-white text-xs font-medium">
                        {conversation.unread_count}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
