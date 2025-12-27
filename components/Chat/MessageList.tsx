'use client';

import { useEffect, useRef } from 'react';
import { Message } from '@/lib/types';
import MessageBubble from './MessageBubble';

interface MessageListProps {
  messages: Message[];
  isLoading: boolean;
}

function LoadingIndicator() {
  return (
    <div className="flex justify-start animate-fadeIn">
      <div className="bg-white/20 rounded-2xl rounded-bl-md px-4 py-3">
        <div className="flex items-center gap-2 mb-2">
          <span className="text-xl">👨‍🔧</span>
          <span className="font-medium text-sm text-white">세탁 장인</span>
        </div>
        <div className="flex items-center gap-1">
          <span className="w-2 h-2 bg-white/60 rounded-full animate-pulse-dot" style={{ animationDelay: '0ms' }} />
          <span className="w-2 h-2 bg-white/60 rounded-full animate-pulse-dot" style={{ animationDelay: '150ms' }} />
          <span className="w-2 h-2 bg-white/60 rounded-full animate-pulse-dot" style={{ animationDelay: '300ms' }} />
        </div>
      </div>
    </div>
  );
}

export default function MessageList({ messages, isLoading }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isLoading]);

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {messages.length === 0 && !isLoading && (
        <div className="flex flex-col items-center justify-center h-full text-white/70 text-center">
          <span className="text-6xl mb-4">👨‍🔧</span>
          <h2 className="text-xl font-medium mb-2">안녕하세요!</h2>
          <p className="text-sm max-w-xs">
            30년 경력의 세탁 장인입니다.<br />
            세탁, 얼룩 제거에 대해 무엇이든 물어보세요.
          </p>
        </div>
      )}

      {messages.map((message) => (
        <MessageBubble key={message.id} message={message} />
      ))}

      {isLoading && <LoadingIndicator />}

      <div ref={bottomRef} />
    </div>
  );
}
