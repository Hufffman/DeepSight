package service

import (
	"DeepSight/internal/config"
	"context"
	"encoding/json"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var LLM *LLMService

type LLMService struct {
	Client             openai.Client
	EmbeddingModel     string
	ChatModel          string
	EmbeddingBatchSize int
}

type ChatMessage struct {
	Role    string
	Content string
}

func InitializeLLM(cfg *config.OpenAIConfig) error {
	client := openai.NewClient(
		option.WithBaseURL(cfg.BaseUrl),
		option.WithAPIKey(cfg.ApiKey),
	)
	LLM = &LLMService{
		Client:             client,
		EmbeddingModel:     cfg.EmbeddingModel,
		ChatModel:          cfg.ChatModel,
		EmbeddingBatchSize: cfg.EmbeddingBatchSize,
	}
	return nil
}

type LLMEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
	index     int
	object    string
}

func (c *LLMService) Embedding(text string) ([]float32, error) {
	ctx := context.Background()
	resp, err := c.Client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{OfString: openai.String(text)},
		Model: c.EmbeddingModel,
	})
	if err != nil {
		return nil, err
	}
	rawResp := resp.JSON.Data.Raw()
	var jsonResp []LLMEmbeddingResponse
	err = json.Unmarshal([]byte(rawResp), &jsonResp)
	if err != nil || len(jsonResp) != 1 {
		return nil, err
	}
	return jsonResp[0].Embedding, nil
}

func (c *LLMService) EmbeddingBatch(texts []string) ([][]float32, error) {
	ctx := context.Background()
	resp, err := c.Client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
		Model: c.EmbeddingModel,
	})
	if err != nil {
		return nil, err
	}
	rawResp := resp.JSON.Data.Raw()
	var jsonResp []LLMEmbeddingResponse
	err = json.Unmarshal([]byte(rawResp), &jsonResp)
	if err != nil {
		return nil, err
	}
	res := make([][]float32, len(jsonResp))
	for i, r := range jsonResp {
		res[i] = r.Embedding
	}
	return res, nil
}

func (c *LLMService) Chat(systemPrompt string, messages []ChatMessage) (string, error) {
	ctx := context.Background()

	chatMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages)+1)

	if systemPrompt != "" {
		chatMessages = append(chatMessages, openai.SystemMessage(systemPrompt))
	}

	for _, msg := range messages {
		if msg.Role == "user" {
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		} else if msg.Role == "assistant" {
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}

	resp, err := c.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.ChatModel,
		Messages: chatMessages,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *LLMService) ChatStream(ctx context.Context, systemPrompt string, messages []ChatMessage) (<-chan string, error) {
	stream := make(chan string, 100)

	chatMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages)+1)

	if systemPrompt != "" {
		chatMessages = append(chatMessages, openai.SystemMessage(systemPrompt))
	}

	for _, msg := range messages {
		if msg.Role == "user" {
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		} else if msg.Role == "assistant" {
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}

	go func() {
		defer close(stream)

		respStream := c.Client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    c.ChatModel,
			Messages: chatMessages,
		})

		for respStream.Next() {
			chunk := respStream.Current()
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				stream <- chunk.Choices[0].Delta.Content
			}
		}

		if err := respStream.Err(); err != nil {
			stream <- "[ERROR]" + err.Error()
		}
	}()

	return stream, nil
}
