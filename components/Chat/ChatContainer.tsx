'use client';

import { useState } from 'react';
import { Message, ChatResponse } from '@/lib/types';
import { generateId } from '@/lib/utils';
import MessageList from './MessageList';
import ChatInput from './ChatInput';

export default function ChatContainer() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const sendMessage = async (content: string) => {
    // 사용자 메시지 추가
    const userMessage: Message = {
      id: generateId(),
      role: 'user',
      content,
      timestamp: new Date(),
    };
    setMessages((prev) => [...prev, userMessage]);
    setIsLoading(true);

    try {
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: content }),
      });

      if (!response.ok) {
        throw new Error('API request failed');
      }

      const data: ChatResponse = await response.json();

      // 봇 응답 추가
      const botMessage: Message = {
        id: generateId(),
        role: 'assistant',
        content: data.answer + (data.disclaimer ? `\n\n${data.disclaimer}` : ''),
        timestamp: new Date(),
        recommended_shops: data.recommended_shops,
      };
      setMessages((prev) => [...prev, botMessage]);
    } catch (error) {
      console.error('Chat error:', error);

      // 에러 메시지 추가
      const errorMessage: Message = {
        id: generateId(),
        role: 'assistant',
        content: '죄송합니다. 일시적인 오류가 발생했습니다. 잠시 후 다시 시도해주세요.',
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex flex-col h-screen max-w-3xl mx-auto">
      {/* Header */}
      <header className="flex items-center justify-center p-4 bg-white/10 backdrop-blur-sm">
        <div className="flex items-center gap-3">
          <span className="text-3xl">🧺</span>
          <div>
            <h1 className="text-white font-bold text-lg">Ask-Me Cleaners</h1>
            <p className="text-white/70 text-xs">세탁 장인 AI</p>
          </div>
        </div>
      </header>

      {/* Messages */}
      <MessageList messages={messages} isLoading={isLoading} />

      {/* Input */}
      <ChatInput onSend={sendMessage} disabled={isLoading} />
    </div>
  );
}
