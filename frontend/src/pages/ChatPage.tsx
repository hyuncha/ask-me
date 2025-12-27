import React, { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import '../styles/HomePage.css';
import chatService, { RecommendedShop } from '../services/chatService';

interface Message {
  id: number;
  text: string;
  sender: 'user' | 'bot';
  timestamp: Date;
  recommendedShops?: RecommendedShop[];
}

const ChatPage: React.FC = () => {
  const navigate = useNavigate();
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 1,
      text: '안녕하세요! 30년 경력의 세탁 장인입니다. 세탁, 얼룩 제거, 의류 관리에 대해 무엇이든 물어보세요.',
      sender: 'bot',
      timestamp: new Date()
    }
  ]);
  const [inputText, setInputText] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [location, setLocation] = useState<string>('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();

    if (inputText.trim() === '' || isLoading) return;

    const userMessage: Message = {
      id: messages.length + 1,
      text: inputText,
      sender: 'user',
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputText('');
    setIsLoading(true);

    try {
      const response = await chatService.sendMessage(inputText, conversationId, location || undefined);

      if (response.conversation_id && !conversationId) {
        setConversationId(response.conversation_id);
      }

      const botResponse: Message = {
        id: messages.length + 2,
        text: response.message,
        sender: 'bot',
        timestamp: new Date(),
        recommendedShops: response.recommended_shops
      };

      setMessages(prev => [...prev, botResponse]);
    } catch (error: any) {
      console.error('Error sending message:', error);

      const errorMessage: Message = {
        id: messages.length + 2,
        text: `죄송합니다. 오류가 발생했습니다: ${error?.message || '알 수 없는 오류'}`,
        sender: 'bot',
        timestamp: new Date()
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleNewChat = () => {
    chatService.resetSession();
    setMessages([
      {
        id: 1,
        text: '안녕하세요! 30년 경력의 세탁 장인입니다. 세탁, 얼룩 제거, 의류 관리에 대해 무엇이든 물어보세요.',
        sender: 'bot',
        timestamp: new Date()
      }
    ]);
    setConversationId(null);
    setLocation('');
  };

  return (
    <div className="home-page">
      <div className="chat-container">
        <div className="chat-header">
          <h1>Cleaners AI Chat</h1>
          <p>AI 기반 세탁 상담 서비스</p>
          <div className="header-links">
            <button
              onClick={handleNewChat}
              className="chatgpt-link"
              style={{ border: 'none', cursor: 'pointer' }}
            >
              새 대화 시작
            </button>
            <a href="/" className="admin-link">
              홈으로
            </a>
          </div>
        </div>

        <div className="chat-messages">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`message ${message.sender === 'user' ? 'user-message' : 'bot-message'}`}
            >
              <div className="message-content">
                <p>{message.text}</p>
                {message.recommendedShops && message.recommendedShops.length > 0 && (
                  <div className="recommended-shops">
                    <h4>🏪 추천 세탁소</h4>
                    {message.recommendedShops.map((shop, index) => (
                      <div key={index} className="shop-card">
                        <div className="shop-name">{shop.name}</div>
                        <div className="shop-details">
                          <span className="shop-rating">⭐ {shop.rating?.toFixed(1) || 'N/A'}</span>
                          <span className="shop-priority">
                            {shop.priority === 'partner' ? '파트너' : '일반'}
                          </span>
                        </div>
                        {shop.specialties && shop.specialties.length > 0 && (
                          <div className="shop-specialties">
                            전문: {shop.specialties.join(', ')}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
                <span className="message-time">
                  {message.timestamp.toLocaleTimeString('ko-KR', {
                    hour: '2-digit',
                    minute: '2-digit'
                  })}
                </span>
              </div>
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        <form className="chat-input-form" onSubmit={handleSendMessage}>
          <input
            type="text"
            className="location-input"
            placeholder="우편번호 (선택)"
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            style={{ width: '120px', marginRight: '8px' }}
          />
          <input
            type="text"
            className="chat-input"
            placeholder="세탁 관련 질문을 입력하세요..."
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
          />
          <button type="submit" className="send-button" disabled={isLoading}>
            {isLoading ? '전송 중...' : '전송'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default ChatPage;
