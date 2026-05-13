import { ReactNode } from 'react';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { MobileDrawer } from './MobileDrawer';
import { uiStore } from '../../stores/uiStore';
import { kbStore } from '../../stores/kbStore';
import './Layout.scss';

interface LayoutProps {
  children: ReactNode;
  onCreateKbClick: () => void;
  onCreateConvClick: () => void;
  onSettingsClick: () => void;
  settingsOpen: boolean;
}

export function Layout({ children, onCreateKbClick, onCreateConvClick, onSettingsClick, settingsOpen }: LayoutProps) {
  const mobileDrawerOpen = uiStore((s) => s.mobileDrawerOpen);
  const setMobileDrawer = uiStore((s) => s.setMobileDrawerOpen);
  const currentKbId = kbStore((s) => s.currentKbId);

  return (
    <div className="app-shell">
      <Header
        onCreateConvClick={onCreateConvClick}
        onSettingsClick={onSettingsClick}
      />
      <MobileDrawer
        open={mobileDrawerOpen}
        onClose={() => setMobileDrawer(false)}
        onCreateKbClick={onCreateKbClick}
      />
      <div className="app-body">
        <div className="app-sidebar">
          <Sidebar onCreateKbClick={onCreateKbClick} />
        </div>
        <main className="app-main">
          {currentKbId || settingsOpen ? children : (
            <div className="app-placeholder">
              <p>请选择或创建一个知识库开始</p>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
