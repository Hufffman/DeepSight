import { Sun, Moon } from 'lucide-react';
import { uiStore } from '../../stores/uiStore';
import { useEffect } from 'react';
import './ThemeToggle.scss';

export function ThemeToggle() {
  const theme = uiStore((s) => s.theme);
  const toggleTheme = uiStore((s) => s.toggleTheme);

  useEffect(() => {
    const saved = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    if (
      saved === 'dark' ||
      (!saved && prefersDark)
    ) {
      document.documentElement.classList.add('dark');
      uiStore.getState().setTheme('dark');
    }
  }, []);

  const handleToggle = () => {
    toggleTheme();
    const newTheme = uiStore.getState().theme;
    document.documentElement.classList.toggle('dark', newTheme === 'dark');
    localStorage.setItem('theme', newTheme);
  };

  return (
    <button
      onClick={handleToggle}
      className="theme-toggle"
      aria-label="切换主题"
    >
      {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
    </button>
  );
}
