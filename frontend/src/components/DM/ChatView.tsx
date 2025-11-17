import React, { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';

import { generateUUID } from '@/utils/uuid';
interface Message {
  message_id: string;
  sender_id: string;
  sender_name: string;
  message_body: string;
  attachments?: Array<{
    media_id: string;
    media_url: string;
    thumbnail_url?: string;
  }>;
  created_at: string;
  is_ai_coach?: boolean;
}

export const ChatView: React.FC = () => {
  const { conversationId } = useParams<{ conversationId: string }>();
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [messages, setMessages] = useState<Message[]>([]);
  const [messageText, setMessageText] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [recipientName] = useState('Coach');
  const [recipientId] = useState('recipient-id');

  useEffect(() => {
    if (conversationId) {
      loadMessages();
    }
  }, [conversationId]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const loadMessages = async () => {
    if (!conversationId) return;

    try {
      const response = await apiClient.getConversationMessages(conversationId);
      setMessages(response.messages || []);

      const cacheKey = `dm_messages_${conversationId}`;
      localStorage.setItem(cacheKey, JSON.stringify({
        messages: response.messages || [],
        timestamp: Date.now(),
      }));
    } catch (err) {
      console.error('Failed to load messages', err);

      const cacheKey = `dm_messages_${conversationId}`;
      const cached = localStorage.getItem(cacheKey);
      if (cached) {
        try {
          const { messages: cachedMessages } = JSON.parse(cached);
          setMessages(cachedMessages);
          console.info('Loaded messages from cache');
        } catch (parseErr) {
          console.error('Failed to parse cached messages', parseErr);
        }
      }
    }
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!user || !conversationId || !messageText.trim()) return;

    setIsSending(true);
    setError(null);

    const tempMessage: Message = {
      message_id: generateUUID(),
      sender_id: user.id,
      sender_name: user.name || 'You',
      message_body: messageText,
      created_at: new Date().toISOString(),
    };

    setMessages([...messages, tempMessage]);
    const currentMessageText = messageText;
    setMessageText('');

    try {
      const event = {
        schema_version: '1.0.0',
        event_type: 'dm.message.sent',
        client_generated_id: tempMessage.message_id,
        user_id: user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          conversation_id: conversationId,
          sender_id: user.id,
          recipient_id: recipientId,
          message_body: currentMessageText,
          attachments: [],
        },
      };

      await apiClient.submitEvent(event);
      console.info('DM message sent', { conversation_id: conversationId });
    } catch (err: any) {
      console.error('Failed to send message', err);
      if (!err.queued) {
        setError('Failed to send message. Please try again.');
        setMessageText(currentMessageText);
        setMessages(messages);
      }
    } finally {
      setIsSending(false);
    }
  };

  const handlePinAttempts = async () => {
    if (!user || !conversationId) return;

    try {
      const event = {
        schema_version: '1.0.0',
        event_type: 'dm.pin.attempts',
        client_generated_id: generateUUID(),
        user_id: user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          conversation_id: conversationId,
        },
      };

      await apiClient.submitEvent(event);
      console.info('Attempts pinned', { conversation_id: conversationId });
    } catch (err: any) {
      console.error('Failed to pin attempts', err);
    }
  };

  return (
    <div className="max-w-4xl mx-auto h-screen flex flex-col">
      <div className="bg-white shadow-sm border-b border-gray-200 p-4 flex items-center justify-between">
        <div className="flex items-center">
          <button
            onClick={() => navigate('/dm')}
            className="mr-4 text-gray-600 hover:text-gray-900"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h2 className="text-lg font-semibold text-gray-900">{recipientName}</h2>
          </div>
        </div>
        <button
          onClick={handlePinAttempts}
          className="px-3 py-1 text-sm bg-blue-50 text-blue-700 rounded-md hover:bg-blue-100"
        >
          Pin Attempts
        </button>
      </div>

      <div className="flex-1 overflow-y-auto bg-gray-50 p-4 space-y-4">
        {messages.map((message) => {
          const isOwnMessage = message.sender_id === user?.id;

          return (
            <div
              key={message.message_id}
              className={`flex ${isOwnMessage ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-sm rounded-lg px-4 py-2 ${
                  isOwnMessage
                    ? 'bg-blue-600 text-white'
                    : message.is_ai_coach
                    ? 'bg-purple-100 text-purple-900 border border-purple-200'
                    : 'bg-white text-gray-900 border border-gray-200'
                }`}
              >
                {!isOwnMessage && (
                  <p className="text-xs font-semibold mb-1 opacity-75">
                    {message.is_ai_coach ? 'AI Coach' : message.sender_name}
                  </p>
                )}
                <p className="text-sm">{message.message_body}</p>
                {message.attachments && message.attachments.length > 0 && (
                  <div className="mt-2 space-y-2">
                    {message.attachments.map((attachment) => (
                      <div key={attachment.media_id}>
                        <video
                          src={attachment.media_url}
                          controls
                          className="w-full rounded"
                          poster={attachment.thumbnail_url}
                        />
                      </div>
                    ))}
                  </div>
                )}
                <p
                  className={`text-xs mt-1 ${
                    isOwnMessage ? 'text-blue-100' : 'text-gray-500'
                  }`}
                >
                  {new Date(message.created_at).toLocaleTimeString()}
                </p>
              </div>
            </div>
          );
        })}
        <div ref={messagesEndRef} />
      </div>

      {error && (
        <div className="bg-red-50 border-t border-red-200 p-3">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      <form onSubmit={handleSendMessage} className="bg-white border-t border-gray-200 p-4">
        <div className="flex items-end gap-2">
          <textarea
            value={messageText}
            onChange={(e) => setMessageText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleSendMessage(e);
              }
            }}
            placeholder="Type a message..."
            rows={2}
            className="flex-1 px-3 py-2 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            disabled={isSending || !messageText.trim()}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </button>
        </div>
      </form>
    </div>
  );
};
