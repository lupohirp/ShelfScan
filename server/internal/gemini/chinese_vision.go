package gemini

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
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
	model := os.Getenv("VISION_MODEL")

	return &ChineseVisionClient{
		APIKey:  key,
		BaseURL: strings.TrimRight(url, "/"),
		Model:   model,
		client: &http.Client{
			Timeout: 20 * time.Second,
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
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
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

func getZhipuAuthHeader(apiKey string) string {
	lastDot := strings.LastIndex(apiKey, ".")
	if lastDot <= 0 || lastDot == len(apiKey)-1 {
		return apiKey
	}
	id := apiKey[:lastDot]
	secret := apiKey[lastDot+1:]

	now := time.Now().UnixNano() / 1e6
	exp := now + 3600000

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","sign_type":"SIGN"}`))
	payloadStr := fmt.Sprintf(`{"api_key":"%s","exp":%d,"timestamp":%d}`, id, exp, now)
	payload := base64.RawURLEncoding.EncodeToString([]byte(payloadStr))

	tokenToSign := header + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(tokenToSign))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return tokenToSign + "." + sig
}

type EndpointConfig struct {
	BaseURL string
	Model   string
	Auth    string
}

func (c *ChineseVisionClient) sendCompletionRequest(ctx context.Context, reqBody ChatCompletionRequest) ([]byte, error) {
	// QwenCloud (DashScope) primary, with multi-provider failover
	endpoints := []EndpointConfig{
		{BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Model: "qwen2.5-vl-7b-instruct", Auth: "bearer"},
		{BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Model: "qwen-vl-max", Auth: "bearer"},
		{BaseURL: "https://dashscope-intl.aliyuncs.com/compatible-mode/v1", Model: "qwen2.5-vl-7b-instruct", Auth: "bearer"},
		{BaseURL: "https://dashscope-intl.aliyuncs.com/compatible-mode/v1", Model: "qwen-vl-max", Auth: "bearer"},
		{BaseURL: "https://api.siliconflow.cn/v1", Model: "Pro/Qwen/Qwen2.5-VL-7B-Instruct", Auth: "bearer"},
		{BaseURL: "https://open.bigmodel.cn/api/paas/v4", Model: "glm-4v-flash", Auth: "bearer"},
		{BaseURL: "https://openrouter.ai/api/v1", Model: "qwen/qwen-2.5-vl-72b-instruct:free", Auth: "bearer"},
	}

	if c.BaseURL != "" {
		epModel := c.Model
		if epModel == "" {
			epModel = "qwen2.5-vl-7b-instruct"
		}
		endpoints = append([]EndpointConfig{{BaseURL: c.BaseURL, Model: epModel, Auth: "bearer"}}, endpoints...)
	}

	var lastErr error
	for _, ep := range endpoints {
		currentReq := reqBody
		currentReq.Model = ep.Model
		jsonBytes, err := json.Marshal(currentReq)
		if err != nil {
			continue
		}

		authToken := c.APIKey
		if ep.Auth == "jwt" {
			authToken = getZhipuAuthHeader(c.APIKey)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", ep.BaseURL+"/chat/completions", bytes.NewReader(jsonBytes))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Success! QwenCloud / Vision API responded from %s using model %s", ep.BaseURL, ep.Model)
			return bodyBytes, nil
		}

		lastErr = fmt.Errorf("Vision API status %d from %s (%s): %s", resp.StatusCode, ep.BaseURL, ep.Model, string(bodyBytes))
		log.Printf("Endpoint %s (%s) status %d, trying next provider...", ep.BaseURL, ep.Model, resp.StatusCode)
	}

	return nil, lastErr
}

func (c *ChineseVisionClient) GenerateContent(ctx context.Context, prompt string, imgData []byte) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("QwenCloud Vision API Key is empty")
	}

	b64Img := base64.StdEncoding.EncodeToString(imgData)
	dataURL := "data:image/jpeg;base64," + b64Img

	reqBody := ChatCompletionRequest{
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

	bodyBytes, err := c.sendCompletionRequest(ctx, reqBody)
	if err != nil {
		return "", err
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", err
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("QwenCloud API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty response choices from QwenCloud API")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *ChineseVisionClient) VerifyMatch(ctx context.Context, cropImg, dbImg []byte, category string) (bool, string, error) {
	if c.APIKey == "" {
		return false, "", fmt.Errorf("QwenCloud Vision API Key is empty")
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

	bodyBytes, err := c.sendCompletionRequest(ctx, reqBody)
	if err != nil {
		return false, "", err
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return false, "", err
	}

	if len(chatResp.Choices) == 0 {
		return false, "", fmt.Errorf("empty response choices from QwenCloud API")
	}

	text := chatResp.Choices[0].Message.Content

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

	log.Printf("Raw text from QwenCloud API: %s", text)
	return false, "", fmt.Errorf("failed to parse JSON from QwenCloud API response")
}
