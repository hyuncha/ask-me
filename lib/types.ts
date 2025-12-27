// Chat Request/Response Types
export interface ChatRequest {
  message: string;
  zipcode?: string;
}

export interface ChatResponse {
  answer: string;
  success_rate?: string;
  risk_level?: 'low' | 'medium' | 'high';
  recommended_shops: PartnerShop[];
  disclaimer: string;
}

// Partner Shop Types
export interface PartnerShop {
  shop_name: string;
  zipcode: string;
  subscription: 'active' | 'inactive';
  specialty: string[];
  rating?: number;
}

// Message Types for UI
export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  recommended_shops?: PartnerShop[];
}

// Pinecone Types
export interface LaundryKnowledge {
  id: string;
  stain_type: string;
  fabric: string;
  success_rate: number;
  risk: 'low' | 'medium' | 'high';
  content: string;
}

export interface PineconeQueryResult {
  id: string;
  score: number;
  metadata: Record<string, unknown>;
}

// OpenRouter Types
export interface OpenRouterMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

export interface OpenRouterRequest {
  model: string;
  messages: OpenRouterMessage[];
  temperature?: number;
  max_tokens?: number;
}

export interface OpenRouterResponse {
  id: string;
  choices: {
    message: {
      role: string;
      content: string;
    };
    finish_reason: string;
  }[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

// Upsert Types
export interface UpsertResult {
  upsertedCount: number;
  errors: UpsertError[];
}

export interface UpsertError {
  id: string;
  error: string;
}

export interface PineconeVector {
  id: string;
  values: number[];
  metadata: Record<string, unknown>;
}

// Input types for upsert (id is optional, auto-generated if not provided)
export interface LaundryKnowledgeInput {
  id?: string;
  stain_type: string;
  fabric: string;
  success_rate: number;
  risk: 'low' | 'medium' | 'high';
  content: string;
}

export interface PartnerShopInput {
  id?: string;
  shop_name: string;
  zipcode: string;
  subscription: 'active' | 'inactive';
  specialty: string[];
  rating?: number;
}
