package research

import (
	"context"
	"fmt"
	"strings"

	"DeepSight/internal/dto"
	"DeepSight/internal/service"
)

const reporterPrompt = `你是技术能力分析报告生成专家。整合所有子任务的执行结果，生成一份完整的 Markdown 能力分析报告。

## 引用格式要求（重要）
- 引用知识库文件时使用：**来源文件: xxx.md (片段 #N)**
- 引用联网搜索结果时使用：**[网页标题](URL)**
- 每个关键结论都要标明出处，不能只写编号

## 报告结构：
## 1. 项目概述
- 项目定位与技术全景（标明信息来源）

## 2. 项目深度复盘
- 关键技术决策（标明来源）
- 遇到的难点与解决方案（标明来源）
- 体现的工程能力（标明来源）

## 3. 能力雷达
| 能力维度 | 掌握程度 | 证据简述 | 来源 |
|----------|---------|---------|------|

## 4. 行业对标
- 当前水平 vs 行业趋势（标明来源）
- 目标岗位匹配度分析

## 5. 发展建议
- 具体学习路径
- 推荐项目方向
- 优先级排序

## 6. 证据索引
列出报告中用到的所有证据，包含知识库文件的片段和网页链接`

// Reporter synthesizes individual task results into a comprehensive analysis report.
type Reporter struct {
	llm *service.LLMService
}

// NewReporter creates a Reporter backed by the given LLM service.
func NewReporter(llm *service.LLMService) *Reporter {
	return &Reporter{llm: llm}
}

// Report stitches together task results via LLM into a final Markdown analysis,
// including source citations for all referenced files, chunks, and web pages.
func (r *Reporter) Report(ctx context.Context, kbName string, taskSummaries map[int]string, sources []dto.SourceInfo, convID uint) (string, error) {
	var context string
	for i := 1; i <= 5; i++ {
		if result, ok := taskSummaries[i]; ok {
			context += fmt.Sprintf("## 子任务 %d 结果\n\n%s\n\n---\n\n", i, result)
		}
	}

	// Build source appendix with full detail
	sourceAppendix := buildSourceAppendix(sources, convID)

	userMsg := fmt.Sprintf(
		"项目名称：%s\n\n各子任务执行结果如下：\n\n%s\n可用的引用来源：\n\n%s\n请生成完整能力分析报告。报告中必须使用上述引用格式标注每条结论的具体来源（文件名+片段编号 或 网页标题+URL），禁止只写编号。",
		kbName, context, sourceAppendix,
	)

	report, err := r.llm.Chat(reporterPrompt, []service.ChatMessage{{Role: "user", Content: userMsg}})
	if err != nil {
		return "", fmt.Errorf("reporter failed: %w", err)
	}
	return report, nil
}

func buildSourceAppendix(sources []dto.SourceInfo, convID uint) string {
	if len(sources) == 0 {
		return "无"
	}

	var b strings.Builder

	// Conversation context
	b.WriteString("### 来源会话\n")
	fmt.Fprintf(&b, "会话 ID: %d\n\n", convID)

	var kbSources, webSources []dto.SourceInfo
	for _, s := range sources {
		if s.Type == "kb" {
			kbSources = append(kbSources, s)
		} else {
			webSources = append(webSources, s)
		}
	}

	if len(kbSources) > 0 {
		b.WriteString("### 知识库文件及片段\n")
		for i, s := range kbSources {
			fmt.Fprintf(&b, "%d. **%s** (片段 #%d)", i+1, s.FileName, s.ChunkIndex)
			if s.Snippet != "" {
				fmt.Fprintf(&b, " — \"%s\"", s.Snippet)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(webSources) > 0 {
		b.WriteString("### 联网搜索\n")
		for i, s := range webSources {
			fmt.Fprintf(&b, "%d. **[%s](%s)**\n", i+1, s.Title, s.URL)
		}
		b.WriteString("\n")
	}

	return b.String()
}
