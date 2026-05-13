/// <reference types="vite/client" />

import { toastStore } from '../stores/toastStore';

export const API_BASE = import.meta.env.VITE_API_BASE || '/api/v1';

export function getHeaders(): HeadersInit {
  const headers: HeadersInit = { 'Content-Type': 'application/json' };
  try {
    const stored = localStorage.getItem('auth-storage');
    if (stored) {
      const { state } = JSON.parse(stored);
      if (state?.token) {
        headers['Authorization'] = `Bearer ${state.token}`;
      }
    }
  } catch {
    // parse error, skip auth header
  }
  return headers;
}

export function getAuthHeaders(): HeadersInit {
  const headers: HeadersInit = {};
  try {
    const stored = localStorage.getItem('auth-storage');
    if (stored) {
      const { state } = JSON.parse(stored);
      if (state?.token) {
        headers['Authorization'] = `Bearer ${state.token}`;
      }
    }
  } catch {
    // parse error, skip auth header
  }
  return headers;
}

export class ApiException extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ApiException';
  }
}

export async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const data = await response.json().catch(() => ({ error: '请求失败' }));
    const message = data?.error || `HTTP ${response.status}`;
    toastStore.getState().show('error', message);
    throw new ApiException(message);
  }
  const text = await response.text();
  if (!text) {
    throw new ApiException('服务器返回空响应');
  }
  return JSON.parse(text) as T;
}
