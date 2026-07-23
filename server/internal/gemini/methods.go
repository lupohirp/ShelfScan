package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

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

	// Priority list of fallback Flash models (Flash-Lite prioritized for high RPM free tier limits)
	baseModels := []string{
		"models/gemini-3.5-flash-lite",
		"models/gemini-3.1-flash-lite",
		"models/gemini-3.6-flash",
		"models/gemini-3.5-flash",
		"models/gemini-3-flash",
	}

	configuredModel := g.GenerativeModel
	if configuredModel == "" || strings.Contains(configuredModel, "gemma") {
		configuredModel = "models/gemini-3.5-flash-lite"
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
		if strings.Contains(err.Error(), "free_tier_requests") || strings.Contains(err.Error(), "limit: 500") {
			return nil, "", fmt.Errorf("quota giornaliera gratuita Gemini superata (limite 500 richieste/giorno): riprova domani o inserisci un'API key con fatturazione abilitata")
		}
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

	prompt := fmt.Sprintf(`Compara l'Immagine 1 (ritaglio scattato dal vivo in negozio) con l'Immagine 2 (foto promozionale di catalogo su sfondo bianco).

PRIMA DI RISPONDERE match: true, DEVI VERIFICARE CRUCIALMENTE QUESTI 5 PUNTI ANATOMICI:
1. LOGO SUL QUADRANTE: Scritta piatta "LIU JO" vs Lettere tridimensionali 3D "L I U J O" vs Monogramma "LJ" vs Cristalli. Se differisce -> match: false.
2. LUNETTA E STRASS: Strass a 360° su tutta la circonferenza vs Solo a Mezzaluna (parte inferiore) vs Lunetta liscia o zigrinata. Se differisce -> match: false.
3. DATARIO E CRONOGRAFO: Presenza di 2 o 3 sotto-quadranti piccoli o finestrella del datario. Se differisce -> match: false.
4. CINTURINO E ATTACCO: Maglie 3-file classiche vs Maglia milanese (mesh) vs Catena rigida con strass/maglie logo vs Bracciale fisso. Se differisce -> match: false.
5. NUMERI ED INDICI: Romani (XII, VI) vs Arabi (12, 6) vs Stanghette semplici vs Brillanti tondi. Se differisce -> match: false.

REGOLE DI VALUTAZIONE ESTREMAMENTE SEVERE:
- Rispondi match: false SE NOTI ANCHE UN SOLO DETTAGLIO STRUTTURALE DIFFERENTE.
- Se hai QUALSIASI DUBBIO o i dettagli non sono nitidi al 100%%, rispondi match: false.
- Rispondi match: true SOLO se l'Immagine 1 è il 100%% IDENTICO GEMELLO dell'Immagine 2 in ogni singolo punto.

%s`, categoryPrompt)

	modelsToTry := []string{
		"models/gemini-3.5-flash-lite",
		"models/gemini-3.1-flash-lite",
		"models/gemini-3.6-flash",
		"models/gemini-2.5-flash",
	}

	var lastErr error
	for _, modelName := range modelsToTry {
		log.Printf("Attempting verification using model: %s", modelName)
		model := g.client.GenerativeModel(modelName)
		model.SetTemperature(0.0)
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text("Sei uno spietato ispettore di controllo qualità per orologeria. Il tuo unico scopo è scartare i falsi positivi. Se le due immagini non raffigurano il GEMELLO 100% IDENTICO dello stesso ed unico modello in ogni singolo dettaglio visivo, DEVI rispondere match: false."),
			},
		}

		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"analisi": {
					Type:        genai.TypeString,
					Description: "Fai un'analisi dettagliata passo-passo dei due articoli. Confronta la forma della cassa, il quadrante, la presenza/assenza di cronografi, numeri, indici, brillanti e cinturino. Identifica esplicitamente eventuali differenze minime.",
				},
				"match": {
					Type:        genai.TypeBoolean,
					Description: "Imposta su true SOLO SE l'analisi conferma che sono ESATTAMENTE lo stesso modello. Altrimenti false.",
				},
			},
			Required: []string{"analisi", "match"},
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
					Match   bool   `json:"match"`
					Analisi string `json:"analisi"`
				}

				var res VerificationResult
				if err := json.Unmarshal([]byte(responseText), &res); err == nil {
					return res.Match, res.Analisi, nil
				} else {
					log.Printf("Failed to parse verification json from model %s: %v", modelName, err)
					lastErr = err
				}
			} else {
				lastErr = fmt.Errorf("empty response from model %s", modelName)
			}
		} else {
			log.Printf("Verification model %s failed: %v, falling back immediately to next model...", modelName, err)
			lastErr = err
		}
	}

	return false, "", fmt.Errorf("all verification models failed. Last error: %w", lastErr)
}

