package gemini

import (
	"sync"

	"github.com/google/generative-ai-go/genai"
)

type GeminiClient struct {
	GenerativeModel string
	Apikey          string
	Prompt          string

	Temperature     float64
	MaxOutputTokens int

	mu     sync.Mutex
	client *genai.Client
}

func NewGeminiClient() *GeminiClient {
	return &GeminiClient{}
}

func (g *GeminiClient) WithGenerativeModel(model string) *GeminiClient {
	g.GenerativeModel = model
	return g
}

func (g *GeminiClient) WithApiKey(key string) *GeminiClient {
	g.Apikey = key
	return g

}

func (g *GeminiClient) WithTemperature(temp float64) *GeminiClient {
	g.Temperature = temp
	return g
}

func (g *GeminiClient) WithMaxOutputTokens(tokens int) *GeminiClient {
	g.MaxOutputTokens = tokens
	return g

}

func (g *GeminiClient) WithPrompt(prompt string) *GeminiClient {
	g.Prompt = prompt
	return g
}
