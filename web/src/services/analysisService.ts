import { API_BASE, getAuthHeaders, handleResponse } from './api';
import type { AnalysisReportListResponse } from '../types/api';
import type { AnalysisReport } from '../types/models';

export interface AnalysisTodo {
  id: number;
  title: string;
  intent: string;
  query: string;
}

export interface DeepAnalysisCallbacks {
  onStatus: (title: string) => void;
  onPlan: (todos: AnalysisTodo[]) => void;
  onTaskStart: (index: number, title: string) => void;
  onTaskCompleted: (index: number, title: string) => void;
  onReport: (content: string) => void;
  onError: (error: string) => void;
}

export function streamDeepAnalysis(kbId: number, convId: number, callbacks: DeepAnalysisCallbacks): void {
  fetch(`${API_BASE}/analysis/${kbId}?conv_id=${convId}`, {
    method: 'POST',
    headers: getAuthHeaders(),
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
                if (done) return;

                if (value) {
                  buffer += decoder.decode(value, { stream: true });
                  const lines = buffer.split('\n\n');
                  buffer = lines.pop() || '';

                  for (const line of lines) {
                    if (!line.startsWith('data: ')) continue;
                    const data = line.substring(6);
                    try {
                      const event = JSON.parse(data);
                      switch (event.type) {
                        case 'status':
                          callbacks.onStatus(event.title);
                          break;
                        case 'plan':
                          callbacks.onPlan(event.todos || []);
                          break;
                        case 'task_start':
                          callbacks.onTaskStart(event.index, event.title);
                          break;
                        case 'task_completed':
                          callbacks.onTaskCompleted(event.index, event.title);
                          break;
                        case 'report':
                          callbacks.onReport(event.content);
                          break;
                        case 'error':
                          callbacks.onError(event.content);
                          break;
                      }
                    } catch {
                      // skip unparseable events
                    }
                  }
                }

                read();
              })
              .catch((err) => callbacks.onError(err.message));
        }

        read();
      })
      .catch((err) => callbacks.onError(err.message));
}

export async function getAnalysisReports(): Promise<AnalysisReport[]> {
  const res = await fetch(`${API_BASE}/analysis/reports?page=1&page_size=200`, {
    headers: getAuthHeaders(),
  });
  const data = await handleResponse<AnalysisReportListResponse>(res);
  return data.reports;
}

export async function getAnalysisReport(id: number): Promise<AnalysisReport> {
  const res = await fetch(`${API_BASE}/analysis/reports/${id}`, {
    headers: getAuthHeaders(),
  });
  return handleResponse<AnalysisReport>(res);
}

export async function deleteAnalysisReport(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/analysis/reports/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  await handleResponse<{ message: string }>(res);
}
