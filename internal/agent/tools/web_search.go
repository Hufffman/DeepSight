package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const tavilySearchURL = "https://api.tavily.com/search"

type WebSearchTool struct {
	apiKey string
}

func NewWebSearchTool(apiKey string) *WebSearchTool {
	return &WebSearchTool{apiKey: apiKey}
}

// Search performs a web search using Tavily API.
func (t *WebSearchTool) Search(ctx context.Context, query string, maxResults int) (string, error) {
	if t.apiKey == "" {
		fmt.Println("Tavily API key not configured")
		return "", fmt.Errorf("Tavily API key not configured")
	}
	if maxResults <= 0 {
		maxResults = 5
	}

	reqBody := fmt.Sprintf(`{"api_key":"%s","query":"%s","max_results":%d,"search_depth":"basic"}`,
		t.apiKey, query, maxResults)

	resp, err := http.Post(tavilySearchURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	type outputResult struct {
		Index   int    `json:"index"`
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	}
	output := struct {
		Answer  string         `json:"answer,omitempty"`
		Results []outputResult `json:"results"`
	}{Answer: result.Answer}

	for i, r := range result.Results {
		output.Results = append(output.Results, outputResult{
			Index: i + 1, Title: r.Title, URL: r.URL, Content: r.Content,
		})
	}

	b, _ := json.Marshal(output)
	//fmt.Println(string(b))

	return string(b), nil
}
