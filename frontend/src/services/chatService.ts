import api from './api';

export interface ChatMessage {
  message: string;
  conversation_id?: string;
  session_id?: string;
  location?: string;
}

export interface RecommendedShop {
  name: string;
  zipcode: string;
  priority: string;
  rating?: number;
  specialties?: string[];
}

export interface ChatResponse {
  message: string;
  conversation_id: string;
  timestamp: string;
  recommended_shops?: RecommendedShop[];
}

class ChatService {
  private sessionId: string | null = null;

  getSessionId(): string {
    if (!this.sessionId) {
      this.sessionId = `session-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    }
    return this.sessionId;
  }

  resetSession(): void {
    this.sessionId = null;
  }

  async sendMessage(
    message: string,
    conversationId?: string | null,
    location?: string
  ): Promise<ChatResponse> {
    const payload: any = {
      message,
      session_id: this.getSessionId()
    };

    if (conversationId) {
      payload.conversation_id = conversationId;
    }

    if (location) {
      payload.location = location;
    }

    const response = await api.post<ChatResponse>('/api/chat/message', payload);

    if (response.error) {
      throw new Error(response.error.message);
    }

    return response.data!;
  }

  async sendLaundryQuestion(
    message: string,
    location?: string
  ): Promise<ChatResponse> {
    const payload: any = {
      message,
      session_id: this.getSessionId()
    };

    if (location) {
      payload.location = location;
    }

    const response = await api.post<ChatResponse>('/api/chat/message', payload);

    if (response.error) {
      throw new Error(response.error.message);
    }

    return response.data!;
  }
}

export default new ChatService();