type VerificationCandidate struct {
	Sku     string
	Name    string
	ImgData []byte
}

func (g *GeminiClient) VerifyMatchCandidates(ctx context.Context, cropImg []byte, candidates []VerificationCandidate, category string) (string, string, error) {
	if len(candidates) == 0 {
		return "NONE", "No candidates provided", nil
	}
	if len(candidates) == 1 {
		match, reason, err := g.VerifyMatch(ctx, cropImg, candidates[0].ImgData, category)
		if err != nil {
			return "NONE", "", err
		}
		if match {
			return candidates[0].Sku, reason, nil
		}
		return "NONE", reason, nil
	}

	g.mu.Lock()
	if g.client == nil {
		c, err := genai.NewClient(context.Background(), option.WithAPIKey(g.Apikey))
		if err != nil {
			g.mu.Unlock()
			return "NONE", "", err
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
	default:
		categoryPrompt = `Presta attenzione a dettagli come la forma complessiva del gioiello, il tipo di metallo (oro, argento, ecc.), la presenza e disposizione di pietre preziose, perle o cristalli, e il design specifico del pezzo.`
	}

	candidatesDescription := ""
	parts := []genai.Part{}
	parts = append(parts, genai.ImageData("jpeg", cropImg))

	for i, cand := range candidates {
		idx := i + 1
		candidatesDescription += fmt.Sprintf("- CANDIDATO (Immagine %d): SKU %s\n", idx+1, cand.Sku)
		parts = append(parts, genai.ImageData("jpeg", cand.ImgData))
	}

	promptText := fmt.Sprintf(`Compara l'Immagine 1 (ritaglio scattato dal vivo in negozio) con le immagini di catalogo dei candidati (Immagini da 2 a %d).

CANDIDATI UFFICIALI:
%s

REGOLE DI VALUTAZIONE ESTREMAMENTE RIGIDE (IL TUO OBIETTIVO È SCARTARE I FALSI POSITIVI):
1. La maggior parte dei ritagli NON corrisponde a nessuno dei candidati forniti. Il valore PREDEFINITO deve essere "matched_sku": "NONE".
2. Gli orologi si assomigliano molto: cerca ogni minimo dettaglio (es. presenza/assenza di cronografi, datari, brillantini sulla ghiera, forma degli indici, trama del cinturino).
3. Se un candidato ha anche UN SOLO DETTAGLIO DIVERSO o se hai QUALSIASI DUBBIO, quel candidato NON corrisponde. Imposta "matched_sku": "NONE".
4. Imposta "matched_sku" allo SKU del candidato (es. 'TLJ2364') SOLO ED ESCLUSIVAMENTE se l'Immagine 1 è il 100%% IDENTICO GEMELLO di uno dei candidati di catalogo.

%s`, len(candidates)+1, candidatesDescription, categoryPrompt)

	parts = append([]genai.Part{genai.Text(promptText)}, parts...)

	modelsToTry := []string{
		"models/gemini-3.5-flash-lite",
		"models/gemini-3.1-flash-lite",
		"models/gemini-3.6-flash",
		"models/gemini-2.5-flash",
	}

	var lastErr error
	for _, modelName := range modelsToTry {
		model := g.client.GenerativeModel(modelName)
		model.SetTemperature(0.0)
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text(`Sei uno spietato ispettore di controllo qualità per orologeria e gioielleria. Il tuo unico scopo è scartare i falsi positivi.

PRIMA DI DICHIARARE UNA CORRISPONDENZA ("matched_sku"), DEVI RIGOROSAMENTE VERIFICARE QUESTI 5 PUNTI ANATOMICI TRA L'IMMAGINE 1 E CIASCUN CANDIDATO:
1. LOGO SUL QUADRANTE: È una scritta stampata piatta "LIU JO", o sono grandi lettere tridimensionali "L I U J O", o è il monogramma "LJ", o è in pave di cristalli? Se il tipo di logo differisce, SCARTA TASSATIVAMENTE (matched_sku: "NONE").
2. LUNETTA E STRASS: Gli strass coprono l'INTERA circonferenza (360°), oppure sono solo sulla parte inferiore (a mezzaluna), oppure la lunetta è liscia/zigrinata senza strass? Se la disposizione degli strass differisce, SCARTA TASSATIVAMENTE.
3. DATARIO E CRONOGRAFO: Ci sono sotto-quadranti piccoli (2 o 3 cerchi) o la finestrella del datario? Se presenti in una e assenti nell'altra, SCARTA TASSATIVAMENTE.
4. CINTURINO E ATTACCO: È un bracciale a maglie classiche 3-file, a maglia milanese/mesh, a catena rigida con strass/maglie logo, o a bracciale fisso? Se la trama del cinturino differisce, SCARTA TASSATIVAMENTE.
5. NUMERI ED INDICI: Ci sono numeri romani (XII, VI), numeri arabi (12, 6), o solo stanghette/brillanti? Se il tipo di indici differisce, SCARTA TASSATIVAMENTE.

SE ANCHE UN SOLO PUNTO DEI 5 DIFFERISCE O NON È CHIARAMENTE VISIBILE, DEVI RESTITUIR "matched_sku": "NONE". Non scegliere mai un candidato se non è il GEMELLO 100% IDENTICO.`),
			},
		}

		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"analisi": {
					Type:        genai.TypeString,
					Description: "Analisi dettagliata del confronto tra l'Immagine 1 e ciascuno dei candidati.",
				},
				"matched_sku": {
					Type:        genai.TypeString,
					Description: "Lo SKU esatto del candidato che corrisponde al 100% all'Immagine 1 (es. 'TLJ2364'). Restituisci 'NONE' se nessuno dei candidati corrisponde esattamente.",
				},
			},
			Required: []string{"analisi", "matched_sku"},
		}

		resp, err := model.GenerateContent(ctx, parts...)
		if err == nil {
			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
				var responseText string
				for _, part := range resp.Candidates[0].Content.Parts {
					if text, ok := part.(genai.Text); ok {
						responseText += string(text)
					}
				}

				type MultiVerificationResult struct {
					MatchedSku string `json:"matched_sku"`
					Analisi    string `json:"analisi"`
				}

				var res MultiVerificationResult
				if err := json.Unmarshal([]byte(responseText), &res); err == nil {
					return res.MatchedSku, res.Analisi, nil
				} else {
					log.Printf("Failed to parse multi verification json from model %s: %v", modelName, err)
					lastErr = err
				}
			} else {
				lastErr = fmt.Errorf("empty response from model %s", modelName)
			}
		} else {
			log.Printf("Multi verification model %s failed: %v", modelName, err)
			lastErr = err
			if strings.Contains(err.Error(), "429") || strings.Contains(strings.ToLower(err.Error()), "quota") {
				time.Sleep(2 * time.Second)
			}
		}
	}

	return "NONE", "", fmt.Errorf("all multi verification models failed. Last error: %w", lastErr)
}
