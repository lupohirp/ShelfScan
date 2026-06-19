package gemini

import (
	"context"
	"fmt"
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

	// Enable JSON mode and define the structured response schema.
	// Requiring explanation ("spiegazione") and empty check ("scaffale_vuoto") before list of items ("articoli")
	// helps prevent the model from getting stuck in hallucination loops on empty shelves.
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"spiegazione": {
				Type:        genai.TypeString,
				Description: "Analisi visiva dell'espositore. Spiega dettagliatamente se vedi articoli reali o solo supporti/cuscinetti vuoti.",
			},
			"scaffale_vuoto": {
				Type:        genai.TypeBoolean,
				Description: "Imposta su true se lo scaffale/espositore è completamente vuoto o contiene solo supporti/espositori vuoti. Imposta su false se ci sono oggetti reali (orologi, gioielli).",
			},
			"articoli": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"desc": {
							Type:        genai.TypeString,
							Description: "Una breve descrizione in italiano dell'articolo di gioielleria o orologio contenente colore e materiale.",
						},
						"box": {
							Type: genai.TypeArray,
							Items: &genai.Schema{
								Type: genai.TypeInteger,
							},
							Description: "Coordinate del bounding box: [ymin, xmin, ymax, xmax] (normalizzato da 0 a 1000).",
						},
					},
					Required: []string{"desc", "box"},
				},
				Description: "Lista degli articoli reali presenti. Deve essere vuota se scaffale_vuoto è true.",
			},
		},
		Required: []string{"spiegazione", "scaffale_vuoto", "articoli"},
	}

	return model
}

func (g *GeminiClient) GenerateContentWithFallback(ctx context.Context, prompt string, imgData []byte) (*genai.GenerateContentResponse, string, error) {
	// Priority list of models to rotate/fall back to
	baseModels := []string{
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

// DescribeImageWithModel generates a simple text description of an image using the specified model and prompt.
// It bypasses the JSON schema constraint applied to the default client model.
func (g *GeminiClient) DescribeImageWithModel(ctx context.Context, modelName string, prompt string, imgData []byte) (string, error) {
	g.mu.Lock()
	if g.client == nil {
		c, err := genai.NewClient(context.Background(), option.WithAPIKey(g.Apikey))
		if err != nil {
			g.mu.Unlock()
			return "", err
		}
		g.client = c
	}
	g.mu.Unlock()

	model := g.client.GenerativeModel(modelName)
	model.SetTemperature(0.2)
	
	resp, err := model.GenerateContent(ctx, genai.Text(prompt), genai.ImageData("jpeg", imgData))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(txt), nil
		}
	}
	return "", fmt.Errorf("no text response generated")
}

