import { Plus, X, MessageSquare } from 'lucide-react';
import type { SessionTabsProps } from '../../types/components';
import './SessionTabs.scss';

export function SessionTabs({
  tabs,
  activeTabId,
  onSelect,
  onClose,
  onNew,
}: SessionTabsProps) {
  return (
    <div className="session-tabs">
      <div className="session-tabs__list">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onSelect(tab.id)}
            className={`session-tabs__tab ${activeTabId === tab.id ? 'session-tabs__tab--active' : ''}`}
          >
            <MessageSquare size={13} className="session-tabs__tab-icon" />
            <span className="session-tabs__tab-title">{tab.title}</span>
            <span
              role="button"
              tabIndex={0}
              onClick={(e) => {
                e.stopPropagation();
                onClose(tab.id);
              }}
              className="session-tabs__tab-close"
            >
              <X size={12} />
            </span>
          </button>
        ))}
      </div>

      <button
        onClick={onNew}
        className="session-tabs__new-btn"
        title="新建会话"
      >
        <Plus size={16} />
      </button>
    </div>
  );
}
