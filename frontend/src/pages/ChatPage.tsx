import React, { useEffect, useState } from 'react';
import { useChat } from '@ai-sdk/react';
import { useNavigate } from 'react-router-dom';
import { apiClient, API_BASE_URL } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';
import axios from 'axios';

interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
}

export const ChatPage: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingSettings, setIsLoadingSettings] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!user) {
      navigate('/login');
      return;
    }
    loadInitialPrompt();
  }, [user, navigate]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const loadInitialPrompt = async () => {
    setIsLoadingSettings(true);
    try {
      const settings = await apiClient.getUserSettings();

      const weeksUntilComp = settings.competition_date
        ? Math.ceil((new Date(settings.competition_date).getTime() - Date.now()) / (1000 * 60 * 60 * 24 * 7))
        : null;

      let contextParts = [
        '👋 Welcome! I\'m your AI assistant.',
        '',
        'Here\'s what I know about you:',
      ];

      if (settings.competition_date) {
        contextParts.push(`📅 Competition in ${weeksUntilComp} weeks`);
      }

      if (settings.training_days_per_week) {
        contextParts.push(`🏋️ Training ${settings.training_days_per_week} days per week`);
      }

      if (settings.best_squat_kg || settings.best_bench_kg || settings.best_dead_kg) {
        contextParts.push('');
        contextParts.push('💪 Current maxes:');
        if (settings.best_squat_kg) contextParts.push(`  - Squat: ${settings.best_squat_kg}kg`);
        if (settings.best_bench_kg) contextParts.push(`  - Bench: ${settings.best_bench_kg}kg`);
        if (settings.best_dead_kg) contextParts.push(`  - Deadlift: ${settings.best_dead_kg}kg`);
      }

      if (settings.squat_goal_value || settings.bench_goal_value || settings.dead_goal_value) {
        contextParts.push('');
        contextParts.push('🎯 Goals:');
        if (settings.squat_goal_value) contextParts.push(`  - Squat: ${settings.squat_goal_value}kg`);
        if (settings.bench_goal_value) contextParts.push(`  - Bench: ${settings.bench_goal_value}kg`);
        if (settings.dead_goal_value) contextParts.push(`  - Deadlift: ${settings.dead_goal_value}kg`);
      }

      contextParts.push('');
      contextParts.push('How can I help you today?');

      setMessages([{
        id: 'initial-system-message',
        role: 'system',
        content: contextParts.join('\n'),
      }]);
    } catch (err) {
      console.error('Failed to get user settings:', err);
      setMessages([{
        id: 'initial-system-message',
        role: 'system',
        content: '👋 Welcome! How can I help you today?',
      }]);
    } finally {
      setIsLoadingSettings(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading) return;

    const userMessage: Message = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: input,
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    try {
      const response = await axios.post(
        `${API_BASE_URL}/v1/chat/completions`,
        {
          model: 'gpt-3.5-turbo',
          messages: [...messages, userMessage].map((m) => ({
            role: m.role,
            content: m.content,
          })),
        },
        {
          headers: {
            'Content-Type': 'application/json',
          },
        }
      );

      const assistantMessage: Message = {
        id: `assistant-${Date.now()}`,
        role: 'assistant',
        content: response.data.choices[0].message.content,
      };

      setMessages((prev) => [...prev, assistantMessage]);
    } catch (err) {
      console.error('Failed to send message:', err);
      const errorMessage: Message = {
        id: `error-${Date.now()}`,
        role: 'assistant',
        content: 'Sorry, I encountered an error. Please try again.',
      };
      setMessages((prev) => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoadingSettings) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-50">
        <div className="text-center">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading chat...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4 shadow-sm">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">AI Chat</h1>
            <p className="text-sm text-gray-600">Chat with AI powered by LiteLLM</p>
          </div>
          <button
            onClick={() => navigate('/feed')}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
          >
            Back
          </button>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto px-6 py-4">
        <div className="max-w-4xl mx-auto space-y-4">
          {messages.length === 0 && (
            <div className="text-center py-12">
              <div className="text-gray-400 text-lg mb-2">👋 Start a conversation</div>
              <p className="text-gray-500 text-sm">Ask me anything!</p>
            </div>
          )}

          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${
                message.role === 'user' ? 'justify-end' : 'justify-start'
              }`}
            >
              <div
                className={`max-w-[80%] rounded-lg px-4 py-3 ${
                  message.role === 'user'
                    ? 'bg-blue-600 text-white'
                    : message.role === 'system'
                    ? 'bg-gray-100 text-gray-800 border border-gray-300'
                    : 'bg-white text-gray-900 border border-gray-200 shadow-sm'
                }`}
              >
                <div className="whitespace-pre-wrap break-words">{message.content}</div>
              </div>
            </div>
          ))}

          {isLoading && (
            <div className="flex justify-start">
              <div className="bg-white border border-gray-200 rounded-lg px-4 py-3 shadow-sm">
                <div className="flex items-center space-x-2">
                  <div className="flex space-x-1">
                    <div
                      className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                      style={{ animationDelay: '0ms' }}
                    />
                    <div
                      className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                      style={{ animationDelay: '150ms' }}
                    />
                    <div
                      className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                      style={{ animationDelay: '300ms' }}
                    />
                  </div>
                  <span className="text-sm text-gray-500">AI is thinking...</span>
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>
      </div>

      <div className="bg-white border-t border-gray-200 px-6 py-4">
        <div className="max-w-4xl mx-auto">
          <form onSubmit={handleSubmit} className="flex items-center space-x-3">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Type your message..."
              disabled={isLoading}
              className="flex-1 px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
            <button
              type="submit"
              disabled={!input.trim() || isLoading}
              className="px-6 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Send
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
