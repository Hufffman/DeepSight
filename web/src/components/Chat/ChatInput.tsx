import { useState, KeyboardEvent } from 'react';
import { Send } from 'lucide-react';
import type { ChatInputProps } from '../../types/components';
import './ChatInput.scss';

export function ChatInput({ disabled, onSend }: ChatInputProps) {
  const [text, setText] = useState('');

  const handleSend = () => {
    const trimmed = text.trim();
    if (!trimmed || disabled) return;
    setText('');
    onSend(trimmed);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="chat-input">
      <div className="chat-input__row">
        <input
          type="text"
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          placeholder={
            disabled ? '请先选择知识库和会话' : '输入消息... (Enter 发送)'
          }
          className="chat-input__field"
        />
        <button
          onClick={handleSend}
          disabled={disabled || !text.trim()}
          className="chat-input__send"
        >
          <Send size={18} />
        </button>
      </div>
    </div>
  );
}
