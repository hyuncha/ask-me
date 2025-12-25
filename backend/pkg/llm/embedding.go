package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type EmbeddingClient struct {
	apiKey     string
	httpClient *http.Client
}

type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func NewEmbeddingClient(apiKey string) *EmbeddingClient {
	return &EmbeddingClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *EmbeddingClient) CreateEmbedding(text string) ([]float32, error) {
	url := "https://api.openai.com/v1/embeddings"

	reqBody := EmbeddingRequest{
		Input: text,
		Model: "text-embedding-ada-002",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embResp EmbeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return embResp.Data[0].Embedding, nil
}

// CreateEmbeddings creates embeddings for multiple texts
func (c *EmbeddingClient) CreateEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := c.CreateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}
