import api from './api';

export interface ChatMessage {
  message: string;
  conversation_id?: string;
}

export interface ChatResponse {
  message: string;
  conversation_id: string;
  timestamp: string;
}

class ChatService {
  async sendMessage(message: string, conversationId?: string | null): Promise<ChatResponse> {
    const payload: any = { message };
    if (conversationId) {
      payload.conversation_id = conversationId;
    }

    const response = await api.post<ChatResponse>('/api/chat/message', payload);

    if (response.error) {
      throw new Error(response.error.message);
    }

    return response.data!;
  }
}

export default new ChatService();
