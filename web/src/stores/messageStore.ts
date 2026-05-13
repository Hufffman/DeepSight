import { create } from 'zustand';
import type { Message } from '../types/models';
import * as convService from '../services/conversationService';

interface MessageState {
  messagesByConv: Record<number, Message[]>;
  streamingContent: string;
  isStreaming: boolean;
  loadingMessages: boolean;
  fetchMessages: (convId: number) => Promise<void>;
  addMessage: (convId: number, message: Message) => void;
  setStreamingContent: (content: string) => void;
  setStreaming: (streaming: boolean) => void;
  clearConvMessages: (convId: number) => void;
}

export const messageStore = create<MessageState>((set) => ({
  messagesByConv: {},
  streamingContent: '',
  isStreaming: false,
  loadingMessages: false,
  fetchMessages: async (convId: number) => {
    set({ loadingMessages: true });
    try {
      const data = await convService.getConversationDetail(convId);
      set((state) => ({
        messagesByConv: { ...state.messagesByConv, [convId]: data.messages },
        loadingMessages: false,
      }));
    } catch {
      set({ loadingMessages: false });
    }
  },
  addMessage: (convId: number, message: Message) => {
    set((state) => ({
      messagesByConv: {
        ...state.messagesByConv,
        [convId]: [...(state.messagesByConv[convId] || []), message],
      },
    }));
  },
  setStreamingContent: (content: string) => set({ streamingContent: content }),
  setStreaming: (streaming: boolean) => set({ isStreaming: streaming }),
  clearConvMessages: (convId: number) => {
    set((state) => {
      const next = { ...state.messagesByConv };
      delete next[convId];
      return { messagesByConv: next };
    });
  },
}));
