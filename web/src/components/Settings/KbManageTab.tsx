import { useState } from 'react';
import { Pencil, Trash2, ChevronDown, FolderOpen, Library } from 'lucide-react';
import { kbStore } from '../../stores/kbStore';
import { toastStore } from '../../stores/toastStore';
import { CreateKbModal } from '../KnowledgeBase/CreateKbModal';
import { Skeleton } from '../common/Skeleton';
import { formatDate } from '../../utils/format';
import { getFileIcon } from '../../utils/fileIcon';
import * as fileService from '../../services/fileService';
import type { KnowledgeBase, FileItem } from '../../types/models';
import './KbManageTab.scss';

const STATUS_LABEL: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  completed: '已完成',
  failed: '失败',
};

function chunkArray<T>(arr: T[], cols: number): T[][] {
  const result: T[][] = Array.from({ length: cols }, () => []);
  arr.forEach((item, i) => {
    result[i % cols].push(item);
  });
  return result;
}

export function KbManageTab() {
  const knowledgeBases = kbStore((s) => s.list);
  const loading = kbStore((s) => s.loading);
  const fetchList = kbStore((s) => s.fetchList);
  const updateKb = kbStore((s) => s.update);
  const deleteKb = kbStore((s) => s.delete);
  const createKb = kbStore((s) => s.create);
  const show = toastStore((s) => s.show);

  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editKb, setEditKb] = useState<KnowledgeBase | null>(null);
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<KnowledgeBase | null>(null);

  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());
  const [fileCache, setFileCache] = useState<Map<number, FileItem[]>>(new Map());
  const [loadingIds, setLoadingIds] = useState<Set<number>>(new Set());

  const handleEdit = (kb: KnowledgeBase) => {
    setEditKb(kb);
    setEditName(kb.name);
    setEditDesc(kb.description || '');
  };

  const handleSaveEdit = async () => {
    if (!editKb) return;
    try {
      await updateKb(editKb.id, editName, editDesc || undefined);
      show('success', '知识库已更新');
      setEditKb(null);
    } catch {
      // handled by service layer
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteKb(deleteTarget.id);
      show('success', `已删除「${deleteTarget.name}」`);
      setDeleteTarget(null);
    } catch {
      // handled by service layer
    }
  };

  const handleCreate = async (name: string, description?: string) => {
    await createKb(name, description);
    setCreateModalOpen(false);
  };

  const handleToggleExpand = async (kbId: number) => {
    const wasExpanded = expandedIds.has(kbId);

    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(kbId)) next.delete(kbId);
      else next.add(kbId);
      return next;
    });

    if (wasExpanded || fileCache.has(kbId)) return;

    setLoadingIds((prev) => new Set(prev).add(kbId));
    try {
      const files = await fileService.getFiles(kbId);
      setFileCache((prev) => new Map(prev).set(kbId, files));
    } catch {
      setFileCache((prev) => new Map(prev).set(kbId, []));
    }
    setLoadingIds((prev) => {
      const next = new Set(prev);
      next.delete(kbId);
      return next;
    });
  };

  const handleAfterCreate = async (name: string, description?: string) => {
    await handleCreate(name, description);
    await fetchList();
  };

  const renderFileStatus = (status: string) => (
    <span className={`kb-card__file-status kb-card__file-status--${status}`}>
      {STATUS_LABEL[status] || status}
    </span>
  );

  const renderCard = (kb: KnowledgeBase) => (
    <div key={kb.id} className="kb-card">
      <div className="kb-card__body">
        <div className="kb-card__header">
          <div className="kb-card__title-row">
            <Library size={18} className="kb-card__title-icon" />
            <h3 className="kb-card__title">{kb.name}</h3>
          </div>
          <div className="kb-card__actions">
            <button
              onClick={() => handleEdit(kb)}
              className="kb-card__action-btn kb-card__action-btn--edit"
              title="编辑"
            >
              <Pencil size={14} />
            </button>
            <button
              onClick={() => setDeleteTarget(kb)}
              className="kb-card__action-btn kb-card__action-btn--delete"
              title="删除"
            >
              <Trash2 size={14} />
            </button>
          </div>
        </div>
        {kb.description && (
          <p className="kb-card__desc">{kb.description}</p>
        )}
        <div className="kb-card__meta">
          <span>文件 {kb.file_count ?? 0}</span>
          <span>{formatDate(kb.created_at || '')}</span>
        </div>
      </div>

      <button
        onClick={() => handleToggleExpand(kb.id)}
        className="kb-card__toggle"
      >
        <ChevronDown
          size={14}
          className={`kb-card__toggle-icon ${expandedIds.has(kb.id) ? 'kb-card__toggle-icon--open' : ''}`}
        />
        文件列表
      </button>

      {expandedIds.has(kb.id) && (
        <div className="kb-card__files">
          {loadingIds.has(kb.id) ? (
            <div className="kb-card__files-loading">
              {[1, 2].map((i) => (
                <Skeleton key={i} className="kb-card__files-skeleton" />
              ))}
            </div>
          ) : (fileCache.get(kb.id) || []).length === 0 ? (
            <p className="kb-card__files-empty">暂无文件</p>
          ) : (
            <div className="kb-card__file-list">
              {(fileCache.get(kb.id) || []).map((file) => (
                <div key={file.id} className="kb-card__file-item">
                  <span className="kb-card__file-icon">{getFileIcon(file.file_name)}</span>
                  <span className="kb-card__file-name">{file.file_name}</span>
                  {renderFileStatus(file.status)}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );

  if (loading) {
    return (
      <div className="kb-manage__loading">
        <div className="kb-manage__skeleton-grid">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="kb-manage__skeleton-card" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="kb-manage">
      <div className="kb-manage__top">
        <p className="kb-manage__count">
          共 {knowledgeBases.length} 个知识库
        </p>
        <button
          onClick={() => setCreateModalOpen(true)}
          className="kb-manage__create-btn"
        >
          + 创建知识库
        </button>
      </div>

      {knowledgeBases.length === 0 ? (
        <div className="kb-manage__empty">
          <FolderOpen size={48} className="kb-manage__empty-icon" />
          <p>暂无知识库</p>
        </div>
      ) : (
        <>
          <div className="kb-grid--3col">
            {chunkArray(knowledgeBases, 3).map((col, ci) => (
              <div key={ci} className="kb-grid__col">
                {col.map((kb) => renderCard(kb))}
              </div>
            ))}
          </div>

          <div className="kb-grid--2col">
            {chunkArray(knowledgeBases, 2).map((col, ci) => (
              <div key={ci} className="kb-grid__col">
                {col.map((kb) => renderCard(kb))}
              </div>
            ))}
          </div>

          <div className="kb-grid--1col">
            {knowledgeBases.map((kb) => renderCard(kb))}
          </div>
        </>
      )}

      <CreateKbModal
        open={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSubmit={handleAfterCreate}
      />

      {editKb && (
        <div className="kb-edit-overlay">
          <div className="kb-edit-backdrop" onClick={() => setEditKb(null)} />
          <div className="kb-edit-modal">
            <h3 className="kb-edit-modal__title">编辑知识库</h3>
            <div className="kb-edit-modal__body">
              <div>
                <label className="kb-edit-modal__label">名称</label>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="kb-edit-modal__input"
                />
              </div>
              <div>
                <label className="kb-edit-modal__label">描述</label>
                <input
                  type="text"
                  value={editDesc}
                  onChange={(e) => setEditDesc(e.target.value)}
                  className="kb-edit-modal__input"
                />
              </div>
            </div>
            <div className="kb-edit-modal__actions">
              <button onClick={() => setEditKb(null)} className="kb-edit-modal__cancel">取消</button>
              <button onClick={handleSaveEdit} className="kb-edit-modal__save">保存</button>
            </div>
          </div>
        </div>
      )}

      {deleteTarget && (
        <div className="kb-delete-overlay">
          <div className="kb-delete-backdrop" onClick={() => setDeleteTarget(null)} />
          <div className="kb-delete-modal">
            <h3 className="kb-delete-modal__title">确认删除</h3>
            <p className="kb-delete-modal__msg">
              确定删除「{deleteTarget.name}」？此操作不可恢复。
            </p>
            <div className="kb-delete-modal__actions">
              <button onClick={() => setDeleteTarget(null)} className="kb-delete-modal__cancel">取消</button>
              <button onClick={handleDelete} className="kb-delete-modal__confirm">确认删除</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
