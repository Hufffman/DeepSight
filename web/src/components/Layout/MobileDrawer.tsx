import { X } from 'lucide-react';
import { kbStore } from '../../stores/kbStore';
import { conversationStore } from '../../stores/conversationStore';
import { uiStore } from '../../stores/uiStore';
import type { MobileDrawerProps } from '../../types/components';
import './MobileDrawer.scss';

export function MobileDrawer({ open, onClose, onCreateKbClick }: MobileDrawerProps) {
  const knowledgeBases = kbStore((s) => s.list);
  const currentKbId = kbStore((s) => s.currentKbId);
  const setCurrent = kbStore((s) => s.setCurrent);
  const fetchConversations = conversationStore((s) => s.fetchList);
  const closeSettings = uiStore((s) => s.setSettingsOpen);

  const handleSelect = (kbId: number) => {
    setCurrent(kbId);
    fetchConversations(kbId);
    closeSettings(false);
    onClose();
  };

  if (!open) return null;

  return (
    <div className="mobile-drawer">
      <div
        className="mobile-drawer__overlay"
        onClick={onClose}
      />
      <div className="mobile-drawer__panel">
        <div className="mobile-drawer__header">
          <h3 className="mobile-drawer__title">
            知识库
          </h3>
          <button
            onClick={onClose}
            className="mobile-drawer__close"
          >
            <X size={20} />
          </button>
        </div>
        <div className="mobile-drawer__body">
          {knowledgeBases.length === 0 ? (
            <p className="mobile-drawer__empty">
              暂无知识库
            </p>
          ) : (
            knowledgeBases.map((kb) => (
              <button
                key={kb.id}
                onClick={() => handleSelect(kb.id)}
                className={`mobile-drawer__item ${currentKbId === kb.id ? 'mobile-drawer__item--active' : ''}`}
              >
                <div className="mobile-drawer__item-name">{kb.name}</div>
                {kb.description && (
                  <div className="mobile-drawer__item-desc">
                    {kb.description}
                  </div>
                )}
              </button>
            ))
          )}
          <button
            onClick={() => {
              onCreateKbClick();
              onClose();
            }}
            className="mobile-drawer__create-btn"
          >
            + 创建知识库
          </button>
        </div>
      </div>
    </div>
  );
}
