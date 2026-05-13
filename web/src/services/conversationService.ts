import {
  API_BASE,
  getHeaders,
  getAuthHeaders,
  handleResponse,
} from './api';
import type {
  ConversationListResponse,
  ConversationDetailResponse,
  CreateConversationRequest,
} from '../types/api';
import type { Conversation } from '../types/models';

export async function getConversations(): Promise<Conversation[]> {
  const res = await fetch(`${API_BASE}/conversations`, {
    headers: getAuthHeaders(),
  });
  const data = await handleResponse<ConversationListResponse>(res);
  return data.conversations;
}

export async function createConversation(
  req: CreateConversationRequest
): Promise<Conversation> {
  const res = await fetch(`${API_BASE}/conversations`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(req),
  });
  return handleResponse<Conversation>(res);
}

export async function getConversationDetail(
  id: number
): Promise<ConversationDetailResponse> {
  const res = await fetch(`${API_BASE}/conversations/${id}`, {
    headers: getAuthHeaders(),
  });
  return handleResponse<ConversationDetailResponse>(res);
}

export function streamChat(
  convId: number,
  question: string,
  callbacks: {
    onMessage: (content: string) => void;
    onError: (error: string) => void;
    onComplete: () => void;
  }
): void {
  fetch(`${API_BASE}/conversations/${convId}/chat`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ question }),
  })
    .then(async (response) => {
      if (!response.ok) {
        callbacks.onError('HTTP error ' + response.status);
        return;
      }

      const reader = response.body!.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      function read(): void {
        reader
          .read()
          .then(({ done, value }) => {
            if (done) {
              callbacks.onComplete();
              return;
            }

            if (value) {
              buffer += decoder.decode(value, { stream: true });
              const lines = buffer.split('\n\n');
              buffer = lines.pop() || '';

              lines.forEach((line) => {
                if (line.startsWith('data: ')) {
                  const content = line.substring(6);
                  if (content.startsWith('[ERROR]')) {
                    callbacks.onError(content.substring(7));
                  } else if (!content.startsWith('[DONE]')) {
                    callbacks.onMessage(content);
                  }
                }
              });
            }

            if (!done) read();
          })
          .catch((err) => callbacks.onError(err.message));
      }

      read();
    })
    .catch((err) => callbacks.onError(err.message));
}

export async function deleteConversation(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/conversations/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  await handleResponse<{ message: string }>(res);
}
