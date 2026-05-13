import { API_BASE, getAuthHeaders, handleResponse } from './api';
import type { FileListResponse } from '../types/api';
import type { FileItem } from '../types/models';

export async function getFiles(kbId: number): Promise<FileItem[]> {
  const res = await fetch(`${API_BASE}/knowledge-bases/${kbId}/files`, {
    headers: getAuthHeaders(),
  });
  const data = await handleResponse<FileListResponse>(res);
  return data.files;
}

export async function uploadFile(kbId: number, file: File): Promise<FileItem> {
  const formData = new FormData();
  formData.append('file', file);

  const res = await fetch(`${API_BASE}/knowledge-bases/${kbId}/files`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: formData,
  });
  return handleResponse<FileItem>(res);
}

export async function deleteFile(kbId: number, fileId: number): Promise<void> {
  const res = await fetch(`${API_BASE}/knowledge-bases/${kbId}/files/${fileId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: '删除失败' }));
    throw new Error(data?.error || `HTTP ${res.status}`);
  }
}
