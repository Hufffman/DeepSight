import { useEffect } from 'react';
import { uiStore } from '../stores/uiStore';

export function useTheme() {
  const theme = uiStore((s) => s.theme);
  const toggleTheme = uiStore((s) => s.toggleTheme);

  useEffect(() => {
    const saved = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    if (saved === 'dark' || (!saved && prefersDark)) {
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

  return { theme, toggleTheme: handleToggle };
}
