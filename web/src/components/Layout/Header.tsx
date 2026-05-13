import { Menu, BookOpen } from 'lucide-react';
import { SessionTabs } from './SessionTabs';
import { ThemeToggle } from '../common/ThemeToggle';
import { UserMenu } from '../common/UserMenu';
import { kbStore } from '../../stores/kbStore';
import { conversationStore } from '../../stores/conversationStore';
import { uiStore } from '../../stores/uiStore';
import { authStore } from '../../stores/authStore';
import './Header.scss';

interface HeaderProps {
  onCreateConvClick: () => void;
  onSettingsClick: () => void;
}

export function Header({ onCreateConvClick, onSettingsClick }: HeaderProps) {
  const currentKbId = kbStore((s) => s.currentKbId);
  const tabs = conversationStore((s) => s.tabs);
  const activeTabId = conversationStore((s) => s.activeTabId);
  const setActiveTab = conversationStore((s) => s.setActiveTab);
  const closeTab = conversationStore((s) => s.closeTab);
  const setMobileDrawer = uiStore((s) => s.setMobileDrawerOpen);
  const closeSettings = uiStore((s) => s.setSettingsOpen);
  const userId = authStore((s) => s.userId);
  const logout = authStore((s) => s.logout);

  return (
    <header className="header">
      <div className="header__row">
        <button
          onClick={() => setMobileDrawer(true)}
          className="header__menu-btn"
        >
          <Menu size={20} />
        </button>

        <div className="header__brand">
          <BookOpen size={22} />
          <span className="header__brand-text">DeepSight</span>
        </div>

        {currentKbId && (
          <div className="header__tabs">
            <SessionTabs
              tabs={tabs}
              activeTabId={activeTabId}
              onSelect={(id) => {
                setActiveTab(id);
                closeSettings(false);
              }}
              onClose={closeTab}
              onNew={onCreateConvClick}
            />
          </div>
        )}

        <div className="header__actions">
          <ThemeToggle />
          <UserMenu userId={userId} onLogout={logout} onSettingsClick={onSettingsClick} />
        </div>
      </div>
    </header>
  );
}
