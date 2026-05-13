import type { KbSelectorProps } from '../../types/components';
import './KbSelector.scss';

export function KbSelector({
  knowledgeBases,
  currentKbId,
  loading,
  onChange,
  onCreateClick,
}: KbSelectorProps) {
  return (
    <div className="kb-selector">
      <select
        value={currentKbId ?? ''}
        onChange={(e) => {
          const val = e.target.value;
          onChange(val ? Number(val) : null);
        }}
        className="kb-selector__select"
      >
        <option value="" disabled={loading}>
          {loading ? '加载中...' : '-- 选择知识库 --'}
        </option>
        {knowledgeBases.map((kb) => (
          <option key={kb.id} value={kb.id}>
            {kb.name}
          </option>
        ))}
      </select>
      <button
        onClick={onCreateClick}
        className="kb-selector__create-btn"
      >
        + 创建
      </button>
    </div>
  );
}
