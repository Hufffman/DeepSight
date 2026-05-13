import type { SettingsTabsProps } from '../../types/components';
import './SettingsTabs.scss';

const TABS = [
  { key: 'profile', label: '个人信息' },
  { key: 'kb', label: '知识库管理' },
  { key: 'conversations', label: '会话管理' },
  { key: 'reports', label: '分析报告' },
];

export function SettingsTabs({ active, onChange }: SettingsTabsProps) {
  return (
    <div className="settings-tabs">
      {TABS.map((tab) => (
        <button
          key={tab.key}
          onClick={() => onChange(tab.key)}
          className={`settings-tabs__tab ${active === tab.key ? 'settings-tabs__tab--active' : ''}`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
