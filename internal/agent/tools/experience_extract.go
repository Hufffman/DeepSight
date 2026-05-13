package tools

import (
	"DeepSight/internal/repository"
	"DeepSight/internal/service"
	"context"
	"fmt"
)

type ExperienceTools struct {
	analysisRepo *repository.AnalysisRepository
	userID       uint
	kbID         uint
}

func NewExperienceTools(analysisRepo *repository.AnalysisRepository, userID, kbID uint) *ExperienceTools {
	return &ExperienceTools{analysisRepo: analysisRepo, userID: userID, kbID: kbID}
}

// ExtractExperience calls LLM to extract structured experiences from search results.
func (t *ExperienceTools) ExtractExperience(ctx context.Context, chunksJSON string, intent string) (string, error) {
	prompt := `你是技术经历抽取专家。从给定的检索结果中抽取与技术能力相关的经历。
只抽取有明确证据支持的内容，不要编造。

对于每条经历，判断类型：knowledge(知识学习), project_action(项目动作), problem(遇到的问题), solution(解决方案)
返回 JSON 数组。`

	messages := []service.ChatMessage{{
		Role: "user",
		Content: fmt.Sprintf(
			"%s\n\n意图：%s\n\n检索结果：%s\n\n返回 JSON 数组，格式：[{\"type\":\"...\",\"title\":\"...\",\"description\":\"...\",\"skill_tags\":[\"...\"],\"evidence\":\"原文片段\",\"confidence\":0.9}]",
			prompt, intent, chunksJSON,
		),
	}}

	result, err := service.LLM.Chat(prompt, messages)
	if err != nil {
		return "", err
	}
	return result, nil
}
