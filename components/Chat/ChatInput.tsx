'use client';

import { useState, FormEvent, KeyboardEvent } from 'react';
import { Send } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled: boolean;
}

export default function ChatInput({ onSend, disabled }: ChatInputProps) {
  const [input, setInput] = useState('');

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    const trimmed = input.trim();
    if (trimmed && !disabled) {
      onSend(trimmed);
      setInput('');
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="p-4 bg-white/10 backdrop-blur-sm">
      <div className="flex items-end gap-2 max-w-3xl mx-auto">
        <div className="flex-1 relative">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="세탁 관련 질문을 입력하세요..."
            disabled={disabled}
            rows={1}
            className={cn(
              'w-full resize-none rounded-2xl px-4 py-3 pr-12',
              'bg-white text-gray-800 placeholder-gray-400',
              'focus:outline-none focus:ring-2 focus:ring-primary/50',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'max-h-32'
            )}
            style={{
              minHeight: '48px',
            }}
          />
        </div>

        <button
          type="submit"
          disabled={disabled || !input.trim()}
          className={cn(
            'flex-shrink-0 w-12 h-12 rounded-full',
            'bg-white text-primary flex items-center justify-center',
            'transition-all duration-200',
            'hover:bg-white/90 hover:scale-105',
            'disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100'
          )}
        >
          <Send className="w-5 h-5" />
        </button>
      </div>

      <p className="text-center text-white/50 text-xs mt-2">
        Enter로 전송, Shift+Enter로 줄바꿈
      </p>
    </form>
  );
}
