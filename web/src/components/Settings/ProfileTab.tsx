import { useState, useEffect } from 'react';
import { User, UserCircle, Mail, Lock, Key } from 'lucide-react';
import { authStore } from '../../stores/authStore';
import { toastStore } from '../../stores/toastStore';
import * as userService from '../../services/userService';
import './ProfileTab.scss';

export function ProfileTab() {
  const userId = authStore((s) => s.userId);
  const show = toastStore((s) => s.show);

  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [infoLoading, setInfoLoading] = useState(false);
  const [pwLoading, setPwLoading] = useState(false);

  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  useEffect(() => {
    if (!userId) return;
    userService
      .getUser(userId)
      .then((u) => {
        setUsername(u.username);
        setEmail(u.email);
      })
      .catch(() => {});
  }, [userId]);

  const handleSaveInfo = async () => {
    if (!userId) return;
    setInfoLoading(true);
    try {
      await userService.updateUser(userId, { username, email });
      show('success', '个人信息已更新');
    } catch {
      // error handled by service layer
    }
    setInfoLoading(false);
  };

  const handleSavePassword = async () => {
    if (!userId) return;
    if (!oldPassword) {
      show('error', '请输入旧密码');
      return;
    }
    if (newPassword !== confirmPassword) {
      show('error', '两次输入的新密码不一致');
      return;
    }
    if (newPassword.length < 6) {
      show('error', '新密码至少需要6个字符');
      return;
    }
    setPwLoading(true);
    try {
      await userService.updatePassword(userId, { old_password: oldPassword, new_password: newPassword });
      show('success', '密码已更新');
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch {
      // error handled by service layer
    }
    setPwLoading(false);
  };

  return (
    <div className="profile-tab">
      <div className="profile-tab__grid">
        <div className="profile-card">
          <h3 className="profile-card__header">
            <User size={16} className="profile-card__header-icon" />
            基本信息
          </h3>
          <div className="profile-card__body">
            <div>
              <label className="profile-field__label">
                <UserCircle size={13} />
                用户名
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="profile-field__input"
              />
            </div>
            <div>
              <label className="profile-field__label">
                <Mail size={13} />
                邮箱
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="profile-field__input"
              />
            </div>
            <button
              onClick={handleSaveInfo}
              disabled={infoLoading}
              className="profile-card__submit"
            >
              {infoLoading ? '保存中...' : '保存修改'}
            </button>
          </div>
        </div>

        <div className="profile-card">
          <h3 className="profile-card__header">
            <Lock size={16} className="profile-card__header-icon" />
            修改密码
          </h3>
          <div className="profile-card__body">
            <div>
              <label className="profile-field__label">
                <Key size={13} />
                旧密码
              </label>
              <input
                type="password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                className="profile-field__input"
              />
            </div>
            <div>
              <label className="profile-field__label">
                新密码
              </label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="profile-field__input"
              />
            </div>
            <div>
              <label className="profile-field__label">
                确认新密码
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="profile-field__input"
              />
            </div>
            <button
              onClick={handleSavePassword}
              disabled={pwLoading}
              className="profile-card__submit"
            >
              {pwLoading ? '更新中...' : '更新密码'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
