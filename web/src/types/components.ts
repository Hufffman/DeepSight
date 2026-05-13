import { Message, FileItem, KnowledgeBase, TabItem } from './models';

export interface ChatInputProps {
  disabled: boolean;
  onSend: (message: string) => void;
}

export interface MessageListProps {
  messages: Message[];
  loading: boolean;
  streamingContent: string;
  isStreaming: boolean;
}

export interface MessageItemProps {
  role: 'user' | 'assistant';
  content: string;
  isStreaming: boolean;
}

export interface FilePanelProps {
  kbId: number;
  files: FileItem[];
  loading: boolean;
  expanded: boolean;
  onToggle: () => void;
  onUpload: (file: File) => void;
  onDelete: (fileId: number) => void;
}

export interface FileListProps {
  files: FileItem[];
  onDelete: (fileId: number) => void;
}

export interface KbSelectorProps {
  knowledgeBases: KnowledgeBase[];
  currentKbId: number | null;
  loading: boolean;
  onChange: (kbId: number | null) => void;
  onCreateClick: () => void;
}

export interface CreateKbModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (name: string, description?: string) => void;
}

export interface SessionTabsProps {
  tabs: TabItem[];
  activeTabId: number | null;
  onSelect: (tabId: number) => void;
  onClose: (tabId: number) => void;
  onNew: () => void;
}

export interface LoginCardProps {
  onSubmit: (username: string, password: string) => void;
  loading: boolean;
  error: string | null;
}

export interface SidebarProps {
  onCreateKbClick: () => void;
}

export interface MobileDrawerProps {
  open: boolean;
  onClose: () => void;
  onCreateKbClick: () => void;
}

export interface UserMenuProps {
  userId: number | null;
  onLogout: () => void;
  onSettingsClick: () => void;
}

export interface ToastItem {
  id: number;
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
}

export interface SkeletonProps {
  className?: string;
}

export interface SettingsTabsProps {
  active: string;
  onChange: (tab: string) => void;
}
