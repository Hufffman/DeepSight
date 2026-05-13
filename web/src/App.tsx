import { useState, useEffect, useCallback } from 'react';
import { Layout } from './components/Layout/Layout';
import { LoginCard } from './components/Auth/LoginCard';
import { ChatPanel } from './components/Chat/ChatPanel';
import { FilePanel } from './components/Files/FilePanel';
import { SettingsPage } from './components/Settings/SettingsPage';
import { CreateKbModal } from './components/KnowledgeBase/CreateKbModal';
import { ToastContainer } from './components/common/Toast';
import { authStore } from './stores/authStore';
import { kbStore } from './stores/kbStore';
import { conversationStore } from './stores/conversationStore';
import { messageStore } from './stores/messageStore';
import { uiStore } from './stores/uiStore';
import { login as apiLogin, register as apiRegister } from './services/authService';
import type { FileItem } from './types/models';
import * as fileService from './services/fileService';

export function App() {
  const isLoggedIn = authStore((s) => s.isLoggedIn);
  const login = authStore((s) => s.login);
  const currentKbId = kbStore((s) => s.currentKbId);
  const fetchKbList = kbStore((s) => s.fetchList);
  const createKb = kbStore((s) => s.create);

  const activeTabId = conversationStore((s) => s.activeTabId);
  const createConv = conversationStore((s) => s.create);

  const fetchMessages = messageStore((s) => s.fetchMessages);
  const filePanelExpanded = uiStore((s) => s.filePanelExpanded);
  const setFilePanelExpanded = uiStore((s) => s.setFilePanelExpanded);
  const settingsOpen = uiStore((s) => s.settingsOpen);
  const setSettingsOpen = uiStore((s) => s.setSettingsOpen);

  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);

  const [kbModalOpen, setKbModalOpen] = useState(false);

  const [files, setFiles] = useState<FileItem[]>([]);
  const [filesLoading, setFilesLoading] = useState(false);

  // Initialize theme on mount
  useEffect(() => {
    const saved = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    if (saved === 'dark' || (!saved && prefersDark)) {
      document.documentElement.classList.add('dark');
      uiStore.getState().setTheme('dark');
    }
  }, []);

  // Load knowledge bases after login
  useEffect(() => {
    if (isLoggedIn) {
      fetchKbList();
    }
  }, [isLoggedIn, fetchKbList]);

  // Load messages when active tab changes
  useEffect(() => {
    if (activeTabId) {
      fetchMessages(activeTabId);
    }
  }, [activeTabId, fetchMessages]);

  // Load files when knowledge base changes
  useEffect(() => {
    if (!currentKbId) {
      setFiles([]);
      return;
    }
    setFilesLoading(true);
    fileService
      .getFiles(currentKbId)
      .then((data) => {
        setFiles(data);
        setFilesLoading(false);
      })
      .catch((err) => {
        setFilesLoading(false);
        console.error('获取文件列表失败:', err);
      });
  }, [currentKbId]);

  const handleLogin = useCallback(
    async (username: string, password: string) => {
      setLoginLoading(true);
      setLoginError(null);
      try {
        const data = await apiLogin({ username, password });
        login(data.token, data.user_id);
      } catch (err: any) {
        setLoginError(err.message || '登录失败');
      }
      setLoginLoading(false);
    },
    [login]
  );

  const handleRegister = useCallback(
    async (username: string, password: string, email: string) => {
      setLoginLoading(true);
      setLoginError(null);
      try {
        await apiRegister({ username, password, email });
        const data = await apiLogin({ username, password });
        login(data.token, data.user_id);
      } catch (err: any) {
        setLoginError(err.message || '注册失败');
      }
      setLoginLoading(false);
    },
    [login]
  );

  const handleCreateConv = useCallback(async () => {
    if (!currentKbId) return;
    try {
      await createConv(currentKbId);
    } catch {
      // handled by service layer
    }
  }, [currentKbId, createConv]);

  const handleUploadFile = useCallback(
    async (file: File) => {
      if (!currentKbId) return;
      await fileService.uploadFile(currentKbId, file);
      const data = await fileService.getFiles(currentKbId);
      setFiles(data);
    },
    [currentKbId]
  );

  const handleDeleteFile = useCallback(
    async (fileId: number) => {
      if (!currentKbId) return;
      await fileService.deleteFile(currentKbId, fileId);
      const data = await fileService.getFiles(currentKbId);
      setFiles(data);
    },
    [currentKbId]
  );

  if (!isLoggedIn) {
    return (
      <>
        <LoginCard
          onLogin={handleLogin}
          onRegister={handleRegister}
          loading={loginLoading}
          error={loginError}
        />
        <ToastContainer />
      </>
    );
  }

  return (
    <>
      <Layout
        onCreateKbClick={() => setKbModalOpen(true)}
        onCreateConvClick={handleCreateConv}
        onSettingsClick={() => setSettingsOpen(true)}
        settingsOpen={settingsOpen}
      >
        {settingsOpen ? (
          <SettingsPage />
        ) : (
          <>
            <ChatPanel />
            <FilePanel
              kbId={currentKbId ?? 0}
              files={files}
              loading={filesLoading}
              expanded={filePanelExpanded}
              onToggle={() => setFilePanelExpanded(!filePanelExpanded)}
              onUpload={handleUploadFile}
              onDelete={handleDeleteFile}
            />
          </>
        )}
      </Layout>

      <CreateKbModal
        open={kbModalOpen}
        onClose={() => setKbModalOpen(false)}
        onSubmit={(name, description) => createKb(name, description || undefined)}
      />

      <ToastContainer />
    </>
  );
}
