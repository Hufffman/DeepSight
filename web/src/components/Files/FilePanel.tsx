import { useState } from 'react';
import { ChevronDown, ChevronRight, Upload, X } from 'lucide-react';
import { FileList } from './FileList';
import { getFileIcon } from '../../utils/fileIcon';
import type { FileItem } from '../../types/models';
import './FilePanel.scss';

interface FilePanelProps {
  kbId: number;
  files: FileItem[];
  loading: boolean;
  expanded: boolean;
  onToggle: () => void;
  onUpload: (file: File) => Promise<void>;
  onDelete: (fileId: number) => Promise<void>;
}

export function FilePanel({
  kbId,
  files,
  loading,
  expanded,
  onToggle,
  onUpload,
  onDelete,
}: FilePanelProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0] ?? null;
    setSelectedFile(file);
  };

  const handleRemoveFile = () => {
    setSelectedFile(null);
    const input = document.getElementById('fileUploadInput') as HTMLInputElement;
    if (input) input.value = '';
  };

  const handleUpload = async () => {
    if (!selectedFile) return;
    setUploading(true);
    try {
      await onUpload(selectedFile);
      setSelectedFile(null);
      const input = document.getElementById('fileUploadInput') as HTMLInputElement;
      if (input) input.value = '';
    } catch {
      // error handled by service layer
    }
    setUploading(false);
  };

  const handleDelete = async (fileId: number) => {
    if (!window.confirm('确定删除此文件？')) return;
    try {
      await onDelete(fileId);
    } catch {
      // error handled by service layer
    }
  };

  if (!kbId) return null;

  return (
    <div className="file-panel">
      <button onClick={onToggle} className="file-panel__toggle">
        {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        文件管理
      </button>

      {expanded && (
        <div className="file-panel__body">
          <div className="file-panel__upload-row">
            <label className="file-panel__file-btn">
              <Upload size={14} />
              选择文件
              <input
                id="fileUploadInput"
                type="file"
                accept=".pdf,.docx,.md,.txt"
                onChange={handleFileChange}
              />
            </label>

            {selectedFile && (
              <div className="file-panel__selected-file">
                <span>{getFileIcon(selectedFile.name)}</span>
                <span className="file-panel__selected-name">
                  {selectedFile.name}
                </span>
                <button
                  onClick={handleRemoveFile}
                  className="file-panel__remove-file"
                >
                  <X size={14} />
                </button>
              </div>
            )}

            <button
              onClick={handleUpload}
              disabled={!selectedFile || uploading}
              className="file-panel__upload-btn"
            >
              {uploading ? '上传中...' : '上传'}
            </button>
          </div>

          {loading ? (
            <div className="file-panel__loading">
              加载中...
            </div>
          ) : (
            <FileList files={files} onDelete={handleDelete} />
          )}
        </div>
      )}
    </div>
  );
}
