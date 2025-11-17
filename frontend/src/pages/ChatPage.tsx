import React, { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { apiClient } from '@/utils/api';
import ReactMarkdown from 'react-markdown';
import { format, differenceInWeeks } from 'date-fns';
import { generateUUID } from '@/utils/uuid';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp?: string;
}

interface UserSettings {
  competition_date?: string;
  best_squat_kg?: number;
  best_bench_kg?: number;
  best_dead_kg?: number;
  squat_goal_value?: number;
  bench_goal_value?: number;
  dead_goal_value?: number;
  training_days_per_week?: number;
}

export const ChatPage: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [userSettings, setUserSettings] = useState<UserSettings | null>(null);
  const [currentProposal, setCurrentProposal] = useState<any>(null);
  const [isInitializing, setIsInitializing] = useState(true);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isLoading]);

  useEffect(() => {
    if (!user) {
      navigate('/login');
      return;
    }
    loadInitialData();
  }, [user, navigate]);

  const loadInitialData = async () => {
    setIsInitializing(true);
    try {
      const [settingsResponse, conversationResponse] = await Promise.all([
        apiClient.getUserSettings().catch(() => null),
        apiClient.getAIConversation().catch(() => ({ has_conversation: false })),
      ]);

      if (settingsResponse) {
        setUserSettings(settingsResponse);
      }

      if (conversationResponse.has_conversation && conversationResponse.conversation) {
        const conv = conversationResponse.conversation;
        const loadedMessages: Message[] = conv.messages.map((msg: any) => ({
          id: msg.id || generateUUID(),
          role: msg.role,
          content: msg.content,
          timestamp: msg.timestamp,
        }));
        setMessages(loadedMessages);

        if (conversationResponse.last_program_proposal) {
          setCurrentProposal(conversationResponse.last_program_proposal);
        }
      } else {
        const introMessage = `I'm ready to design your personalized powerlifting program. I have your profile information loaded.

To get started, just say **"Let's create my program"** or tell me about your goals, and I'll propose a structured training plan for your upcoming competition.

I'll show you:
- Phase breakdown (hypertrophy, strength, peaking)
- Weekly training structure
- Sets, reps, and intensity for each lift

You can then adjust anything you'd like before approving it.`;

        setMessages([
          {
            id: 'intro',
            role: 'assistant',
            content: introMessage,
          },
        ]);
      }
    } catch (err) {
      console.error('Failed to load initial data:', err);
    } finally {
      setIsInitializing(false);
    }
  };

  const sendMessage = async (userMessage: string) => {
    if (!userMessage.trim() || isLoading) return;

    setError(null);
    setIsLoading(true);

    const userMsg: Message = {
      id: generateUUID(),
      role: 'user',
      content: userMessage,
      timestamp: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMsg]);
    setInput('');

    try {
      const response = await apiClient.chatWithAI(userMessage);

      const assistantMsg: Message = {
        id: generateUUID(),
        role: 'assistant',
        content: response.message,
        timestamp: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, assistantMsg]);

      if (response.program_proposal) {
        setCurrentProposal(response.program_proposal);
      }
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || err.message || 'Failed to send message';
      setError(errorMsg);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault();
    sendMessage(input);
  };

  const getWeeksUntilComp = () => {
    if (!userSettings?.competition_date) return null;
    const compDate = new Date(userSettings.competition_date);
    const weeks = differenceInWeeks(compDate, new Date());
    return weeks > 0 ? weeks : 0;
  };

  const weeksUntilComp = getWeeksUntilComp();

  if (!user) {
    return null;
  }

  if (isInitializing) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-50">
        <div className="text-center">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading your coaching session...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-4 py-3 shadow-sm">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-gray-900">AI Powerlifting Coach</h1>
            <div className="flex items-center space-x-4 text-sm text-gray-600">
              {weeksUntilComp !== null && (
                <span className="font-medium text-blue-600">
                  {weeksUntilComp} weeks until competition
                </span>
              )}
              {userSettings?.training_days_per_week && (
                <span>{userSettings.training_days_per_week} days/week</span>
              )}
            </div>
          </div>
          <div className="flex items-center space-x-2">
            {currentProposal && (
              <button
                onClick={() => {
                  console.log('Current proposal:', currentProposal);
                  alert('Program proposal ready. Implement approval flow here.');
                }}
                className="px-4 py-2 text-sm font-medium text-white bg-green-600 hover:bg-green-700 rounded-lg transition-colors"
              >
                Review Program
              </button>
            )}
            <button
              onClick={() => navigate('/feed')}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
            >
              Back
            </button>
          </div>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto px-4 py-6">
        <div className="max-w-3xl mx-auto space-y-6">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-[85%] rounded-2xl px-4 py-3 shadow-sm ${
                  message.role === 'user'
                    ? 'bg-blue-600 text-white'
                    : 'bg-white text-gray-900 border border-gray-200'
                }`}
              >
                {message.role === 'assistant' ? (
                  <div className="prose prose-sm max-w-none">
                    <ReactMarkdown
                      components={{
                        p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
                        ul: ({ children }) => (
                          <ul className="list-disc pl-4 mb-2 space-y-1">{children}</ul>
                        ),
                        ol: ({ children }) => (
                          <ol className="list-decimal pl-4 mb-2 space-y-1">{children}</ol>
                        ),
                        li: ({ children }) => <li className="text-gray-800">{children}</li>,
                        strong: ({ children }) => (
                          <strong className="font-semibold text-gray-900">{children}</strong>
                        ),
                        table: ({ children }) => (
                          <div className="overflow-x-auto my-2">
                            <table className="min-w-full border-collapse border border-gray-300">
                              {children}
                            </table>
                          </div>
                        ),
                        th: ({ children }) => (
                          <th className="border border-gray-300 px-2 py-1 bg-gray-100 text-left text-xs font-medium">
                            {children}
                          </th>
                        ),
                        td: ({ children }) => (
                          <td className="border border-gray-300 px-2 py-1 text-xs">{children}</td>
                        ),
                        code: ({ children, className }) => {
                          const isBlock = className?.includes('language-');
                          if (isBlock) {
                            return (
                              <code className="block bg-gray-900 text-gray-100 p-3 rounded-lg overflow-x-auto text-xs">
                                {children}
                              </code>
                            );
                          }
                          return (
                            <code className="bg-gray-100 px-1.5 py-0.5 rounded text-sm font-mono text-gray-800">
                              {children}
                            </code>
                          );
                        },
                        pre: ({ children }) => <div className="my-2">{children}</div>,
                      }}
                    >
                      {message.content}
                    </ReactMarkdown>
                  </div>
                ) : (
                  <div className="whitespace-pre-wrap">{message.content}</div>
                )}

                {message.timestamp && (
                  <div
                    className={`text-xs mt-2 ${
                      message.role === 'user' ? 'text-blue-200' : 'text-gray-400'
                    }`}
                  >
                    {format(new Date(message.timestamp), 'h:mm a')}
                  </div>
                )}
              </div>
            </div>
          ))}

          {isLoading && (
            <div className="flex justify-start">
              <div className="bg-white border border-gray-200 rounded-2xl px-4 py-3 shadow-sm">
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
                  <span className="text-sm text-gray-500">Analyzing and generating response...</span>
                </div>
              </div>
            </div>
          )}

          {error && (
            <div className="flex justify-center">
              <div className="bg-red-50 border border-red-200 rounded-lg px-4 py-3 text-red-700 text-sm">
                {error}
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>
      </div>

      <div className="bg-white border-t border-gray-200 px-4 py-4">
        <div className="max-w-3xl mx-auto">
          <form onSubmit={handleSubmit} className="flex items-end space-x-3">
            <div className="flex-1 relative">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Describe your goals or ask to create your program..."
                rows={1}
                className="w-full resize-none rounded-xl border border-gray-300 px-4 py-3 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 transition-colors"
                style={{ minHeight: '48px', maxHeight: '120px' }}
                onInput={(e) => {
                  const target = e.target as HTMLTextAreaElement;
                  target.style.height = 'auto';
                  target.style.height = `${Math.min(target.scrollHeight, 120)}px`;
                }}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    handleSubmit();
                  }
                }}
                disabled={isLoading}
              />
            </div>
            <button
              type="submit"
              disabled={!input.trim() || isLoading}
              className="flex-shrink-0 bg-blue-600 text-white rounded-xl px-6 py-3 font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Send
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
