import { API_BASE, getHeaders, getAuthHeaders, handleResponse } from './api';
import type { KnowledgeBaseListResponse, CreateKnowledgeBaseRequest } from '../types/api';
import type { KnowledgeBase } from '../types/models';

export async function getKnowledgeBases(): Promise<KnowledgeBase[]> {
  const res = await fetch(`${API_BASE}/knowledge-bases`, {
    headers: getAuthHeaders(),
  });
  const data = await handleResponse<KnowledgeBaseListResponse>(res);
  return data.knowledge_bases;
}

export async function createKnowledgeBase(req: CreateKnowledgeBaseRequest): Promise<KnowledgeBase> {
  const res = await fetch(`${API_BASE}/knowledge-bases`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(req),
  });
  return handleResponse<KnowledgeBase>(res);
}

export async function updateKnowledgeBase(
  id: number,
  req: { name: string; description?: string }
): Promise<KnowledgeBase> {
  const res = await fetch(`${API_BASE}/knowledge-bases/${id}`, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(req),
  });
  return handleResponse<KnowledgeBase>(res);
}

export async function deleteKnowledgeBase(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/knowledge-bases/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  await handleResponse<{ message: string }>(res);
}
