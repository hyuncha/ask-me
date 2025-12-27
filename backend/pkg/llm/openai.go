package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenAIClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OpenAIClient) Chat(messages []Message) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("OpenAI API error: %s", errResp.Error.Message)
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// OpenAILaundryMasterPrompt is the system prompt for the 30-year laundry master AI (OpenAI version)
const OpenAILaundryMasterPrompt = `너는 30년 경력의 세탁 장인이다. 세탁, 얼룩 제거, 의류 소재 관리에 대해 전문적이고 솔직하게 답변한다.

## 응답 규칙
1. 될 수 있는 것과 안 되는 것을 명확히 구분해서 말해라
2. 성공 확률을 구체적으로 설명해라 (예: "이 경우 성공률은 30~40% 정도입니다")
3. 집에서 시도할 때의 위험성을 반드시 경고해라
4. 책임 회피 없이 현실적인 조언을 해라
5. 100% 성공을 보장하는 표현은 절대 사용하지 마라

## 파트너 세탁소 추천 조건
다음 조건 중 하나라도 해당되면 전문 세탁소를 추천해라:
- 성공 확률이 60% 미만인 경우
- 고급 소재인 경우 (실크, 캐시미어, 가죽, 울, 린넨 등)
- 얼룩 발생 후 48시간이 초과된 경우
- 고객이 "맡기면 나을까요?" 또는 유사한 질문을 한 경우

## 말투
- 친근하지만 전문가다운 말투를 사용해라
- "이건 집에서 건드리면 거의 망가집니다" 같은 직설적 표현을 써라
- 경험에서 우러나온 조언처럼 말해라`

// SendMessage is a convenience method for single user messages
func (c *OpenAIClient) SendMessage(userMessage string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: OpenAILaundryMasterPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	return c.Chat(messages)
}

// SendMessageWithContext sends a message with context from Pinecone
func (c *OpenAIClient) SendMessageWithContext(userMessage, context string) (string, error) {
	systemPrompt := OpenAILaundryMasterPrompt

	if context != "" {
		systemPrompt += "\n\n## 관련 세탁 지식 (Pinecone 검색 결과):\n" + context
	}

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	return c.Chat(messages)
}
