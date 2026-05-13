import { useState, useRef, useEffect } from 'react';
import { LogOut, User, Settings } from 'lucide-react';
import * as userService from '../../services/userService';
import type { UserMenuProps } from '../../types/components';
import './UserMenu.scss';

export function UserMenu({ userId, onLogout, onSettingsClick }: UserMenuProps) {
  const [open, setOpen] = useState(false);
  const [username, setUsername] = useState<string | null>(null);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  useEffect(() => {
    if (userId !== null) {
      userService.getUser(userId).then((u) => setUsername(u.username)).catch(() => {});
    } else {
      setUsername(null);
    }
  }, [userId]);

  if (userId === null) {
    return (
      <div className="user-menu--guest">
        <User size={16} />
        <span>未登录</span>
      </div>
    );
  }

  return (
    <div className="user-menu" ref={ref}>
      <button
        onClick={() => setOpen(!open)}
        className="user-menu__trigger"
      >
        <User size={16} />
        <span className="user-menu__name">{username || `用户 #${userId}`}</span>
      </button>

      {open && (
        <div className="user-menu__dropdown">
          <div className="user-menu__info">
            {username || `用户 ID: ${userId}`}
          </div>
          <button
            onClick={() => {
              setOpen(false);
              onSettingsClick();
            }}
            className="user-menu__action"
          >
            <Settings size={14} />
            设置
          </button>
          <button
            onClick={() => {
              setOpen(false);
              onLogout();
            }}
            className="user-menu__action user-menu__action--danger"
          >
            <LogOut size={14} />
            退出登录
          </button>
        </div>
      )}
    </div>
  );
}
