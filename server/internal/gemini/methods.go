package gemini

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func (g *GeminiClient) GetClient(ctx context.Context) *genai.GenerativeModel {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.client == nil {
		c, err := genai.NewClient(context.Background(), option.WithAPIKey(g.Apikey))
		if err != nil {
			log.Printf("Error creating GenAI client: %v", err)
			return nil
		}
		g.client = c
	}

	model := g.client.GenerativeModel(g.GenerativeModel)
	model.SetTemperature(float32(g.Temperature))
	model.SetMaxOutputTokens(int32(g.MaxOutputTokens))

	return model
}
