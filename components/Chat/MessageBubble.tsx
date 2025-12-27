'use client';

import { Message, PartnerShop } from '@/lib/types';
import { cn, formatTime } from '@/lib/utils';

interface MessageBubbleProps {
  message: Message;
}

function ShopCard({ shop }: { shop: PartnerShop }) {
  return (
    <div className="bg-white/10 rounded-lg p-3 mt-2">
      <div className="flex items-center justify-between">
        <span className="font-medium">{shop.shop_name}</span>
        {shop.rating && (
          <span className="text-yellow-300 text-sm">★ {shop.rating.toFixed(1)}</span>
        )}
      </div>
      <div className="text-sm text-white/70 mt-1">
        📍 {shop.zipcode}
      </div>
      {shop.specialty.length > 0 && (
        <div className="flex flex-wrap gap-1 mt-2">
          {shop.specialty.map((spec, idx) => (
            <span
              key={idx}
              className="bg-white/20 text-xs px-2 py-0.5 rounded-full"
            >
              {spec}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

export default function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user';

  return (
    <div
      className={cn(
        'flex animate-fadeIn',
        isUser ? 'justify-end' : 'justify-start'
      )}
    >
      <div
        className={cn(
          'max-w-[80%] rounded-2xl px-4 py-3',
          isUser
            ? 'bg-white text-gray-800 rounded-br-md'
            : 'bg-white/20 text-white rounded-bl-md'
        )}
      >
        {!isUser && (
          <div className="flex items-center gap-2 mb-2">
            <span className="text-xl">👨‍🔧</span>
            <span className="font-medium text-sm">세탁 장인</span>
          </div>
        )}

        <div className="whitespace-pre-wrap">{message.content}</div>

        {/* 파트너 세탁소 추천 */}
        {message.recommended_shops && message.recommended_shops.length > 0 && (
          <div className="mt-4 pt-3 border-t border-white/20">
            <div className="text-sm font-medium mb-2">
              🏪 추천 세탁소
            </div>
            {message.recommended_shops.map((shop, idx) => (
              <ShopCard key={idx} shop={shop} />
            ))}
          </div>
        )}

        <div
          className={cn(
            'text-xs mt-2',
            isUser ? 'text-gray-400' : 'text-white/50'
          )}
        >
          {formatTime(message.timestamp)}
        </div>
      </div>
    </div>
  );
}
