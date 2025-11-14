import React, { useState, useEffect, useRef } from 'react';
import { apiClient } from '../../utils/api';
import { ProgramArtifact } from './ProgramArtifact';
import { format } from 'date-fns';

interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  artifacts?: any[];
}

interface ChatInterfaceProps {
  programId?: string;
  competitionDate?: Date;
  currentMaxes?: {
    squat_kg?: number;
    bench_kg?: number;
    deadlift_kg?: number;
  };
  onProgramGenerated?: (program: any) => void;
}

export const ChatInterface: React.FC<ChatInterfaceProps> = ({
  programId,
  competitionDate,
  currentMaxes,
  onProgramGenerated,
}) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showArtifact, setShowArtifact] = useState<any>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    // Initialize with system context if we have competition data
    if (competitionDate || currentMaxes) {
      const contextMessage = buildContextMessage();
      if (contextMessage) {
        setMessages([{
          id: 'system-context',
          role: 'system',
          content: contextMessage,
          timestamp: new Date(),
        }]);
      }
    }

    // Load chat history if programId exists
    if (programId) {
      loadChatHistory();
    }
  }, [programId, competitionDate, currentMaxes]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const buildContextMessage = () => {
    const parts = [];

    if (competitionDate) {
      const weeksUntilComp = Math.ceil(
        (new Date(competitionDate).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24 * 7)
      );
      parts.push(`📅 Competition in ${weeksUntilComp} weeks (${format(new Date(competitionDate), 'MMM d, yyyy')})`);
    }

    if (currentMaxes) {
      const maxParts = [];
      if (currentMaxes.squat_kg) maxParts.push(`Squat: ${currentMaxes.squat_kg}kg`);
      if (currentMaxes.bench_kg) maxParts.push(`Bench: ${currentMaxes.bench_kg}kg`);
      if (currentMaxes.deadlift_kg) maxParts.push(`Deadlift: ${currentMaxes.deadlift_kg}kg`);
      if (maxParts.length > 0) {
        parts.push(`💪 Current maxes: ${maxParts.join(' | ')}`);
      }
    }

    return parts.length > 0 ? parts.join('\n') : '';
  };

  const loadChatHistory = async () => {
    // TODO: Implement chat history loading from backend
    // For now, just showing a welcome message
    setMessages(prev => [...prev, {
      id: 'welcome',
      role: 'assistant',
      content: "Hi! I'm your AI powerlifting coach. I can help you design your competition prep program, adjust your training based on your progress, and answer any questions about powerlifting. What would you like to work on today?",
      timestamp: new Date(),
    }]);
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = async () => {
    if (!inputMessage.trim() || isLoading) return;

    const userMessage: Message = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: inputMessage,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputMessage('');
    setIsLoading(true);

    try {
      const response = await apiClient.chatWithAI(inputMessage, programId, false);

      const assistantMessage: Message = {
        id: `assistant-${Date.now()}`,
        role: 'assistant',
        content: response.message || response.response || 'I apologize, but I encountered an error. Please try again.',
        timestamp: new Date(),
        artifacts: response.artifacts,
      };

      setMessages(prev => [...prev, assistantMessage]);

      // If program was generated, notify parent
      if (response.program && onProgramGenerated) {
        onProgramGenerated(response.program);
      }
    } catch (error) {
      console.error('Chat error:', error);
      setMessages(prev => [...prev, {
        id: `error-${Date.now()}`,
        role: 'assistant',
        content: 'I apologize, but I encountered an error processing your message. Please try again.',
        timestamp: new Date(),
      }]);
    } finally {
      setIsLoading(false);
      inputRef.current?.focus();
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const getWeeksUntilCompetition = () => {
    if (!competitionDate) return null;
    const weeks = Math.ceil(
      (new Date(competitionDate).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24 * 7)
    );
    return weeks;
  };

  const getCurrentPhase = (weeksUntil: number | null) => {
    if (!weeksUntil) return null;
    if (weeksUntil > 12) return 'Volume/Hypertrophy';
    if (weeksUntil > 6) return 'Strength';
    if (weeksUntil > 2) return 'Peaking';
    if (weeksUntil > 0) return 'Taper';
    return 'Competition Week';
  };

  const weeksUntilComp = getWeeksUntilCompetition();
  const currentPhase = getCurrentPhase(weeksUntilComp);

  return (
    <div className="flex flex-col h-full bg-white rounded-lg shadow-lg">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-t-lg">
        <h2 className="text-xl font-bold">AI Coach</h2>
        {(competitionDate || currentMaxes) && (
          <div className="mt-2 text-sm space-y-1 opacity-90">
            {competitionDate && weeksUntilComp !== null && (
              <div className="flex items-center gap-2">
                <span>📅 {weeksUntilComp} weeks until competition</span>
                {currentPhase && (
                  <span className="px-2 py-0.5 bg-white/20 rounded text-xs">
                    {currentPhase} Phase
                  </span>
                )}
              </div>
            )}
            {currentMaxes && (
              <div className="flex items-center gap-4 text-xs">
                {currentMaxes.squat_kg && <span>Squat: {currentMaxes.squat_kg}kg</span>}
                {currentMaxes.bench_kg && <span>Bench: {currentMaxes.bench_kg}kg</span>}
                {currentMaxes.deadlift_kg && <span>Deadlift: {currentMaxes.deadlift_kg}kg</span>}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-4">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[80%] rounded-lg px-4 py-3 ${
                message.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : message.role === 'system'
                  ? 'bg-gray-100 text-gray-600 text-sm italic'
                  : 'bg-gray-100 text-gray-900'
              }`}
            >
              <div className="whitespace-pre-wrap">{message.content}</div>

              {message.artifacts && message.artifacts.length > 0 && (
                <div className="mt-3 space-y-2">
                  {message.artifacts.map((artifact, idx) => (
                    <button
                      key={idx}
                      onClick={() => setShowArtifact(artifact)}
                      className="block w-full text-left px-3 py-2 bg-white rounded border border-gray-300 hover:border-blue-500 hover:bg-blue-50 transition"
                    >
                      <div className="flex items-center gap-2">
                        <span className="text-2xl">📋</span>
                        <div className="flex-1">
                          <div className="font-medium text-sm">{artifact.name || 'Program Preview'}</div>
                          <div className="text-xs text-gray-600">{artifact.description || 'Click to view details'}</div>
                        </div>
                        <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                      </div>
                    </button>
                  ))}
                </div>
              )}

              <div className="mt-2 text-xs opacity-70">
                {format(new Date(message.timestamp), 'h:mm a')}
              </div>
            </div>
          </div>
        ))}

        {isLoading && (
          <div className="flex justify-start">
            <div className="bg-gray-100 rounded-lg px-4 py-3">
              <div className="flex items-center gap-2">
                <div className="flex gap-1">
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
                </div>
                <span className="text-sm text-gray-600">AI is thinking...</span>
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
        <div className="flex gap-3">
          <input
            ref={inputRef}
            type="text"
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Ask about training, program adjustments, technique..."
            className="flex-1 px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={isLoading}
          />
          <button
            onClick={handleSendMessage}
            disabled={!inputMessage.trim() || isLoading}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </button>
        </div>

        <div className="mt-2 text-xs text-gray-500">
          💡 Tip: Ask me to generate a program, analyze your recent workouts, or answer technique questions
        </div>
      </div>

      {/* Artifact Modal */}
      {showArtifact && (
        <ProgramArtifact
          artifact={showArtifact}
          onClose={() => setShowArtifact(null)}
          onApprove={onProgramGenerated}
        />
      )}
    </div>
  );
};
