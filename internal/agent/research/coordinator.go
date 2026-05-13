package research

import (
	"context"
	"encoding/json"
	"fmt"

	"DeepSight/internal/agent/tools"
	"DeepSight/internal/dto"
	"DeepSight/internal/service"
	"DeepSight/internal/model"
	"DeepSight/internal/repository"
)

// Coordinator orchestrates the 3-stage deep research pipeline:
// Stage 1 (Plan), Stage 2 (Execute), Stage 3 (Report).
type Coordinator struct {
	planner      *Planner
	executor     *Executor
	reporter     *Reporter
	analysisRepo *repository.AnalysisRepository
	kbRepo       *repository.KnowledgeBaseRepository
	msgRepo      *repository.MessageRepository
	chunkRepo    *repository.ChunkRepository
	fileRepo     *repository.FileRepository
	userID       uint
	kbID         uint
	convID       uint
}

// NewCoordinator wires up all dependencies and returns a ready-to-use Coordinator.
func NewCoordinator(
	llm *service.LLMService,
	analysisRepo *repository.AnalysisRepository,
	chunkRepo *repository.ChunkRepository,
	fileRepo *repository.FileRepository,
	kbRepo *repository.KnowledgeBaseRepository,
	msgRepo *repository.MessageRepository,
	webSearch *tools.WebSearchTool,
	userID, kbID, convID uint,
) *Coordinator {
	expTools := tools.NewExperienceTools(analysisRepo, userID, kbID)
	return &Coordinator{
		planner:      NewPlanner(llm),
		executor:     NewExecutor(llm, expTools, webSearch, chunkRepo, fileRepo),
		reporter:     NewReporter(llm),
		analysisRepo: analysisRepo,
		kbRepo:       kbRepo,
		msgRepo:      msgRepo,
		chunkRepo:    chunkRepo,
		fileRepo:     fileRepo,
		userID:       userID,
		kbID:         kbID,
		convID:       convID,
	}
}

// Run drives the full 3-stage pipeline, emitting SSE events and persisting the final report.
func (c *Coordinator) Run(ctx context.Context, emit func(event dto.AnalysisSSEEvent)) (string, error) {
	kb, err := c.kbRepo.GetByID(c.kbID)
	if err != nil {
		return "", fmt.Errorf("knowledge base not found: %w", err)
	}

	// Stage 1: Plan — collect chat messages and document chunks for context-aware planning
	emit(dto.AnalysisSSEEvent{Type: "status", Title: "正在分析聊天记录和项目文档..."})

	chatMsgs, _ := c.msgRepo.GetRecentByConversationID(c.convID, 50)
	chats := make([]string, len(chatMsgs))
	for i, m := range chatMsgs {
		role := "用户"
		if m.Role == "assistant" {
			role = "AI"
		}
		chats[i] = fmt.Sprintf("[%s] %s", role, m.Content)
	}

	hashes, _ := c.fileRepo.GetHashesByKnowledgeBaseID(c.kbID)
	queryEmbedding, _ := service.LLM.Embedding(kb.Name + " " + kb.Description)
	chunkList, _ := c.chunkRepo.SearchSimilar(queryEmbedding, hashes, 15)
	chunks := make([]string, len(chunkList))
	for i, c := range chunkList {
		chunks[i] = c.Content
	}

	emit(dto.AnalysisSSEEvent{Type: "status", Title: "正在规划研究任务..."})
	todos, err := c.planner.Plan(ctx, kb.Name, kb.Description, chats, chunks)
	if err != nil {
		emit(dto.AnalysisSSEEvent{Type: "error", Content: err.Error()})
		return "", err
	}
	todosJSON, _ := json.Marshal(todos)
	emit(dto.AnalysisSSEEvent{Type: "plan", Content: string(todosJSON), Todos: todos})

	// Stage 2: Execute — collect task results with source info
	taskSummaries := make(map[int]string)
	var allSources []dto.SourceInfo

	for _, todo := range todos {
		result, err := c.executor.Execute(ctx, todo, c.kbID, emit)
		if err != nil {
			emit(dto.AnalysisSSEEvent{Type: "error", Content: fmt.Sprintf("子任务 %d 失败: %v", todo.ID, err)})
			taskSummaries[todo.ID] = fmt.Sprintf("执行失败: %v", err)
			continue
		}
		taskSummaries[todo.ID] = result.Summary
		allSources = append(allSources, result.Sources...)
	}

	// Deduplicate sources
	allSources = deduplicateSources(allSources)

	// Stage 3: Report — include source info for citations
	emit(dto.AnalysisSSEEvent{Type: "status", Title: "整合所有分析结果，生成能力报告..."})
	report, err := c.reporter.Report(ctx, kb.Name, taskSummaries, allSources, c.convID)
	if err != nil {
		emit(dto.AnalysisSSEEvent{Type: "error", Content: err.Error()})
		return "", err
	}

	// Persist report
	execLogJSON, _ := json.Marshal(taskSummaries)
	analysisReport := &model.AnalysisReport{
		UserID:          c.userID,
		KnowledgeBaseID: c.kbID,
		ReportType:      "full",
		Content:         report,
		PlanJSON:        string(todosJSON),
		ExecutionLog:    string(execLogJSON),
	}
	if err := c.analysisRepo.CreateAnalysisReport(analysisReport); err != nil {
		return "", fmt.Errorf("save report: %w", err)
	}

	emit(dto.AnalysisSSEEvent{Type: "report", Content: report})
	return report, nil
}

func deduplicateSources(sources []dto.SourceInfo) []dto.SourceInfo {
	seen := make(map[string]bool)
	var result []dto.SourceInfo
	for _, s := range sources {
		key := fmt.Sprintf("%s|%s|%d|%s", s.Type, s.FileName, s.ChunkIndex, s.URL)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, s)
	}
	return result
}
