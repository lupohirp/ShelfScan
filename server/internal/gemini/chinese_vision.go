package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ChineseVisionClient struct {
	APIKey  string
	BaseURL string
	Model   string
	client  *http.Client
}

func NewChineseVisionClient() *ChineseVisionClient {
	key := os.Getenv("ZHIPU_API_KEY")
	if key == "" {
		key = os.Getenv("VISION_API_KEY")
	}
	if key == "" {
		key = "sk-ws-H.XLLLDL.cgeQ.MEUCICid6D1La3nwxvachHZGdlu3N4qu61WUU6tqUkEygwb0AiEA1M_AMef-n4bTfmaYUu9s_9UIy6woh2mqPhZxK7XRoeE"
	}

	url := os.Getenv("VISION_BASE_URL")
	if url == "" {
		url = "https://open.bigmodel.cn/api/paas/v4"
	}

	model := os.Getenv("VISION_MODEL")
	if model == "" {
		model = "glm-4v-flash"
	}

	return &ChineseVisionClient{
		APIKey:  key,
		BaseURL: strings.TrimRight(url, "/"),
		Model:   model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ChatMessageContentPart struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL *ChatMessageImage `json:"image_url,omitempty"`
}

type ChatMessageImage struct {
	URL string `json:"url"`
}

type ChatMessage struct {
	Role    string                   `json:"role"`
	Content []ChatMessageContentPart `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Temperature float64        `json:"temperature"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (c *ChineseVisionClient) GenerateContent(ctx context.Context, prompt string, imgData []byte) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("Chinese Vision API Key is empty")
	}

	b64Img := base64.StdEncoding.EncodeToString(imgData)
	dataURL := "data:image/jpeg;base64," + b64Img

	reqBody := ChatCompletionRequest{
		Model: c.Model,
		Messages: []ChatMessage{
			{
				Role: "user",
				Content: []ChatMessageContentPart{
					{Type: "text", Text: prompt},
					{Type: "image_url", ImageURL: &ChatMessageImage{URL: dataURL}},
				},
			},
		},
		Temperature: 0.1,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Chinese Vision API status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", err
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("Chinese Vision API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty response choices from Chinese Vision API")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *ChineseVisionClient) VerifyMatch(ctx context.Context, cropImg, dbImg []byte, category string) (bool, string, error) {
	if c.APIKey == "" {
		return false, "", fmt.Errorf("Chinese Vision API Key is empty")
	}

	b64Crop := base64.StdEncoding.EncodeToString(cropImg)
	b64Db := base64.StdEncoding.EncodeToString(dbImg)

	categoryPrompt := ""
	if category != "" {
		categoryPrompt = fmt.Sprintf("Categoria prodotto rilevata: %s.", category)
	}

	prompt := fmt.Sprintf(`Sei un esperto di controllo qualità in gioielleria.
Confronta le seguenti due immagini:
L'Immagine 1 è un ritaglio scattato dal vivo in negozio.
L'Immagine 2 è la foto promozionale di catalogo su sfondo bianco.

REGOLE DI VALUTAZIONE:
1. NON confondere differenze di illuminazione, riflessi sul vetro, ombre o angolazione con differenze di prodotto!
2. Rispondi match: true se l'articolo nella foto dal vivo (Immagine 1) è LO STESSO MODELLO rappresentato nel catalogo (Immagine 2).
3. Rispondi match: false SOLO se noti chiare differenze strutturali di design (es. modello diverso, cassa diversa, cronografi vs semplice).

%s

Rispondi TASSATIVAMENTE in formato JSON valido:
{"match": true, "motivo": "Spiegazione in italiano"}`, categoryPrompt)

	reqBody := ChatCompletionRequest{
		Model: c.Model,
		Messages: []ChatMessage{
			{
				Role: "user",
				Content: []ChatMessageContentPart{
					{Type: "text", Text: prompt},
					{Type: "image_url", ImageURL: &ChatMessageImage{URL: "data:image/jpeg;base64," + b64Crop}},
					{Type: "image_url", ImageURL: &ChatMessageImage{URL: "data:image/jpeg;base64," + b64Db}},
				},
			},
		},
		Temperature: 0.1,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return false, "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewReader(jsonBytes))
	if err != nil {
		return false, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("Chinese Vision API status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return false, "", err
	}

	if len(chatResp.Choices) == 0 {
		return false, "", fmt.Errorf("empty response choices from Chinese Vision API")
	}

	text := chatResp.Choices[0].Message.Content

	// Clean JSON fences if present
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		idx := strings.Index(text, "\n")
		if idx != -1 {
			text = text[idx+1:]
		}
		if lastIdx := strings.LastIndex(text, "```"); lastIdx != -1 {
			text = text[:lastIdx]
		}
		text = strings.TrimSpace(text)
	}

	type VerificationResult struct {
		Match  bool   `json:"match"`
		Motivo string `json:"motivo"`
	}

	var res VerificationResult
	if err := json.Unmarshal([]byte(text), &res); err == nil {
		return res.Match, res.Motivo, nil
	}

	log.Printf("Raw text from Chinese Vision API: %s", text)
	return false, "", fmt.Errorf("failed to parse JSON from Chinese Vision API response")
}
