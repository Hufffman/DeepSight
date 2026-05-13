import { create } from 'zustand';
import type { KnowledgeBase } from '../types/models';
import * as kbService from '../services/kbService';

interface KBState {
  list: KnowledgeBase[];
  currentKbId: number | null;
  loading: boolean;
  fetchList: () => Promise<void>;
  setCurrent: (kbId: number | null) => void;
  create: (name: string, description?: string) => Promise<KnowledgeBase>;
  update: (id: number, name: string, description?: string) => Promise<void>;
  delete: (id: number) => Promise<void>;
}

export const kbStore = create<KBState>((set, get) => ({
  list: [],
  currentKbId: null,
  loading: false,
  fetchList: async () => {
    set({ loading: true });
    try {
      const data = await kbService.getKnowledgeBases();
      set({ list: data, loading: false });
    } catch {
      set({ loading: false });
    }
  },
  setCurrent: (kbId: number | null) => set({ currentKbId: kbId }),
  create: async (name: string, description?: string) => {
    const kb = await kbService.createKnowledgeBase({ name, description });
    await get().fetchList();
    return kb;
  },
  update: async (id: number, name: string, description?: string) => {
    await kbService.updateKnowledgeBase(id, { name, description });
    await get().fetchList();
  },
  delete: async (id: number) => {
    await kbService.deleteKnowledgeBase(id);
    await get().fetchList();
  },
}));
