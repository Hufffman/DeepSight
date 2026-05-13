import { useState, FormEvent } from 'react';
import './LoginCard.scss';

interface LoginCardProps {
  onLogin: (username: string, password: string) => void;
  onRegister: (username: string, password: string, email: string) => void;
  loading: boolean;
  error: string | null;
}

export function LoginCard({ onLogin, onRegister, loading, error }: LoginCardProps) {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [email, setEmail] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const reset = () => {
    setUsername('');
    setPassword('');
    setConfirmPassword('');
    setEmail('');
    setLocalError(null);
  };

  const switchMode = () => {
    reset();
    setMode(mode === 'login' ? 'register' : 'login');
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    setLocalError(null);

    if (!username.trim() || !password.trim()) {
      setLocalError('请填写用户名和密码');
      return;
    }

    if (mode === 'register') {
      if (!email.trim()) {
        setLocalError('请填写邮箱');
        return;
      }
      if (password.length < 6) {
        setLocalError('密码至少需要6个字符');
        return;
      }
      if (password !== confirmPassword) {
        setLocalError('两次输入的密码不一致');
        return;
      }
      onRegister(username, password, email);
    } else {
      onLogin(username, password);
    }
  };

  const displayError = localError || error;

  return (
    <div className="login-page">
      <div className="login-card">
        <h1 className="login-card__title">
          DeepSight
        </h1>

        <form onSubmit={handleSubmit} className="login-card__form">
          <div>
            <label className="login-card__label">
              用户名
            </label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="login-card__input"
              placeholder="输入用户名"
              required
            />
          </div>

          {mode === 'register' && (
            <div>
              <label className="login-card__label">
                邮箱
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="login-card__input"
                placeholder="输入邮箱"
                required
              />
            </div>
          )}

          <div>
            <label className="login-card__label">
              密码
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="login-card__input"
              placeholder={mode === 'register' ? '至少6个字符' : '输入密码'}
              required
            />
          </div>

          {mode === 'register' && (
            <div>
              <label className="login-card__label">
                确认密码
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="login-card__input"
                placeholder="再次输入密码"
                required
              />
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="login-card__submit"
          >
            {loading ? (mode === 'register' ? '注册中...' : '登录中...') : (mode === 'register' ? '注册' : '登录')}
          </button>

          {displayError && (
            <p className="login-card__error">{displayError}</p>
          )}
        </form>

        <p className="login-card__footer">
          {mode === 'login' ? '还没有账号？' : '已有账号？'}
          <button
            onClick={switchMode}
            className="login-card__switch-btn"
          >
            {mode === 'login' ? '立即注册' : '返回登录'}
          </button>
        </p>
      </div>
    </div>
  );
}
