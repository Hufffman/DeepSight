import { useState, useEffect, useMemo } from 'react';
import { Trash2, MessageSquare, Layers } from 'lucide-react';
import { kbStore } from '../../stores/kbStore';
import { conversationStore } from '../../stores/conversationStore';
import { toastStore } from '../../stores/toastStore';
import { uiStore } from '../../stores/uiStore';
import { Skeleton } from '../common/Skeleton';
import { formatDate } from '../../utils/format';
import * as convService from '../../services/conversationService';
import type { Conversation } from '../../types/models';
import './ConversationTab.scss';

export function ConversationTab() {
  const knowledgeBases = kbStore((s) => s.list);
  const setCurrentKb = kbStore((s) => s.setCurrent);
  const openTab = conversationStore((s) => s.openTab);
  const fetchList = conversationStore((s) => s.fetchList);
  const deleteConv = conversationStore((s) => s.deleteConv);
  const closeSettings = uiStore((s) => s.setSettingsOpen);
  const show = toastStore((s) => s.show);

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState<Conversation | null>(null);

  useEffect(() => {
    setLoading(true);
    convService
      .getConversations()
      .then(setConversations)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const grouped = useMemo(() => {
    const map: Record<number, { kbName: string; convs: Conversation[] }> = {};
    for (const conv of conversations) {
      if (!map[conv.knowledge_base_id]) {
        const kb = knowledgeBases.find((k) => k.id === conv.knowledge_base_id);
        map[conv.knowledge_base_id] = {
          kbName: kb?.name || `知识库 #${conv.knowledge_base_id}`,
          convs: [],
        };
      }
      map[conv.knowledge_base_id].convs.push(conv);
    }
    return map;
  }, [conversations, knowledgeBases]);

  const handleJump = async (conv: Conversation) => {
    setCurrentKb(conv.knowledge_base_id);
    await fetchList(conv.knowledge_base_id);
    openTab(conv);
    closeSettings(false);
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteConv(deleteTarget.id);
      show('success', `已删除「${deleteTarget.title || `会话 #${deleteTarget.id}`}」`);
      setConversations((prev) => prev.filter((c) => c.id !== deleteTarget.id));
      setDeleteTarget(null);
    } catch {
      // handled by service layer
    }
  };

  if (loading) {
    return (
      <div className="conv-tab__loading">
        <div className="conv-tab__loading-group">
          <Skeleton className="conv-tab__loading-title" />
          <Skeleton className="conv-tab__loading-row" />
          <Skeleton className="conv-tab__loading-row" />
        </div>
        <div className="conv-tab__loading-group" style={{ marginTop: '24px' }}>
          <Skeleton className="conv-tab__loading-title" />
          <Skeleton className="conv-tab__loading-row" />
          <Skeleton className="conv-tab__loading-row" />
        </div>
      </div>
    );
  }

  const kbIds = Object.keys(grouped).map(Number);

  return (
    <div className="conv-tab">
      <div className="conv-tab__summary">
        <p className="conv-tab__summary-text">
          共 {conversations.length} 个会话，{kbIds.length} 个知识库
        </p>
      </div>

      {kbIds.length === 0 ? (
        <div className="conv-tab__empty">
          <MessageSquare size={48} className="conv-tab__empty-icon" />
          <p>暂无会话</p>
        </div>
      ) : (
        <div className="conv-tab__groups">
          {kbIds.map((kbId) => (
            <div key={kbId}>
              <div className="conv-group__header">
                <Layers size={16} className="conv-group__icon" />
                <h3 className="conv-group__name">
                  {grouped[kbId].kbName}
                </h3>
                <span className="conv-group__count">
                  {grouped[kbId].convs.length} 个会话
                </span>
              </div>
              <div className="conv-table">
                <table style={{ width: '100%', fontSize: '14px' }}>
                  <thead className="conv-table__head">
                    <tr>
                      <th className="conv-table__th conv-table__th--left">会话名称</th>
                      <th className="conv-table__th conv-table__th--right">创建时间</th>
                      <th className="conv-table__th conv-table__th--center">删除</th>
                    </tr>
                  </thead>
                  <tbody className="conv-table__body">
                    {grouped[kbId].convs.map((conv) => (
                      <tr key={conv.id} className="conv-table__row">
                        <td className="conv-table__td conv-table__td--left">
                          <button
                            onClick={() => handleJump(conv)}
                            className="conv-table__link"
                          >
                            <MessageSquare size={13} className="conv-table__link-icon" />
                            {conv.title || `会话 #${conv.id}`}
                          </button>
                        </td>
                        <td className="conv-table__td conv-table__td--right">
                          <span className="conv-table__date">{formatDate(conv.created_at || '')}</span>
                        </td>
                        <td className="conv-table__td conv-table__td--center">
                          <button
                            onClick={() => setDeleteTarget(conv)}
                            className="conv-table__delete-btn"
                            title="删除"
                          >
                            <Trash2 size={14} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </div>
      )}

      {deleteTarget && (
        <div className="conv-confirm-overlay">
          <div className="conv-confirm-backdrop" onClick={() => setDeleteTarget(null)} />
          <div className="conv-confirm">
            <h3 className="conv-confirm__title">
              确认删除
            </h3>
            <p className="conv-confirm__msg">
              确定删除「{deleteTarget.title || `会话 #${deleteTarget.id}`}」？
            </p>
            <div className="conv-confirm__actions">
              <button
                onClick={() => setDeleteTarget(null)}
                className="conv-confirm__cancel"
              >
                取消
              </button>
              <button
                onClick={handleDelete}
                className="conv-confirm__danger"
              >
                确认删除
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
