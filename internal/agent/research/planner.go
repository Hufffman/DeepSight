package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"DeepSight/internal/dto"
	"DeepSight/internal/service"
)

const summaryPrompt = `你是项目分析专家。根据以下聊天记录和项目文档内容，用 300 字以内总结：
1. 这个项目主要是做什么的
2. 使用了哪些关键技术和工具
3. 用户在这个项目中关注什么、遇到了什么问题
4. 从聊天中能看出用户具备哪些能力`

const plannerPrompt = `你是技术能力分析规划专家。
根据下方的项目摘要，将"分析用户能力"分解为 5 个子任务。

固定五个维度：
1. 技术栈梳理 — 结合摘要中提到的具体技术生成检索 query
2. 难点与解决方案 — 结合摘要中提到的问题生成检索 query
3. 能力维度评估 — 结合摘要中体现的能力，评级: aware(了解)/learned(学习过)/practiced(实践过)/proficient(精通)
4. 行业对标 — 结合当前行业趋势，对比用户能力与目标岗位要求
5. 发展路线推荐 — 基于能力缺口推荐具体学习路径和项目方向

每个子任务生成对应的 search_query 用于检索知识库。
title 和 intent 要具体，不能是"技术栈梳理"这种泛化标题，要包含摘要中的具体技术名词。

输出严格 JSON，输出严格 JSON，输出严格 JSON，格式如下:
{"todos": [{"id": 1, "title": "...", "intent": "...", "query": "..."}]}`

type Planner struct {
	llm *service.LLMService
}

func NewPlanner(llm *service.LLMService) *Planner {
	return &Planner{llm: llm}
}

func (p *Planner) Plan(ctx context.Context, kbName, kbDescription string, chats []string, chunks []string) ([]dto.TodoItem, error) {
	// Step 1: LLM 总结聊天记录和文档，生成项目摘要
	chatText := strings.Join(chats, "\n")
	chunkText := strings.Join(chunks, "\n---\n")

	summaryUserMsg := fmt.Sprintf(
		"项目名称：%s\n项目描述：%s\n\n聊天记录：\n%s\n\n文档片段：\n%s",
		kbName, kbDescription, chatText, chunkText,
	)
	summary, err := p.llm.Chat(summaryPrompt, []service.ChatMessage{{Role: "user", Content: summaryUserMsg}})
	if err != nil {
		return nil, fmt.Errorf("planner summary failed: %w", err)
	}

	// Step 2: 基于摘要生成 5 个针对性 TODO
	todoUserMsg := fmt.Sprintf(
		"项目名称：%s\n\n项目摘要：\n%s\n\n请生成 5 个针对性的分析子任务。",
		kbName, summary,
	)
	result, err := p.llm.Chat(plannerPrompt, []service.ChatMessage{{Role: "user", Content: todoUserMsg}})
	if err != nil {
		return nil, fmt.Errorf("planner todo generation failed: %w", err)
	}

	var output struct {
		Todos []dto.TodoItem `json:"todos"`
	}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		return nil, fmt.Errorf("failed to parse planner output: %w", err)
	}

	for i := range output.Todos {
		output.Todos[i].ID = i + 1
	}

	return output.Todos, nil
}
