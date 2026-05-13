import { create } from 'zustand';
import type { ToastItem } from '../types/components';

interface ToastState {
  toasts: ToastItem[];
  show: (type: ToastItem['type'], message: string) => void;
  dismiss: (id: number) => void;
}

let toastId = 0;

export const toastStore = create<ToastState>((set) => ({
  toasts: [],
  show: (type, message) => {
    const id = ++toastId;
    set((state) => ({
      toasts:
        state.toasts.length >= 3
          ? [...state.toasts.slice(1), { id, type, message }]
          : [...state.toasts, { id, type, message }],
    }));
    setTimeout(() => {
      set((state) => ({
        toasts: state.toasts.filter((t) => t.id !== id),
      }));
    }, 3000);
  },
  dismiss: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    })),
}));
