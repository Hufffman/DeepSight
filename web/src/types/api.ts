import { KnowledgeBase, Conversation, Message, FileItem, AnalysisReport } from './models';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
}

export interface LoginResponse {
  token: string;
  user_id: number;
}

export interface ApiError {
  error: string;
}

export interface KnowledgeBaseListResponse {
  knowledge_bases: KnowledgeBase[];
}

export interface ConversationListResponse {
  conversations: Conversation[];
}

export interface ConversationDetailResponse {
  id: number;
  knowledge_base_id: number;
  title?: string;
  messages: Message[];
}

export interface FileListResponse {
  files: FileItem[];
}

export interface CreateKnowledgeBaseRequest {
  name: string;
  description?: string;
}

export interface CreateConversationRequest {
  knowledge_base_id: number;
  title?: string;
}

export interface UserResponse {
  id: number;
  username: string;
  email: string;
}

export interface UpdateUserRequest {
  username: string;
  email: string;
}

export interface UpdatePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface AnalysisReportListResponse {
  reports: AnalysisReport[];
  total: number;
  page: number;
  page_size: number;
}
