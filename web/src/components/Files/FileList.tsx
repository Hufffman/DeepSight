import { Trash2, FileText } from 'lucide-react';
import type { FileListProps } from '../../types/components';
import { getFileIcon } from '../../utils/fileIcon';
import './FileList.scss';

const STATUS_LABEL: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  completed: '已完成',
  failed: '失败',
};

export function FileList({ files, onDelete }: FileListProps) {
  if (files.length === 0) {
    return (
      <div className="file-list__empty">
        <FileText size={32} className="file-list__empty-icon" />
        <p className="file-list__empty-text">暂无文件</p>
      </div>
    );
  }

  return (
    <div className="file-list">
      {files.map((file) => (
        <div key={file.id} className="file-card">
          <span className="file-card__icon">{getFileIcon(file.file_name)}</span>
          <div className="file-card__info">
            <p className="file-card__name">{file.file_name}</p>
            <span className={`file-card__status file-card__status--${file.status}`}>
              {STATUS_LABEL[file.status] || file.status}
            </span>
          </div>
          <button
            onClick={() => onDelete(file.id)}
            className="file-card__delete"
            title="删除"
          >
            <Trash2 size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}
