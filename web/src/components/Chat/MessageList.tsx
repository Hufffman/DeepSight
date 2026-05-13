import { useEffect, useRef } from 'react';
import { MessageItem } from './MessageItem';
import { Skeleton } from '../common/Skeleton';
import type { MessageListProps } from '../../types/components';
import './MessageList.scss';

export function MessageList({
  messages,
  loading,
  streamingContent,
  isStreaming,
}: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, streamingContent]);

  if (loading) {
    return (
      <div className="message-list__loading">
        <div className="message-list__loading-row message-list__loading-row--end">
          <Skeleton className="message-list__skeleton message-list__skeleton--medium" />
        </div>
        <div className="message-list__loading-row message-list__loading-row--start">
          <Skeleton className="message-list__skeleton message-list__skeleton--large" />
        </div>
        <div className="message-list__loading-row message-list__loading-row--end">
          <Skeleton className="message-list__skeleton message-list__skeleton--small" />
        </div>
      </div>
    );
  }

  if (messages.length === 0 && !isStreaming) {
    return (
      <div className="message-list message-list--center">
        <div className="message-list__empty">
          <p className="message-list__empty-title">开始新的对话</p>
          <p className="message-list__empty-sub">在下方输入问题发送给 AI</p>
        </div>
      </div>
    );
  }

  return (
    <div className="message-list">
      {messages.map((msg, i) => (
        <MessageItem
          key={msg.id ?? i}
          role={msg.role}
          content={msg.content}
          isStreaming={false}
        />
      ))}
      {isStreaming && (
        <MessageItem
          role="assistant"
          content={streamingContent}
          isStreaming={true}
        />
      )}
      <div ref={bottomRef} />
    </div>
  );
}
