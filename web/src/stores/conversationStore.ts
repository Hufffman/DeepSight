import { create } from 'zustand';
import type { Conversation, TabItem } from '../types/models';
import * as convService from '../services/conversationService';

interface ConversationState {
  list: Conversation[];
  tabs: TabItem[];
  activeTabId: number | null;
  loading: boolean;
  messagesByConv: Record<number, any[]>;
  fetchList: (kbId?: number) => Promise<void>;
  openTab: (conv: Conversation) => void;
  closeTab: (tabId: number) => void;
  setActiveTab: (tabId: number | null) => void;
  create: (kbId: number) => Promise<Conversation>;
  deleteConv: (convId: number) => Promise<void>;
  clear: () => void;
}

export const conversationStore = create<ConversationState>((set, get) => ({
  list: [],
  tabs: [],
  activeTabId: null,
  loading: false,
  messagesByConv: {},
  fetchList: async (kbId?: number) => {
    set({ loading: true });
    try {
      const list = await convService.getConversations();
      set({ list, loading: false });
      if (kbId) {
        const filtered = list.filter((c) => c.knowledge_base_id === kbId);
        const newTabs: TabItem[] = filtered.map((c) => ({
          id: c.id,
          title: c.title || `会话 #${c.id}`,
          kbId: c.knowledge_base_id,
        }));
        set({
          tabs: newTabs,
          activeTabId: newTabs.length > 0 ? newTabs[0].id : null,
        });
      }
    } catch {
      set({ loading: false });
    }
  },
  openTab: (conv: Conversation) => {
    const { tabs } = get();
    const exists = tabs.find((t) => t.id === conv.id);
    if (!exists) {
      set({
        tabs: [
          ...tabs,
          {
            id: conv.id,
            title: conv.title || `会话 #${conv.id}`,
            kbId: conv.knowledge_base_id,
          },
        ],
      });
    }
    set({ activeTabId: conv.id });
  },
  closeTab: (tabId: number) => {
    const { tabs, activeTabId } = get();
    const newTabs = tabs.filter((t) => t.id !== tabId);
    if (activeTabId === tabId) {
      set({
        tabs: newTabs,
        activeTabId: newTabs.length > 0 ? newTabs[newTabs.length - 1].id : null,
      });
    } else {
      set({ tabs: newTabs });
    }
  },
  setActiveTab: (tabId: number | null) => set({ activeTabId: tabId }),
  create: async (kbId: number) => {
    const conv = await convService.createConversation({
      knowledge_base_id: kbId,
      title: '',
    });
    await get().fetchList();
    get().openTab(conv);
    return conv;
  },
  deleteConv: async (convId: number) => {
    await convService.deleteConversation(convId);
    const { tabs, activeTabId } = get();
    const newTabs = tabs.filter((t) => t.id !== convId);
    const newActiveTabId = activeTabId === convId
      ? (newTabs.length > 0 ? newTabs[newTabs.length - 1].id : null)
      : activeTabId;
    set({ tabs: newTabs, activeTabId: newActiveTabId });
    set((state) => {
      const next = { ...state.messagesByConv };
      delete next[convId];
      return { messagesByConv: next };
    });
  },
  clear: () => set({ list: [], tabs: [], activeTabId: null, loading: false, messagesByConv: {} }),
}));
