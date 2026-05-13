import { create } from 'zustand';

type Theme = 'light' | 'dark';

interface UIState {
  theme: Theme;
  filePanelExpanded: boolean;
  mobileDrawerOpen: boolean;
  settingsOpen: boolean;
  toggleTheme: () => void;
  setTheme: (theme: Theme) => void;
  setFilePanelExpanded: (expanded: boolean) => void;
  setMobileDrawerOpen: (open: boolean) => void;
  setSettingsOpen: (open: boolean) => void;
}

export const uiStore = create<UIState>((set) => ({
  theme: 'light',
  filePanelExpanded: false,
  mobileDrawerOpen: false,
  settingsOpen: false,
  toggleTheme: () =>
    set((state) => ({
      theme: state.theme === 'light' ? 'dark' : 'light',
    })),
  setTheme: (theme: Theme) => set({ theme }),
  setFilePanelExpanded: (expanded: boolean) =>
    set({ filePanelExpanded: expanded }),
  setMobileDrawerOpen: (open: boolean) =>
    set({ mobileDrawerOpen: open }),
  setSettingsOpen: (open: boolean) => set({ settingsOpen: open }),
}));
