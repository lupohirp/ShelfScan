package embedding

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

type GeminiEmbedRequest struct {
	Content              GeminiContent `json:"content"`
	OutputDimensionality int           `json:"outputDimensionality"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

type GeminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func (e *EmbeddingClient) GetEmbedding(file io.Reader, filename string) ([]float32, error) {
	imgData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}
	return e.GetEmbeddingBytes(imgData, filename)
}

func (e *EmbeddingClient) GetEmbeddingBytes(imgData []byte, filename string) ([]float32, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("gemini API key is not configured for EmbeddingClient")
	}

	encoded := base64.StdEncoding.EncodeToString(imgData)
	mimeType := detectMimeType(filename)

	reqBody := GeminiEmbedRequest{
		Content: GeminiContent{
			Parts: []GeminiPart{
				{
					InlineData: &GeminiInlineData{
						MimeType: mimeType,
						Data:     encoded,
					},
				},
			},
		},
		OutputDimensionality: 768,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-2:embedContent?key=%s", e.apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP post to Gemini: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API returned status %s: %s", resp.Status, string(bodyBytes))
	}

	var geminiResp GeminiEmbedResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return geminiResp.Embedding.Values, nil
}
