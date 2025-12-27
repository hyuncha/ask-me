package vector

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const pineconeAssistantBaseURL = "https://prod-1-data.ke.pinecone.io"

type PineconeAssistantClient struct {
	apiKey      string
	assistantID string
	httpClient  *http.Client
}

type AssistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AssistantChatRequest struct {
	Messages []AssistantMessage `json:"messages"`
	Stream   bool               `json:"stream"`
	Model    string             `json:"model"`
}

type AssistantChatResponse struct {
	ID      string `json:"id,omitempty"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
}

type AssistantStreamChunk struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
}

func NewPineconeAssistantClient(apiKey, assistantID string) *PineconeAssistantClient {
	return &PineconeAssistantClient{
		apiKey:      apiKey,
		assistantID: assistantID,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Chat sends a message to Pinecone Assistant and returns the response
func (c *PineconeAssistantClient) Chat(query string) (string, error) {
	url := fmt.Sprintf("%s/assistant/chat/%s", pineconeAssistantBaseURL, c.assistantID)

	reqBody := AssistantChatRequest{
		Messages: []AssistantMessage{
			{
				Role:    "user",
				Content: query,
			},
		},
		Stream: false,
		Model:  "gpt-4o",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Pinecone Assistant API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp AssistantChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return chatResp.Message.Content, nil
}

// ChatWithStream sends a message and processes streaming response
func (c *PineconeAssistantClient) ChatWithStream(query string) (string, error) {
	url := fmt.Sprintf("%s/assistant/chat/%s", pineconeAssistantBaseURL, c.assistantID)

	reqBody := AssistantChatRequest{
		Messages: []AssistantMessage{
			{
				Role:    "user",
				Content: query,
			},
		},
		Stream: true,
		Model:  "gpt-4o",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Pinecone Assistant API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Process SSE stream
	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and SSE comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Skip [DONE] marker
			if data == "[DONE]" {
				break
			}

			var chunk AssistantStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.Content != "" {
				fullContent.WriteString(chunk.Content)
			}
			if chunk.Message.Content != "" {
				fullContent.WriteString(chunk.Message.Content)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading stream: %w", err)
	}

	return fullContent.String(), nil
}

// Search queries the Pinecone Assistant for relevant context
func (c *PineconeAssistantClient) Search(query string) (string, error) {
	// Use streaming chat to get context from the assistant
	return c.ChatWithStream(query)
}
