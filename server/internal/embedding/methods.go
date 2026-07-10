package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (e *EmbeddingClient) GetEmbedding(file io.Reader, filename string) ([]float32, error) {
	imgData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}
	return e.GetEmbeddingBytes(imgData, filename)
}

func (e *EmbeddingClient) GetEmbeddingBytes(imgData []byte, _ string) ([]float32, error) {
	url, err := e.embedURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("failed to POST to embeddings sidecar: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sidecar response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embeddings sidecar returned %s: %s", resp.Status, string(body))
	}

	var parsed embedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse sidecar response: %w", err)
	}
	if len(parsed.Embedding) == 0 {
		return nil, fmt.Errorf("embeddings sidecar returned empty vector")
	}
	return parsed.Embedding, nil
}

func (e *EmbeddingClient) embedURL() (string, error) {
	if e.host == "" {
		return "", fmt.Errorf("EMBEDDINGS_HOST is not configured")
	}
	base := strings.TrimRight(e.host, "/")
	if e.port != "" {
		base = fmt.Sprintf("%s:%s", base, e.port)
	}
	return base + "/embed", nil
}
