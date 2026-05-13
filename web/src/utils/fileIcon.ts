const FILE_ICON_MAP: Record<string, string> = {
  pdf: '📕',
  txt: '📄',
  doc: '📘',
  docx: '📘',
  md: '📝',
  csv: '📊',
  xls: '📗',
  xlsx: '📗',
  ppt: '📙',
  pptx: '📙',
};

export function getFileIcon(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase() ?? '';
  return FILE_ICON_MAP[ext] || '📄';
}
