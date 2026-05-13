import { useState } from 'react';
import { X } from 'lucide-react';
import type { CreateKbModalProps } from '../../types/components';
import './CreateKbModal.scss';

export function CreateKbModal({ open, onClose, onSubmit }: CreateKbModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  if (!open) return null;

  const handleSubmit = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) {
      setError('请输入知识库名称');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await onSubmit(trimmedName, description.trim() || undefined);
      setName('');
      setDescription('');
      onClose();
    } catch {
      setError('创建失败，请重试');
    }
    setLoading(false);
  };

  return (
    <div
      className="modal-overlay"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="modal">
        <div className="modal__header">
          <h3 className="modal__title">
            创建知识库
          </h3>
          <button
            onClick={onClose}
            className="modal__close"
          >
            <X size={18} />
          </button>
        </div>

        <div className="modal__body">
          <div>
            <label className="modal__label">
              名称
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="modal__input"
              placeholder="输入知识库名称"
            />
          </div>
          <div>
            <label className="modal__label">
              描述
            </label>
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="modal__input"
              placeholder="输入描述（可选）"
            />
          </div>

          <button
            onClick={handleSubmit}
            disabled={loading}
            className="modal__submit"
          >
            {loading ? '创建中...' : '创建'}
          </button>

          {error && (
            <p className="modal__error">{error}</p>
          )}
        </div>
      </div>
    </div>
  );
}
