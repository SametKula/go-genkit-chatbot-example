package vectorstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SimpleQdrantClient is a lightweight HTTP client for Qdrant.
type SimpleQdrantClient struct {
	Host string
	Port int
}

func NewSimpleQdrantClient(host string, port int) *SimpleQdrantClient {
	return &SimpleQdrantClient{
		Host: host,
		Port: port,
	}
}

func (c *SimpleQdrantClient) InitCollection(ctx context.Context, collectionName string, vectorSize int) error {
	url := fmt.Sprintf("http://%s:%d/collections/%s", c.Host, c.Port, collectionName)
	
	// Create collection
	payload := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     vectorSize,
			"distance": "Cosine",
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != 400 { // 400 could be 'collection already exists'
		return fmt.Errorf("failed to create collection: status %d", resp.StatusCode)
	}

	return nil
}

type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

func (c *SimpleQdrantClient) Upsert(ctx context.Context, collectionName string, points []Point) error {
	url := fmt.Sprintf("http://%s:%d/collections/%s/points?wait=true", c.Host, c.Port, collectionName)
	
	payload := map[string]interface{}{
		"points": points,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to upsert points: status %d", resp.StatusCode)
	}

	return nil
}

type SearchResult struct {
	Result []struct {
		ID      string                 `json:"id"`
		Score   float32                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"result"`
}

func (c *SimpleQdrantClient) Search(ctx context.Context, collectionName string, vector []float32, limit int) ([]string, error) {
	url := fmt.Sprintf("http://%s:%d/collections/%s/points/search", c.Host, c.Port, collectionName)
	
	payload := map[string]interface{}{
		"vector": vector,
		"limit":  limit,
		"with_payload": true,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to search points: status %d", resp.StatusCode)
	}

	var searchRes SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchRes); err != nil {
		return nil, err
	}

	var texts []string
	for _, res := range searchRes.Result {
		if text, ok := res.Payload["text"].(string); ok {
			texts = append(texts, text)
		}
	}
	return texts, nil
}
