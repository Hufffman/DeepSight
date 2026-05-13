import { useState } from 'react';
import { ArrowLeft } from 'lucide-react';
import { SettingsTabs } from './SettingsTabs';
import { ProfileTab } from './ProfileTab';
import { KbManageTab } from './KbManageTab';
import { ConversationTab } from './ConversationTab';
import { ReportTab } from './ReportTab';
import { uiStore } from '../../stores/uiStore';
import './SettingsPage.scss';

export function SettingsPage() {
  const [activeTab, setActiveTab] = useState('profile');
  const closeSettings = uiStore((s) => s.setSettingsOpen);

  return (
    <div className="settings-page">
      <div className="settings-page__top-bar">
        <button
          onClick={() => closeSettings(false)}
          className="settings-page__back-btn"
        >
          <ArrowLeft size={18} />
        </button>
        <h2 className="settings-page__title">设置</h2>
      </div>

      <SettingsTabs active={activeTab} onChange={setActiveTab} />

      {activeTab === 'profile' && <ProfileTab />}
      {activeTab === 'kb' && <KbManageTab />}
      {activeTab === 'conversations' && <ConversationTab />}
      {activeTab === 'reports' && <ReportTab />}
    </div>
  );
}
