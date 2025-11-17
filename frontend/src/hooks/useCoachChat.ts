import { useState, useCallback, useRef } from 'react';
import { apiClient } from '@/utils/api';
import { generateUUID } from '@/utils/uuid';

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  createdAt?: Date;
}

export interface UseCoachChatOptions {
  programId?: string;
  coachContextEnabled?: boolean;
  onProgramGenerated?: (program: any) => void;
  initialMessages?: Message[];
}

export interface UseCoachChatReturn {
  messages: Message[];
  input: string;
  setInput: (input: string) => void;
  handleInputChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  handleSubmit: (e?: React.FormEvent) => Promise<void>;
  isLoading: boolean;
  error: Error | null;
  append: (message: Message) => Promise<void>;
  reload: () => Promise<void>;
  stop: () => void;
  setMessages: (messages: Message[]) => void;
}

export const useCoachChat = (options: UseCoachChatOptions = {}): UseCoachChatReturn => {
  const {
    programId,
    coachContextEnabled = false,
    onProgramGenerated,
    initialMessages = [],
  } = options;

  const [messages, setMessages] = useState<Message[]>(initialMessages);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setInput(e.target.value);
    },
    []
  );

  const sendMessage = useCallback(
    async (userMessage: string) => {
      if (!userMessage.trim()) return;

      setError(null);
      setIsLoading(true);

      const userMsg: Message = {
        id: generateUUID(),
        role: 'user',
        content: userMessage,
        createdAt: new Date(),
      };

      setMessages((prev) => [...prev, userMsg]);

      try {
        abortControllerRef.current = new AbortController();

        const response = await apiClient.chatWithAI(
          userMessage,
          programId,
          coachContextEnabled
        );

        const assistantMsg: Message = {
          id: generateUUID(),
          role: 'assistant',
          content: response.message || response.response || 'I encountered an error. Please try again.',
          createdAt: new Date(),
        };

        setMessages((prev) => [...prev, assistantMsg]);

        // Handle program generation if present
        if (response.program && onProgramGenerated) {
          onProgramGenerated(response.program);
        }
      } catch (err: any) {
        if (err.name === 'AbortError') {
          return;
        }
        const errorMsg = err.response?.data?.error || err.message || 'Failed to send message';
        setError(new Error(errorMsg));

        // Add error message to chat
        const errorMessage: Message = {
          id: generateUUID(),
          role: 'assistant',
          content: `I apologize, but I encountered an error: ${errorMsg}. Please try again.`,
          createdAt: new Date(),
        };
        setMessages((prev) => [...prev, errorMessage]);
      } finally {
        setIsLoading(false);
        abortControllerRef.current = null;
      }
    },
    [programId, coachContextEnabled, onProgramGenerated]
  );

  const handleSubmit = useCallback(
    async (e?: React.FormEvent) => {
      e?.preventDefault();
      if (!input.trim() || isLoading) return;

      const message = input;
      setInput('');
      await sendMessage(message);
    },
    [input, isLoading, sendMessage]
  );

  const append = useCallback(
    async (message: Message) => {
      if (message.role === 'user') {
        await sendMessage(message.content);
      } else {
        setMessages((prev) => [...prev, message]);
      }
    },
    [sendMessage]
  );

  const reload = useCallback(async () => {
    // Get the last user message and resend it
    const lastUserMessage = [...messages].reverse().find((m) => m.role === 'user');
    if (lastUserMessage) {
      // Remove the last assistant response
      setMessages((prev) => {
        // Find last assistant index manually for older TS compatibility
        let lastAssistantIndex = -1;
        for (let i = prev.length - 1; i >= 0; i--) {
          if (prev[i].role === 'assistant') {
            lastAssistantIndex = i;
            break;
          }
        }
        if (lastAssistantIndex > -1) {
          return prev.slice(0, lastAssistantIndex);
        }
        return prev;
      });
      await sendMessage(lastUserMessage.content);
    }
  }, [messages, sendMessage]);

  const stop = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      setIsLoading(false);
    }
  }, []);

  return {
    messages,
    input,
    setInput,
    handleInputChange,
    handleSubmit,
    isLoading,
    error,
    append,
    reload,
    stop,
    setMessages,
  };
};
