import { create } from 'zustand';
import { logout as apiLogout } from '../services/authService';

const AUTH_KEY = 'auth-storage';

function loadState(): { token: string | null; userId: number | null } {
  try {
    const stored = localStorage.getItem(AUTH_KEY);
    if (stored) {
      const { state } = JSON.parse(stored);
      return { token: state?.token || null, userId: state?.userId || null };
    }
  } catch { /* ignore */ }
  return { token: null, userId: null };
}

function saveState(token: string | null, userId: number | null) {
  try {
    localStorage.setItem(AUTH_KEY, JSON.stringify({ state: { token, userId } }));
  } catch { /* ignore */ }
}

interface AuthState {
  token: string | null;
  userId: number | null;
  isLoggedIn: boolean;
  login: (token: string, userId: number) => void;
  logout: () => void;
  setToken: (token: string) => void;
}

const initial = loadState();

export const authStore = create<AuthState>((set, get) => ({
  token: initial.token,
  userId: initial.userId,
  isLoggedIn: !!initial.token,
  login: (token: string, userId: number) => {
    saveState(token, userId);
    set({ token, userId, isLoggedIn: true });
  },
  logout: () => {
    const token = get().token;
    if (token) {
      apiLogout(token).catch(() => { /* ignore network errors */ });
    }
    saveState(null, null);
    set({ token: null, userId: null, isLoggedIn: false });
  },
  setToken: (token: string) => {
    saveState(token, null);
    set({ token });
  },
}));
