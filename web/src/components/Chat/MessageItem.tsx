import { MarkdownRenderer } from './MarkdownRenderer';
import type { MessageItemProps } from '../../types/components';
import './MessageItem.scss';

export function MessageItem({ role, content, isStreaming }: MessageItemProps) {
  const isUser = role === 'user';

  return (
    <div className={`message ${isUser ? 'message--user' : 'message--ai'}`}>
      <div className={`message__bubble ${isUser ? 'message__bubble--user' : 'message__bubble--ai'}`}>
        <div className="message__role">
          {isUser ? '你' : 'AI'}
          {isStreaming && !isUser && (
            <span className="message__streaming-dot" />
          )}
        </div>
        <div className="message__content">
          {content ? (
            isUser ? (
              <span className="message__text--user">{content}</span>
            ) : (
              <MarkdownRenderer content={content} />
            )
          ) : (
            isStreaming ? '思考中...' : ''
          )}
        </div>
      </div>
    </div>
  );
}
