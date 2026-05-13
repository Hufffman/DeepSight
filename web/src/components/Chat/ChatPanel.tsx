import { useCallback, useState } from 'react';
import { MessageList } from './MessageList';
import { ChatInput } from './ChatInput';
import { messageStore } from '../../stores/messageStore';
import { conversationStore } from '../../stores/conversationStore';
import { streamChat } from '../../services/conversationService';
import { streamDeepAnalysis } from '../../services/analysisService.ts';
import { kbStore } from '../../stores/kbStore';
import './ChatPanel.scss';

export function ChatPanel() {
  const activeTabId = conversationStore((s) => s.activeTabId);
  const messagesByConv = messageStore((s) => s.messagesByConv);
  const loadingMessages = messageStore((s) => s.loadingMessages);
  const streamingContent = messageStore((s) => s.streamingContent);
  const isStreaming = messageStore((s) => s.isStreaming);
  const setStreamingContent = messageStore((s) => s.setStreamingContent);
  const setStreaming = messageStore((s) => s.setStreaming);
  const addMessage = messageStore((s) => s.addMessage);
  const currentKbId = kbStore((s) => s.currentKbId);

  const messages = activeTabId ? (messagesByConv[activeTabId] ?? []) : [];

  const handleSend = useCallback(
    (question: string) => {
      if (!activeTabId) return;

      addMessage(activeTabId, {
        role: 'user',
        content: question,
      });

      setStreamingContent('');
      setStreaming(true);
      let fullContent = '';

      streamChat(activeTabId, question, {
        onMessage: (content: string) => {
          fullContent += content;
          setStreamingContent(fullContent);
        },
        onError: (error: string) => {
          setStreaming(false);
          addMessage(activeTabId, {
            role: 'assistant',
            content: `错误: ${error}`,
          });
        },
        onComplete: () => {
          setStreaming(false);
          setStreamingContent('');
          if (fullContent) {
            addMessage(activeTabId, {
              role: 'assistant',
              content: fullContent,
            });
          }
        },
      });
    },
    [activeTabId, addMessage, setStreamingContent, setStreaming]
  );

  const [isDeepAnalyzing, setIsDeepAnalyzing] = useState(false);

  const handleDeepAnalysis = async () => {
    if (!currentKbId || !activeTabId) return;
    setIsDeepAnalyzing(true);

    let progress = '\u{1F52C} 深度分析已启动...\n';
    setStreamingContent(progress);
    setStreaming(true);

    streamDeepAnalysis(currentKbId, activeTabId, {
      onStatus: (title) => {
        progress += `⏳ ${title}\n`;
        setStreamingContent(progress);
      },
      onPlan: (todos) => {
        progress += '\n## \u{1F4CB} 研究计划\n';
        todos.forEach((t) => {
          progress += `- \u{1F4CC} **${t.title}**: ${t.intent}\n`;
        });
        progress += '\n';
        setStreamingContent(progress);
      },
      onTaskStart: (index, title) => {
        progress += `⏳ [${index}/5] ${title} 执行中...\n`;
        setStreamingContent(progress);
      },
      onTaskCompleted: (index, title) => {
        progress += `✅ [${index}/5] ${title} 完成\n`;
        setStreamingContent(progress);
      },
      onReport: (content) => {
        setStreaming(false);
        setStreamingContent('');
        addMessage(activeTabId, {
          role: 'assistant',
          content: `## \u{1F52C} 深度能力分析报告\n\n${content}`,
          created_at: new Date().toISOString(),
        });
        setIsDeepAnalyzing(false);
      },
      onError: (error) => {
        setStreaming(false);
        setStreamingContent('');
        addMessage(activeTabId, {
          role: 'assistant',
          content: `❌ 深度分析失败: ${error}`,
          created_at: new Date().toISOString(),
        });
        setIsDeepAnalyzing(false);
      },
    });
  };

  return (
    <div className="chat-panel">
      {activeTabId && (
        <div style={{
          padding: '12px 16px',
          borderBottom: '1px solid var(--border-color)',
          display: 'flex',
          gap: '8px',
          alignItems: 'center',
          background: 'linear-gradient(135deg, rgba(118,75,162,0.05), rgba(102,126,234,0.05))'
        }}>
          <button
            onClick={handleDeepAnalysis}
            disabled={isDeepAnalyzing}
            style={{
              padding: '6px 14px',
              background: isDeepAnalyzing ? '#ccc' : 'linear-gradient(135deg, #667eea, #764ba2)',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: isDeepAnalyzing ? 'not-allowed' : 'pointer',
              fontSize: '13px',
              fontWeight: 600
            }}
          >
            {isDeepAnalyzing ? '深度分析中...' : '\u{1F52C} 深度分析'}
          </button>
          <span style={{ fontSize: '12px', color: '#999' }}>
            项目分析 + 能力评估 + 发展建议 (含联网搜索)
          </span>
        </div>
      )}
      <MessageList
        messages={messages}
        loading={loadingMessages}
        streamingContent={streamingContent}
        isStreaming={isStreaming}
      />
      <ChatInput disabled={!activeTabId || isStreaming} onSend={handleSend} />
    </div>
  );
}
