export interface KnowledgeBase {
  id: number;
  name: string;
  description?: string;
  created_at?: string;
  file_count?: number;
}

export interface Conversation {
  id: number;
  knowledge_base_id: number;
  title?: string;
  created_at?: string;
  message_count?: number;
}

export interface Message {
  id?: number;
  role: 'user' | 'assistant';
  content: string;
  created_at?: string;
}

export interface FileItem {
  id: number;
  file_name: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
}

export interface TabItem {
  id: number;
  title: string;
  kbId: number;
}

export interface AnalysisReport {
  id: number;
  knowledge_base_id: number;
  report_type: string;
  content: string;
  created_at: string;
}
