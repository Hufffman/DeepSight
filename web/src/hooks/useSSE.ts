import { useCallback, useRef } from 'react';

interface SSECallbacks {
  onMessage: (content: string) => void;
  onError: (error: string) => void;
  onComplete: () => void;
}

export function useSSE() {
  const abortRef = useRef<AbortController | null>(null);

  const connect = useCallback(
    (url: string, body: unknown, callbacks: SSECallbacks) => {
      const controller = new AbortController();
      abortRef.current = controller;

      fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
        signal: controller.signal,
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
                      } else if (content.startsWith('[DONE]')) {
                        callbacks.onComplete();
                      } else {
                        callbacks.onMessage(content);
                      }
                    }
                  });
                }

                if (!done) read();
              })
              .catch((err) => {
                if (err.name !== 'AbortError') {
                  callbacks.onError(err.message);
                }
              });
          }

          read();
        })
        .catch((err) => {
          if (err.name !== 'AbortError') {
            callbacks.onError(err.message);
          }
        });
    },
    []
  );

  const abort = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  return { connect, abort };
}
