import { useState, useEffect, useMemo } from 'react';
import { FileText, Eye, Trash2, BookOpen } from 'lucide-react';
import { MarkdownRenderer } from '../Chat/MarkdownRenderer';
import { kbStore } from '../../stores/kbStore';
import * as analysisService from '../../services/analysisService';
import type { AnalysisReport } from '../../types/models';
import './ReportTab.scss';

export function ReportTab() {
  const knowledgeBases = kbStore((s) => s.list);

  const [reports, setReports] = useState<AnalysisReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState<AnalysisReport | null>(null);
  const [viewTarget, setViewTarget] = useState<AnalysisReport | null>(null);

  useEffect(() => {
    analysisService
      .getAnalysisReports()
      .then(setReports)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const grouped = useMemo(() => {
    const map: Record<number, AnalysisReport[]> = {};
    for (const r of reports) {
      if (!map[r.knowledge_base_id]) {
        map[r.knowledge_base_id] = [];
      }
      map[r.knowledge_base_id].push(r);
    }
    return map;
  }, [reports]);

  const kbName = (id: number) => {
    const kb = knowledgeBases.find((k) => k.id === id);
    return kb ? kb.name : `知识库 #${id}`;
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await analysisService.deleteAnalysisReport(deleteTarget.id);
      setReports((prev) => prev.filter((r) => r.id !== deleteTarget.id));
    } catch {
      // handled by service layer
    }
    setDeleteTarget(null);
  };

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Loading
  if (loading) {
    return (
      <div className="report-tab">
        <div className="report-tab__loading-group">
          <div className="report-tab__skeleton report-tab__skeleton--title" />
          <div className="report-tab__skeleton report-tab__skeleton--row" />
          <div className="report-tab__skeleton report-tab__skeleton--row" />
        </div>
      </div>
    );
  }

  // Empty
  if (reports.length === 0) {
    return (
      <div className="report-tab">
        <div className="report-tab__empty">
          <BookOpen size={40} className="report-tab__empty-icon" />
          <p>暂无分析报告</p>
          <p className="report-tab__empty-hint">对知识库执行深度分析后，报告将显示在此处</p>
        </div>
      </div>
    );
  }

  const groupIds = Object.keys(grouped).map(Number);

  return (
    <div className="report-tab">
      <div className="report-tab__summary">
        <span className="report-tab__summary-text">
          共 {reports.length} 份报告，{groupIds.length} 个知识库
        </span>
      </div>

      <div className="report-tab__groups">
        {groupIds.map((kbId) => (
          <div key={kbId} className="report-group">
            <div className="report-group__header">
              <FileText size={16} className="report-group__icon" />
              <span className="report-group__name">{kbName(kbId)}</span>
              <span className="report-group__count">{grouped[kbId].length} 份</span>
            </div>

            <div className="report-table">
              <table>
                <thead className="report-table__head">
                  <tr>
                    <th className="report-table__th report-table__th--left">类型</th>
                    <th className="report-table__th report-table__th--left">创建时间</th>
                    <th className="report-table__th report-table__th--center">操作</th>
                  </tr>
                </thead>
                <tbody className="report-table__body">
                  {grouped[kbId].map((r) => (
                    <tr key={r.id} className="report-table__row">
                      <td className="report-table__td report-table__td--left">
                        <button
                          onClick={() => setViewTarget(r)}
                          className="report-table__link"
                        >
                          <Eye size={14} className="report-table__link-icon" />
                          {r.report_type === 'full' ? '完整分析' : r.report_type}
                        </button>
                      </td>
                      <td className="report-table__td report-table__td--left">
                        <span className="report-table__date">{formatDate(r.created_at)}</span>
                      </td>
                      <td className="report-table__td report-table__td--center">
                        <button
                          onClick={() => setDeleteTarget(r)}
                          className="report-table__delete-btn"
                        >
                          <Trash2 size={16} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
      </div>

      {/* View Markdown Modal */}
      {viewTarget && (
        <div className="report-view-overlay">
          <div className="report-view-backdrop" onClick={() => setViewTarget(null)} />
          <div className="report-view">
            <div className="report-view__header">
              <h3 className="report-view__title">
                {kbName(viewTarget.knowledge_base_id)} — 分析报告
              </h3>
              <button
                onClick={() => setViewTarget(null)}
                className="report-view__close"
              >
                关闭
              </button>
            </div>
            <div className="report-view__body">
              <MarkdownRenderer content={viewTarget.content} />
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirm Modal */}
      {deleteTarget && (
        <div className="report-confirm-overlay">
          <div className="report-confirm-backdrop" onClick={() => setDeleteTarget(null)} />
          <div className="report-confirm">
            <h3 className="report-confirm__title">删除报告</h3>
            <p className="report-confirm__msg">
              确定删除「{kbName(deleteTarget.knowledge_base_id)}」的这份分析报告吗？此操作不可恢复。
            </p>
            <div className="report-confirm__actions">
              <button
                onClick={() => setDeleteTarget(null)}
                className="report-confirm__cancel"
              >
                取消
              </button>
              <button
                onClick={handleDelete}
                className="report-confirm__danger"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
