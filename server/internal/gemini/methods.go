package gemini

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func (g *GeminiClient) GetClient(ctx context.Context) *genai.GenerativeModel {

	//apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(g.Apikey))
	if err != nil {
		log.Printf("Error creating GenAI client: %v", err)
	}

	defer client.Close()

	model := client.GenerativeModel(g.GenerativeModel)
	model.SetTemperature(float32(g.Temperature))
	model.SetMaxOutputTokens(int32(g.MaxOutputTokens))

	return model
}
