import React, { useState, useRef, useEffect } from 'react';
import '../styles/HomePage.css';
import chatService from '../services/chatService';
import axios from 'axios';

interface Message {
  id: number;
  text: string;
  sender: 'user' | 'bot';
  timestamp: Date;
  imageUrl?: string;
}

const CATEGORIES = [
  { id: 'stain', name: '얼룩 제거', icon: '🧴' },
  { id: 'fabric', name: '원단 관리', icon: '🧵' },
  { id: 'label', name: '세탁 라벨', icon: '🏷️' },
  { id: 'machine', name: '세탁기 사용', icon: '🧺' },
  { id: 'dry', name: '드라이클리닝', icon: '👔' },
  { id: 'ironing', name: '다림질', icon: '🔥' },
  { id: 'storage', name: '보관 방법', icon: '📦' },
  { id: 'inquiry', name: '1:1문의', icon: '✉️', isEmail: true },
];

const INQUIRY_EMAIL = 'yesibo2@gmail.com';

const CHATGPT_LINK = 'https://chatgpt.com/g/g-sCqYnBBT5-professional-dry-cleaners-comcleaners';

const HomePage: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 1,
      text: '안녕하세요! 세탁 관련 질문이 있으시면 무엇이든 물어보세요. 세탁 라벨이나 얼룩 사진을 업로드하시면 더 정확한 답변을 드릴 수 있습니다.',
      sender: 'bot',
      timestamp: new Date()
    }
  ]);
  const [inputText, setInputText] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadedImageUrl, setUploadedImageUrl] = useState<string | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setSelectedFile(file);
    setIsLoading(true);

    try {
      const formData = new FormData();
      formData.append('file', file);

      const response = await axios.post('/api/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      const imageUrl = response.data.file_url;
      setUploadedImageUrl(imageUrl);

      // Add image message
      const imageMessage: Message = {
        id: messages.length + 1,
        text: '이미지를 업로드했습니다.',
        sender: 'user',
        timestamp: new Date(),
        imageUrl: imageUrl,
      };

      setMessages(prev => [...prev, imageMessage]);
    } catch (error) {
      console.error('Error uploading file:', error);
      alert('파일 업로드에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();

    if (inputText.trim() === '' || isLoading) return;

    let messageText = inputText;
    if (uploadedImageUrl) {
      messageText = `[이미지 첨부] ${inputText}`;
    }

    const userMessage: Message = {
      id: messages.length + 1,
      text: inputText,
      sender: 'user',
      timestamp: new Date(),
      imageUrl: uploadedImageUrl || undefined,
    };

    setMessages(prev => [...prev, userMessage]);
    setInputText('');
    setUploadedImageUrl(null);
    setSelectedFile(null);
    setIsLoading(true);

    try {
      // Call backend API
      const response = await chatService.sendMessage(messageText, conversationId);

      // Save conversation ID
      if (response.conversation_id && !conversationId) {
        setConversationId(response.conversation_id);
      }

      const botResponse: Message = {
        id: messages.length + 2,
        text: response.message,
        sender: 'bot',
        timestamp: new Date()
      };

      setMessages(prev => [...prev, botResponse]);
    } catch (error) {
      const errorMessage: Message = {
        id: messages.length + 2,
        text: '죄송합니다. 오류가 발생했습니다. 나중에 다시 시도해주세요.',
        sender: 'bot',
        timestamp: new Date()
      };
      setMessages(prev => [...prev, errorMessage]);
      console.error('Error sending message:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCategoryClick = (categoryId: string) => {
    const category = CATEGORIES.find(c => c.id === categoryId);
    
    if (category && (category as any).isEmail) {
      window.location.href = `mailto:${INQUIRY_EMAIL}?subject=[Cleaners AI] 1:1 문의`;
      return;
    }
    
    setSelectedCategory(categoryId === selectedCategory ? null : categoryId);
    if (category && categoryId !== selectedCategory) {
      setInputText(`[${category.name}] `);
    } else {
      setInputText('');
    }
  };

  return (
    <div className="home-page">
      <div className="chat-container">
        <div className="chat-header">
          <h1>Cleaners AI</h1>
          <p>AI 기반 세탁 지식 서비스</p>
          <div className="header-links">
            <a
              href={CHATGPT_LINK}
              target="_blank"
              rel="noopener noreferrer"
              className="chatgpt-link"
              data-testid="link-chatgpt"
            >
              ChatGPT 버전으로 이동
            </a>
            <a
              href="/admin/knowledge"
              className="admin-link"
              data-testid="link-admin"
            >
              관리자 페이지
            </a>
          </div>
        </div>

        <div className="category-selector">
          {CATEGORIES.map((category) => (
            <button
              key={category.id}
              className={`category-button ${selectedCategory === category.id ? 'selected' : ''}`}
              onClick={() => handleCategoryClick(category.id)}
              data-testid={`button-category-${category.id}`}
            >
              <span className="category-icon">{category.icon}</span>
              <span className="category-name">{category.name}</span>
            </button>
          ))}
        </div>

        <div className="chat-messages">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`message ${message.sender === 'user' ? 'user-message' : 'bot-message'}`}
            >
              <div className="message-content">
                {message.imageUrl && (
                  <img
                    src={message.imageUrl}
                    alt="Uploaded"
                    className="message-image"
                  />
                )}
                <p>{message.text}</p>
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
            type="file"
            ref={fileInputRef}
            onChange={handleFileSelect}
            accept="image/*"
            style={{ display: 'none' }}
          />
          <button
            type="button"
            className="attach-button"
            onClick={() => fileInputRef.current?.click()}
            disabled={isLoading}
            title="이미지 업로드"
          >
            📎
          </button>
          {selectedFile && (
            <span className="file-name">{selectedFile.name}</span>
          )}
          <input
            type="text"
            className="chat-input"
            placeholder="세탁에 대해 궁금한 점을 입력하세요..."
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

export default HomePage;
