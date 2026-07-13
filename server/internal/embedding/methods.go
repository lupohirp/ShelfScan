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

type geminiEmbedRequest struct {
	Content              geminiContent `json:"content"`
	OutputDimensionality int           `json:"outputDimensionality"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func detectMimeType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
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
		return nil, fmt.Errorf("GEMINI_API_KEY is not configured for EmbeddingClient")
	}

	reqBody := geminiEmbedRequest{
		Content: geminiContent{
			Parts: []geminiPart{{
				InlineData: &geminiInlineData{
					MimeType: detectMimeType(filename),
					Data:     base64.StdEncoding.EncodeToString(imgData),
				},
			}},
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
		return nil, fmt.Errorf("failed to POST to Gemini: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gemini response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini returned %s: %s", resp.Status, string(body))
	}

	var parsed geminiEmbedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}
	if len(parsed.Embedding.Values) == 0 {
		return nil, fmt.Errorf("Gemini returned empty vector")
	}
	return parsed.Embedding.Values, nil
}
