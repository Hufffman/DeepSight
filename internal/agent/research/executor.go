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

// Executor runs a single todo item: search knowledge base, optionally web-search,
// and extract structured experiences via LLM.
type Executor struct {
	llm       *service.LLMService
	expTools  *tools.ExperienceTools
	webSearch *tools.WebSearchTool
	chunkRepo *repository.ChunkRepository
	fileRepo  *repository.FileRepository
}

// NewExecutor creates an Executor with all required dependencies.
func NewExecutor(
	llm *service.LLMService,
	expTools *tools.ExperienceTools,
	webSearch *tools.WebSearchTool,
	chunkRepo *repository.ChunkRepository,
	fileRepo *repository.FileRepository,
) *Executor {
	return &Executor{llm: llm, expTools: expTools, webSearch: webSearch, chunkRepo: chunkRepo, fileRepo: fileRepo}
}

// chunkWithFile holds chunk content with its source file name.
type chunkWithFile struct {
	Content  string
	FileName string
}

// webResult holds structured web search result.
type webResult struct {
	Title   string
	URL     string
	Content string
}

// Execute performs the full pipeline for a single todo item, emitting SSE events along the way.
func (e *Executor) Execute(ctx context.Context, todo dto.TodoItem, kbID uint, emit func(event dto.AnalysisSSEEvent)) (*dto.TaskResult, error) {
	emit(dto.AnalysisSSEEvent{Type: "task_start", Index: todo.ID, Title: todo.Title})

	// 1. Search knowledge base
	embedding, err := service.LLM.Embedding(todo.Query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	fileHashes, err := e.fileRepo.GetHashesByKnowledgeBaseID(kbID)
	if err != nil {
		return nil, fmt.Errorf("get file hashes: %w", err)
	}

	chunks, err := e.chunkRepo.SearchSimilar(embedding, fileHashes, 10)
	if err != nil {
		return nil, fmt.Errorf("search kb: %w", err)
	}
	emit(dto.AnalysisSSEEvent{Type: "status", Title: fmt.Sprintf("检索到 %d 条相关内容", len(chunks))})

	// 2. Resolve file names for chunks
	fileNameMap := e.resolveFileNames(chunks)
	chunksWithFiles := make([]chunkWithFile, len(chunks))
	for i, c := range chunks {
		chunksWithFiles[i] = chunkWithFile{
			Content:  c.Content,
			FileName: fileNameMap[c.FileHash],
		}
	}

	// 3. Build context and collect KB sources
	kbSources := make([]dto.SourceInfo, 0)
	chunkContexts := make([]struct {
		Index    int    `json:"index"`
		FileName string `json:"file_name"`
		Content  string `json:"content"`
	}, len(chunksWithFiles))

	for i, c := range chunksWithFiles {
		chunkContexts[i] = struct {
			Index    int    `json:"index"`
			FileName string `json:"file_name"`
			Content  string `json:"content"`
		}{Index: i + 1, FileName: c.FileName, Content: c.Content}

		snippet := c.Content
		if len(snippet) > 120 {
			snippet = snippet[:120] + "..."
		}
		kbSources = append(kbSources, dto.SourceInfo{
			Type:       "kb",
			FileName:   c.FileName,
			ChunkIndex: chunkContexts[i].Index,
			Snippet:    snippet,
		})
	}
	chunksJSON, _ := json.Marshal(chunkContexts)
	contextText := string(chunksJSON)

	// 4. WebSearch for todo items 4/5 (industry benchmarking, roadmap)
	var webSources []dto.SourceInfo
	var webResults []webResult

	if todo.ID >= 4 {
		emit(dto.AnalysisSSEEvent{Type: "status", Title: fmt.Sprintf("联网搜索: %s", todo.Query)})
		results, searchErr := e.webSearch.Search(ctx, todo.Query, 5)
		if searchErr == nil {
			webResults = e.parseWebResults(results)
			for _, wr := range webResults {
				webSources = append(webSources, dto.SourceInfo{
					Type:  "web",
					Title: wr.Title,
					URL:   wr.URL,
				})
			}

			webContext := make([]struct {
				Index   int    `json:"index"`
				Title   string `json:"title"`
				URL     string `json:"url"`
				Content string `json:"content"`
			}, len(webResults))
			for i, wr := range webResults {
				webContext[i] = struct {
					Index   int    `json:"index"`
					Title   string `json:"title"`
					URL     string `json:"url"`
					Content string `json:"content"`
				}{Index: i + 1, Title: wr.Title, URL: wr.URL, Content: wr.Content}
			}
			wcJSON, _ := json.Marshal(webContext)
			contextText += "\n\n联网搜索结果:\n" + string(wcJSON)
		}
	}

	// 5. Extract experiences via LLM
	emit(dto.AnalysisSSEEvent{Type: "status", Title: fmt.Sprintf("分析中: %s", todo.Title)})
	summary, err := e.expTools.ExtractExperience(ctx, contextText, todo.Intent)
	if err != nil {
		return nil, fmt.Errorf("extract experience: %w", err)
	}

	allSources := append(kbSources, webSources...)

	emit(dto.AnalysisSSEEvent{Type: "task_completed", Index: todo.ID, Title: todo.Title, Content: summary})
	return &dto.TaskResult{Summary: summary, Sources: allSources}, nil
}

// resolveFileNames looks up file names for a set of chunks.
func (e *Executor) resolveFileNames(chunks []model.Chunk) map[string]string {
	nameMap := make(map[string]string)
	seen := make(map[string]bool)
	for _, c := range chunks {
		if seen[c.FileHash] {
			continue
		}
		seen[c.FileHash] = true
		files, err := e.fileRepo.GetByHash(c.FileHash)
		if err != nil || len(files) == 0 {
			nameMap[c.FileHash] = c.FileHash[:8]
			continue
		}
		nameMap[c.FileHash] = files[0].FileName
	}
	return nameMap
}

// parseWebResults extracts structured web search results from JSON response.
func (e *Executor) parseWebResults(rawJSON string) []webResult {
	var parsed struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return nil
	}
	results := make([]webResult, len(parsed.Results))
	for i, r := range parsed.Results {
		results[i] = webResult{Title: r.Title, URL: r.URL, Content: r.Content}
	}
	return results
}
