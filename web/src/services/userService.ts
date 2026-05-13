import { API_BASE, getHeaders, getAuthHeaders, handleResponse } from './api';
import type { UserResponse, UpdateUserRequest, UpdatePasswordRequest } from '../types/api';

export async function getUser(id: number): Promise<UserResponse> {
  const res = await fetch(`${API_BASE}/users/${id}`, {
    headers: getAuthHeaders(),
  });
  return handleResponse<UserResponse>(res);
}

export async function updateUser(id: number, req: UpdateUserRequest): Promise<UserResponse> {
  const res = await fetch(`${API_BASE}/users/${id}`, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(req),
  });
  return handleResponse<UserResponse>(res);
}

export async function updatePassword(id: number, req: UpdatePasswordRequest): Promise<void> {
  const res = await fetch(`${API_BASE}/users/${id}/password`, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(req),
  });
  await handleResponse<{ message: string }>(res);
}
