import { Plus, Library } from 'lucide-react';
import { kbStore } from '../../stores/kbStore';
import { conversationStore } from '../../stores/conversationStore';
import { uiStore } from '../../stores/uiStore';
import './Sidebar.scss';

interface SidebarProps {
  onCreateKbClick: () => void;
}

export function Sidebar({ onCreateKbClick }: SidebarProps) {
  const knowledgeBases = kbStore((s) => s.list);
  const currentKbId = kbStore((s) => s.currentKbId);
  const setCurrent = kbStore((s) => s.setCurrent);
  const fetchConversations = conversationStore((s) => s.fetchList);

  const closeSettings = uiStore((s) => s.setSettingsOpen);

  const handleSelect = (kbId: number) => {
    setCurrent(kbId);
    fetchConversations(kbId);
    closeSettings(false);
  };

  return (
    <aside className="sidebar">
      <div className="sidebar__header">
        <h3 className="sidebar__title">
          知识库
        </h3>
      </div>
      <div className="sidebar__list">
        {knowledgeBases.length === 0 ? (
          <p className="sidebar__empty">
            暂无知识库
          </p>
        ) : (
          knowledgeBases.map((kb) => (
            <button
              key={kb.id}
              onClick={() => handleSelect(kb.id)}
              className={`sidebar__item ${currentKbId === kb.id ? 'sidebar__item--active' : ''}`}
            >
              <div className="sidebar__item-name">
                <Library size={14} className="sidebar__item-icon" />
                <span className="sidebar__item-text">{kb.name}</span>
              </div>
              {kb.description && (
                <div className="sidebar__item-desc">
                  {kb.description}
                </div>
              )}
            </button>
          ))
        )}
      </div>
      <div className="sidebar__footer">
        <button
          onClick={onCreateKbClick}
          className="sidebar__create-btn"
        >
          <Plus size={14} />
          创建知识库
        </button>
      </div>
    </aside>
  );
}
