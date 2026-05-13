package dto

// StartAnalysisRequest 启动深度分析请求
type StartAnalysisRequest struct {
	KnowledgeBaseID uint `json:"knowledge_base_id" binding:"required"`
}

// TodoItem Planner 输出的子任务
type TodoItem struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Intent string `json:"intent"`
	Query  string `json:"query"`
}

// AnalysisSSEEvent SSE 事件
type AnalysisSSEEvent struct {
	Type    string      `json:"type"`
	Title   string      `json:"title,omitempty"`
	Index   int         `json:"index,omitempty"`
	Content string      `json:"content,omitempty"`
	Todos   []TodoItem  `json:"todos,omitempty"`
	Cmd     string      `json:"cmd,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// SkillProfileResponse 能力画像响应
type SkillProfileResponse struct {
	ID              uint   `json:"id"`
	SkillName       string `json:"skill_name"`
	Category        string `json:"category"`
	Level           string `json:"level"`
	EvidenceCount   int    `json:"evidence_count"`
	Summary         string `json:"summary"`
	LastEvidencedAt string `json:"last_evidenced_at"`
}

// SkillProfileListResponse 能力画像列表响应
type SkillProfileListResponse struct {
	Skills   []SkillProfileResponse `json:"skills"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// AnalysisReportResponse 分析报告响应
type AnalysisReportResponse struct {
	ID              uint   `json:"id"`
	KnowledgeBaseID uint   `json:"knowledge_base_id"`
	ReportType      string `json:"report_type"`
	Content         string `json:"content"`
	CreatedAt       string `json:"created_at"`
}

// AnalysisReportListResponse 分析报告列表响应
type AnalysisReportListResponse struct {
	Reports  []AnalysisReportResponse `json:"reports"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

// SourceInfo 分析引用的来源信息
type SourceInfo struct {
	Type       string `json:"type"` // "kb" 或 "web"
	FileName   string `json:"file_name,omitempty"`
	ChunkIndex int    `json:"chunk_index,omitempty"`
	Snippet    string `json:"snippet,omitempty"`
	URL        string `json:"url,omitempty"`
	Title      string `json:"title,omitempty"`
}

// TaskResult 子任务执行结果（含来源）
type TaskResult struct {
	Summary string       `json:"summary"`
	Sources []SourceInfo `json:"sources"` // 本任务用到的所有数据来源
}

// ExperienceItemResponse 经历条目响应
type ExperienceItemResponse struct {
	ID          uint     `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	SkillTags   []string `json:"skill_tags"`
	Evidence    string   `json:"evidence"`
	Confidence  float64  `json:"confidence"`
}
