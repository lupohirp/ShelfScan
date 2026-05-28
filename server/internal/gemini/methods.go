package gemini

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func (g *GeminiClient) GetClient(ctx context.Context) *genai.GenerativeModel {
	return g.GetClientForModel(ctx, g.GenerativeModel)
}

func (g *GeminiClient) GetClientForModel(ctx context.Context, modelName string) *genai.GenerativeModel {
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

	model := g.client.GenerativeModel(modelName)
	model.SetTemperature(float32(g.Temperature))
	if g.MaxOutputTokens > 0 {
		model.SetMaxOutputTokens(int32(g.MaxOutputTokens))
	}

	// Enable Structured JSON Schema for reliable output structure
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeArray,
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"desc": {
					Type:        genai.TypeString,
					Description: "A short description of the jewelry item including color and material.",
				},
				"box": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeInteger,
					},
					Description: "Bounding box coordinates: [ymin, xmin, ymax, xmax] (normalized 0 to 1000).",
				},
			},
			Required: []string{"desc", "box"},
		},
	}

	return model
}

func (g *GeminiClient) GenerateContentWithFallback(ctx context.Context, prompt string, imgData []byte) (*genai.GenerateContentResponse, string, error) {
	// Priority list of models to rotate/fall back to
	baseModels := []string{
		"models/gemini-3.5-flash",
		"models/gemini-3.1-flash-lite",
		"models/gemini-3-flash",
		"models/gemini-2.5-flash",
		"models/gemini-2.5-flash-lite",
	}

	configuredModel := g.GenerativeModel
	if configuredModel == "" {
		configuredModel = "models/gemini-2.5-flash"
	}

	modelsToTry := []string{configuredModel}
	for _, m := range baseModels {
		// Avoid duplicate configuredModel
		if m != configuredModel {
			modelsToTry = append(modelsToTry, m)
		}
	}

	var lastErr error
	for _, modelName := range modelsToTry {
		log.Printf("Attempting content generation using model: %s", modelName)
		model := g.GetClientForModel(ctx, modelName)
		if model == nil {
			log.Printf("Failed to initialize model %s, skipping...", modelName)
			continue
		}

		resp, err := model.GenerateContent(ctx, genai.Text(prompt), genai.ImageData("jpeg", imgData))
		if err == nil {
			log.Printf("Success! Generated content using model: %s", modelName)
			return resp, modelName, nil
		}

		log.Printf("Model %s failed with error: %v", modelName, err)
		lastErr = err
	}

	return nil, "", lastErr
}
