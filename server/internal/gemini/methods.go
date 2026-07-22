package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

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
		"models/gemini-3.6-flash",
		"models/gemini-3.1-flash-lite",
		"models/gemini-3.5-flash-lite",
		"models/gemini-3-flash",
	}

	configuredModel := g.GenerativeModel
	if configuredModel == "" {
		configuredModel = "models/gemini-3.6-flash"
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

func (g *GeminiClient) VerifyMatch(ctx context.Context, cropImg []byte, dbImg []byte, category string) (bool, string, error) {
	g.mu.Lock()
	if g.client == nil {
		c, err := genai.NewClient(context.Background(), option.WithAPIKey(g.Apikey))
		if err != nil {
			g.mu.Unlock()
			return false, "", err
		}
		g.client = c
	}
	g.mu.Unlock()

	categoryPrompt := ""
	switch strings.ToLower(category) {
	case "watch", "orologio", "cronografo", "smartwatch":
		categoryPrompt = `Per gli orologi, presta attenzione a dettagli come:
- La presenza o meno di datari, sotto-quadranti o cronografi.
- I numeri sul quadrante (romani, arabi, o semplici trattini/brillanti).
- Il colore esatto del quadrante e del cinturino.
- La trama della maglia del cinturino.
- La forma esatta della lunetta (liscia, zigrinata, con brillanti, ecc.).`
	case "ring", "anello", "anelli", "fede", "fedina", "solitario", "veretta":
		categoryPrompt = `Per gli anelli, presta attenzione a dettagli come:
- Il tipo di anello (solitario, veretta, trilogy, fascia, fede).
- La forma, il taglio e il colore della pietra centrale (rotondo, ovale, smeraldo, ecc.).
- La presenza di piccoli diamanti sulla fascia.
- Il colore e materiale del metallo (oro bianco, oro giallo, oro rosa, argento).`
	case "earring", "orecchini", "orecchino", "lobo", "monachella":
		categoryPrompt = `Per gli orecchini, presta attenzione a dettagli come:
- Il tipo di orecchino (a lobo, pendente, a cerchio, monachella).
- La forma geometrica e la simmetria del design.
- La disposizione di pietre, brillanti o perle.
- Il colore e materiale del metallo.`
	case "necklace", "collana", "pendant", "pendente", "girocollo", "ciondolo", "collier", "catena", "catenina":
		categoryPrompt = `Per le collane e ciondoli, presta attenzione a dettagli come:
- La forma, dimensione e disegno del ciondolo/pendente.
- Lo spessore e il tipo di maglia della catena (sottile, spessa, torchon, traversino).
- La presenza di pietre o incisioni specifiche.
- Il colore e materiale del metallo.`
	case "bracelet", "bracciale", "braccialetto", "bangle":
		categoryPrompt = `Per i bracciali, presta attenzione a dettagli come:
- Il tipo di bracciale (rigido/bangle, a catena, tennis, multifilo).
- Lo stile delle maglie e la chiusura.
- La presenza di charm, pietre incastonate o pendenti.
- Il colore e materiale del metallo.`
	default:
		categoryPrompt = `Presta attenzione a dettagli come la forma complessiva del gioiello, il tipo di metallo (oro, argento, ecc.), la presenza e disposizione di pietre preziose, perle o cristalli, e il design specifico del pezzo.`
	}

	prompt := fmt.Sprintf(`Compara queste due immagini per determinare se raffigurano lo STESSO MODELLO di prodotto.

L'Immagine 1 è un ritaglio (crop) scattato dal vivo in negozio. Può presentare riflessi di luce, ombre, angolazioni diverse o leggera distorsione prospettica.
L'Immagine 2 è la foto promozionale di catalogo su sfondo bianco.

REGOLE DI VALUTAZIONE:
1. NON confondere differenze di illuminazione, riflessi sul vetro, ombre o angolazione con differenze di prodotto!
2. Rispondi match: true se l'articolo nella foto dal vivo (Immagine 1) è LO STESSO MODELLO rappresentato nel catalogo (Immagine 2), anche se fotografato in condizioni di luce o angolazioni diverse.
3. Rispondi match: false SOLO se noti chiare differenze strutturali di design (es. un modello diverso di gioiello, forma della cassa diversa, quadrante con cronografi vs quadrante semplice, pietre mancanti o posizionate in modo nettamente diverso).

%s

Se si tratta dello stesso modello di prodotto, rispondi con match: true.`, categoryPrompt)

	modelsToTry := []string{
		"models/gemini-3.6-flash",
		"models/gemini-3.1-flash-lite",
		"models/gemini-3.5-flash-lite",
		"models/gemini-3-flash",
	}

	var lastErr error
	for _, modelName := range modelsToTry {
		log.Printf("Attempting verification using model: %s", modelName)
		model := g.client.GenerativeModel(modelName)
		model.SetTemperature(0.1)

		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"match": {
					Type:        genai.TypeBoolean,
					Description: "Imposta su true se le due immagini mostrano lo STESSO IDENTICO modello di orologio o gioiello. Imposta su false se sono modelli differenti (anche se simili).",
				},
				"motivo": {
					Type:        genai.TypeString,
					Description: "Una breve spiegazione in italiano che descrive le differenze rilevate o i punti di corrispondenza che confermano l'identità del modello.",
				},
			},
			Required: []string{"match", "motivo"},
		}

		resp, err := model.GenerateContent(ctx,
			genai.Text(prompt),
			genai.ImageData("jpeg", cropImg),
			genai.ImageData("jpeg", dbImg),
		)
		if err == nil {
			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
				var responseText string
				for _, part := range resp.Candidates[0].Content.Parts {
					if text, ok := part.(genai.Text); ok {
						responseText += string(text)
					}
				}

				type VerificationResult struct {
					Match  bool   `json:"match"`
					Motivo string `json:"motivo"`
				}

				var res VerificationResult
				if err := json.Unmarshal([]byte(responseText), &res); err == nil {
					return res.Match, res.Motivo, nil
				} else {
					log.Printf("Failed to parse verification json from model %s: %v", modelName, err)
					lastErr = err
				}
			} else {
				lastErr = fmt.Errorf("empty response from model %s", modelName)
			}
		} else {
			log.Printf("Verification model %s failed: %v, skipping...", modelName, err)
			lastErr = err
		}
	}

	return false, "", fmt.Errorf("all verification models failed. Last error: %w", lastErr)
}


