package vector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PineconeClient struct {
	apiKey      string
	environment string
	indexName   string
	httpClient  *http.Client
	indexHost   string
}

type Vector struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type UpsertRequest struct {
	Vectors   []Vector `json:"vectors"`
	Namespace string   `json:"namespace,omitempty"`
}

type QueryRequest struct {
	Vector          []float32              `json:"vector"`
	TopK            int                    `json:"topK"`
	IncludeMetadata bool                   `json:"includeMetadata"`
	Namespace       string                 `json:"namespace,omitempty"`
	Filter          map[string]interface{} `json:"filter,omitempty"`
}

type QueryMatch struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Values   []float32              `json:"values,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type QueryResponse struct {
	Matches []QueryMatch `json:"matches"`
}

func NewPineconeClient(apiKey, environment, indexName string) *PineconeClient {
	return &PineconeClient{
		apiKey:      apiKey,
		environment: environment,
		indexName:   indexName,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		// Pinecone index host format: https://INDEX_NAME-PROJECT_ID.svc.ENVIRONMENT.pinecone.io
		indexHost: fmt.Sprintf("https://%s.svc.%s.pinecone.io", indexName, environment),
	}
}

func (c *PineconeClient) Upsert(vectors []Vector, namespace string) error {
	url := fmt.Sprintf("%s/vectors/upsert", c.indexHost)

	reqBody := UpsertRequest{
		Vectors:   vectors,
		Namespace: namespace,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pinecone API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *PineconeClient) Query(vector []float32, topK int, namespace string, filter map[string]interface{}) (*QueryResponse, error) {
	url := fmt.Sprintf("%s/query", c.indexHost)

	reqBody := QueryRequest{
		Vector:          vector,
		TopK:            topK,
		IncludeMetadata: true,
		Namespace:       namespace,
		Filter:          filter,
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
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pinecone API error (status %d): %s", resp.StatusCode, string(body))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &queryResp, nil
}

func (c *PineconeClient) Delete(ids []string, namespace string) error {
	url := fmt.Sprintf("%s/vectors/delete", c.indexHost)

	reqBody := map[string]interface{}{
		"ids":       ids,
		"namespace": namespace,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pinecone API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
