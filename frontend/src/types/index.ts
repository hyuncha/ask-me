// User types
export enum UserRole {
  CONSUMER = 'consumer',
  OWNER = 'owner',
  ADMIN = 'admin',
}

export interface User {
  id: string;
  email: string;
  name: string;
  picture: string;
  role: UserRole;
  language: 'KR' | 'EN';
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
}

// Knowledge types
export enum KnowledgeCategory {
  STAIN_REMOVAL = 'stain_removal',
  FABRIC_UNDERSTANDING = 'fabric_understanding',
  ACCIDENT_PREVENTION = 'accident_prevention',
  LAUNDRY_TECHNIQUE = 'laundry_technique',
  EQUIPMENT_OPERATION = 'equipment_operation',
  MARKETING = 'marketing',
  OTHERS = 'others',
}

export enum KnowledgeDifficulty {
  BASIC = 'basic',
  EXPERT = 'expert',
}

export interface Knowledge {
  id: string;
  title: string;
  content: string;
  category: KnowledgeCategory;
  difficulty: KnowledgeDifficulty;
  tags: string[];
  language: 'KR' | 'EN';
  status: 'active' | 'inactive' | 'draft';
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

// Conversation types
export enum MessageRole {
  USER = 'user',
  ASSISTANT = 'assistant',
  SYSTEM = 'system',
}

export interface Message {
  id: string;
  conversationId: string;
  role: MessageRole;
  content: string;
  createdAt: string;
}

export interface Conversation {
  id: string;
  userId: string;
  title: string;
  language: 'KR' | 'EN';
  createdAt: string;
  updatedAt: string;
}

// Subscription types
export enum SubscriptionPlan {
  FREE = 'free',
  MONTHLY = 'monthly',
  YEARLY = 'yearly',
}

export interface Subscription {
  id: string;
  userId: string;
  plan: SubscriptionPlan;
  status: 'active' | 'canceled' | 'expired';
  currentPeriodEnd?: string;
  createdAt: string;
  updatedAt: string;
}

// API Response types
export interface ApiResponse<T> {
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}
