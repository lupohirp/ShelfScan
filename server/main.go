package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/disintegration/imaging"
	"github.com/google/generative-ai-go/genai"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/api/option"
)

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type ProductResult struct {
	Name     string  `json:"name"`
	ImageURL string  `json:"imageUrl"`
	Score    float32 `json:"score"`
}

type MissingItem struct {
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

type Detection struct {
	Desc  string `json:"desc"`
	Box   []int  `json:"box"`
	Box2D []int  `json:"box_2d"`
}

type ImageResult struct {
	Detections []Detection `json:"detections"`
}

type AnalysisResponse struct {
	Found        []ProductResult `json:"found"`
	Missing      []MissingItem   `json:"missing"`
	ImageResults []ImageResult   `json:"imageResults"`
}

type MCPClient struct {
	conn *websocket.Conn
}

func NewMCPClient(url string) (*MCPClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &MCPClient{conn: conn}, nil
}

func (c *MCPClient) Close() {
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close()
}

func (c *MCPClient) CallVectorSearch(embedding []float32) (string, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "vector_search",
			"arguments": map[string]interface{}{
				"embedding": embedding,
			},
		},
	}
	err := c.conn.WriteJSON(req)
	if err != nil {
		return "", err
	}

	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return "", err
	}

	var res struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(message, &res); err != nil {
		return "", err
	}
	if len(res.Result.Content) == 0 {
		return "", fmt.Errorf("no content in result")
	}
	return res.Result.Content[0].Text, nil
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func fixURL(imgURL string, host string) string {
	if imgURL == "" {
		return ""
	}
	if strings.HasPrefix(imgURL, "http") {
		if strings.Contains(imgURL, "localhost") && !strings.Contains(host, "localhost") {
			return strings.Replace(imgURL, "localhost:8080", host, 1)
		}
		return imgURL
	}
	if strings.HasPrefix(imgURL, "/") {
		return "http://" + host + imgURL
	}
	return "http://" + host + "/uploads/" + imgURL
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	qdrantURL := os.Getenv("QDRANT_URL")
	if qdrantURL == "" {
		qdrantURL = "http://localhost:6333"
	}

	embeddingsURL := os.Getenv("EMBEDDINGS_URL")
	if embeddingsURL == "" {
		embeddingsURL = "http://localhost:8001"
	}

	http.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "healthy")
	}))

	http.HandleFunc("/inventory", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			items, err := listInventory(qdrantURL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, item := range items {
				payload, ok := item["payload"].(map[string]any)
				if !ok {
					continue
				}
				if img, ok := payload["imageUrl"].(string); ok {
					payload["imageUrl"] = fixURL(img, r.Host)
				}
			}
			json.NewEncoder(w).Encode(items)
			return
		}

		if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "Missing id", http.StatusBadRequest)
				return
			}
			err := deleteFromQdrant(qdrantURL, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}))

	http.HandleFunc("/search", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Missing image", http.StatusBadRequest)
			return
		}
		defer file.Close()

		embedding, err := getEmbedding(embeddingsURL, file, handler.Filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Embedding error: %v", err), http.StatusInternalServerError)
			return
		}

		results, err := performVectorSearch(qdrantURL, embedding)
		if err != nil {
			http.Error(w, fmt.Sprintf("Search error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}))

	http.HandleFunc("/upload", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(50 << 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "Missing jewelry name", http.StatusBadRequest)
			return
		}

		color := r.FormValue("color")
		material := r.FormValue("material")

		files := r.MultipartForm.File["images"]
		if len(files) == 0 {
			http.Error(w, "No images uploaded", http.StatusBadRequest)
			return
		}

		os.MkdirAll("uploads", 0755)
		var savedURL string

		var embeddings [][]float32
		for idx, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Error opening file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			if idx == 0 {
				filename := fmt.Sprintf("%d_%s", SystemID(name), fileHeader.Filename)
				dst, err := os.Create("uploads/" + filename)
				if err == nil {
					defer dst.Close()
					io.Copy(dst, file)
					file.Seek(0, 0)
					savedURL = "/uploads/" + filename
				}
			}

			embedding, err := getEmbedding(embeddingsURL, file, fileHeader.Filename)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error generating embedding for %s: %v", fileHeader.Filename, err), http.StatusInternalServerError)
				return
			}
			embeddings = append(embeddings, embedding)
		}

		err = saveMultipleToQdrant(qdrantURL, name, savedURL, color, material, embeddings)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving to Qdrant: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully uploaded and indexed %d views for %s", len(embeddings), name)
	}))

	uploadsFS := http.FileServer(http.Dir("uploads"))
	http.Handle("/uploads/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/uploads/", uploadsFS).ServeHTTP(w, r)
	}))

	http.HandleFunc("/analyze", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received /analyze request")
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(100 << 20) // Up to 100MB for multiple photos
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		imageFiles := r.MultipartForm.File["images"]
		if len(imageFiles) == 0 {
			// Fallback to single "image" field for backward compatibility
			if _, _, err := r.FormFile("image"); err == nil {
				imageFiles = r.MultipartForm.File["image"]
			}
		}

		if len(imageFiles) == 0 {
			http.Error(w, "Missing images", http.StatusBadRequest)
			return
		}

		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = r.FormValue("apiKey")
		}
		if apiKey == "" {
			log.Printf("Error: Missing API Key")
			http.Error(w, "Missing API Key", http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Printf("Error creating GenAI client: %v", err)
			http.Error(w, fmt.Sprintf("AI SDK error: %v", err), http.StatusInternalServerError)
			return
		}
		defer client.Close()

		model := client.GenerativeModel("gemma-4-26b-a4b-it")
		model.SetTemperature(0.1)
		model.SetMaxOutputTokens(1024)

		mcpUrl := os.Getenv("MCP_URL")
		if mcpUrl == "" {
			mcpUrl = "ws://mcp:8081/ws"
		}
		embeddingsURL := os.Getenv("EMBEDDINGS_URL")
		if embeddingsURL == "" {
			embeddingsURL = "http://localhost:8001"
		}
		qdrantURL := os.Getenv("QDRANT_URL")
		if qdrantURL == "" {
			qdrantURL = "http://localhost:6333"
		}

		productMaxScores := make(map[string]float32)
		productImageURLs := make(map[string]string)
		var allImageResults []ImageResult

		for imgIdx, fileHeader := range imageFiles {
			log.Printf("Processing image %d/%d: %s", imgIdx+1, len(imageFiles), fileHeader.Filename)
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error opening image %d: %v", imgIdx, err)
				continue
			}
			imgData, _ := io.ReadAll(file)
			file.Close()

			img, _, err := image.Decode(bytes.NewReader(imgData))
			if err != nil {
				log.Printf("Error decoding image %d: %v", imgIdx, err)
				continue
			}

			bounds := img.Bounds()
			width, height := bounds.Dx(), bounds.Dy()

			var geminiImgData []byte
			if width > 1024 || height > 1024 {
				resized := imaging.Fit(img, 1024, 1024, imaging.Lanczos)
				buf := new(bytes.Buffer)
				jpeg.Encode(buf, resized, &jpeg.Options{Quality: 80})
				geminiImgData = buf.Bytes()
			} else {
				buf := new(bytes.Buffer)
				jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
				geminiImgData = buf.Bytes()
			}

			prompt := `Analyze this jewelry display. List each distinct jewelry item.
Return ONLY a valid JSON array of objects. Each object must have:
"desc": a short description including color/material.
"box": an array [ymin, xmin, ymax, xmax] (normalized 0 to 1000).`

			var resp *genai.GenerateContentResponse
			for i := 0; i < 2; i++ {
				reqCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
				resp, err = model.GenerateContent(reqCtx, genai.Text(prompt), genai.ImageData("jpeg", geminiImgData))
				cancel()
				if err == nil {
					break
				}
			}
			if err != nil {
				log.Printf("AI error for image %d: %v", imgIdx, err)
				continue
			}

			var responseText string
			for _, part := range resp.Candidates[0].Content.Parts {
				if text, ok := part.(genai.Text); ok {
					responseText += string(text)
				}
			}

			var extracted string
			if start := strings.LastIndex(responseText, "```"); start != -1 {
				temp := responseText[:start]
				if blockStart := strings.LastIndex(temp, "```"); blockStart != -1 {
					inner := responseText[blockStart+3 : start]
					inner = strings.TrimPrefix(inner, "json")
					extracted = strings.TrimSpace(inner)
				}
			}
			if extracted == "" {
				lastOpen := strings.LastIndex(responseText, "[")
				lastClose := strings.LastIndex(responseText, "]")
				if lastOpen != -1 && lastClose != -1 && lastClose > lastOpen {
					candidate := responseText[lastOpen : lastClose+1]
					if strings.Contains(candidate, "\"desc\"") && (strings.Contains(candidate, "\"box\"") || strings.Contains(candidate, "\"box_2d\"")) {
						extracted = candidate
					}
				}
			}
			if extracted != "" {
				responseText = extracted
			}

			var detections []Detection
			json.Unmarshal([]byte(responseText), &detections)
			allImageResults = append(allImageResults, ImageResult{Detections: detections})

			mcpClient, err := NewMCPClient(mcpUrl)
			if err == nil {
				for _, det := range detections {
					box := det.Box
					if len(box) != 4 {
						box = det.Box2D
					}
					if len(box) != 4 {
						continue
					}

					ymin, xmin, ymax, xmax := box[0]*height/1000, box[1]*width/1000, box[2]*height/1000, box[3]*width/1000
					if xmin >= xmax || ymin >= ymax || xmin < 0 || ymin < 0 || xmax > width || ymax > height {
						continue
					}

					cropped := imaging.Crop(img, image.Rect(xmin, ymin, xmax, ymax))
					cropBuf := new(bytes.Buffer)
					jpeg.Encode(cropBuf, cropped, nil)

					emb, err := getEmbeddingBytes(embeddingsURL, cropBuf.Bytes(), "crop.jpg")
					if err != nil {
						continue
					}

					rawResults, err := mcpClient.CallVectorSearch(emb)
					if err != nil {
						continue
					}

					var hits []map[string]interface{}
					json.Unmarshal([]byte(rawResults), &hits)

					var bestMatchName string
					var bestMatchImgUrl string
					var bestMatchScore float32 = -1.0

					for _, hit := range hits {
						name, okName := hit["name"].(string)
						imgUrl, okImg := hit["imageUrl"].(string)
						if !okName || !okImg {
							continue
						}
						hitColor, _ := hit["color"].(string)
						hitMaterial, _ := hit["material"].(string)

						var score float32
						if s, ok := hit["score"].(float32); ok {
							score = s
						} else if s, ok := hit["score"].(float64); ok {
							score = float32(s)
						}

						descLower := strings.ToLower(det.Desc)
						var boost float32 = 0
						if hitColor != "" && strings.Contains(descLower, strings.ToLower(hitColor)) {
							boost += 0.20
						}
						if hitMaterial != "" && strings.Contains(descLower, strings.ToLower(hitMaterial)) {
							boost += 0.10
						}
						finalScore := score + boost

						if finalScore > bestMatchScore {
							bestMatchScore = finalScore
							bestMatchName = name
							bestMatchImgUrl = imgUrl
						}
					}

					if bestMatchScore > 0.70 {
						if currentMax, ok := productMaxScores[bestMatchName]; !ok || bestMatchScore > currentMax {
							productMaxScores[bestMatchName] = bestMatchScore
							productImageURLs[bestMatchName] = bestMatchImgUrl
							log.Printf("Match accepted from img %d: %s (%f)", imgIdx, bestMatchName, bestMatchScore)
						}
					}
				}
				mcpClient.Close()
			}
		}

		inventory, _ := listInventory(qdrantURL)
		foundResults := []ProductResult{}
		missingItems := []MissingItem{}

		for _, item := range inventory {
			payload, ok := item["payload"].(map[string]any)
			if !ok {
				continue
			}
			name, _ := payload["name"].(string)

			if score, found := productMaxScores[name]; found {
				foundResults = append(foundResults, ProductResult{
					Name:     name,
					ImageURL: fixURL(productImageURLs[name], r.Host),
					Score:    score,
				})
			} else {
				imgUrl, _ := payload["imageUrl"].(string)
				missingItems = append(missingItems, MissingItem{
					Name:     name,
					ImageURL: fixURL(imgUrl, r.Host),
				})
			}
		}

		log.Printf("Analysis complete across %d images, found %d products", len(imageFiles), len(foundResults))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AnalysisResponse{
			Found:        foundResults,
			Missing:      missingItems,
			ImageResults: allImageResults,
		})
	}))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func listInventory(url string) ([]map[string]any, error) {
	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	}
	if err != nil {
		return nil, err
	}
	defer client.Close()

	points, err := client.Scroll(context.Background(), &qdrant.ScrollPoints{
		CollectionName: "jewelry_inventory",
		Limit:          qdrant.PtrOf(uint32(100)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	var items []map[string]any
	for _, p := range points {
		payload := make(map[string]any)
		for k, v := range p.Payload {
			payload[k] = v.GetStringValue()
		}
		items = append(items, map[string]any{"id": fmt.Sprintf("%d", p.Id.GetNum()), "payload": payload})
	}
	return items, nil
}

func deleteFromQdrant(url string, idStr string) error {
	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	}
	if err != nil {
		return err
	}
	defer client.Close()
	var id uint64
	fmt.Sscanf(idStr, "%d", &id)
	_, err = client.Delete(context.Background(), &qdrant.DeletePoints{
		CollectionName: "jewelry_inventory",
		Points:         qdrant.NewPointsSelector(qdrant.NewIDNum(id)),
	})
	return err
}

func getEmbedding(url string, file io.Reader, filename string) ([]float32, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, file)
	writer.Close()
	req, _ := http.NewRequest("POST", url+"/embed", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var res EmbeddingResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Embedding, nil
}

func getEmbeddingBytes(url string, imgData []byte, filename string) ([]float32, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, bytes.NewReader(imgData))
	writer.Close()
	req, _ := http.NewRequest("POST", url+"/embed", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var res EmbeddingResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Embedding, nil
}

func saveMultipleToQdrant(url string, name string, imageUrl string, color string, material string, vectors [][]float32) error {
	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	}
	if err != nil {
		return err
	}
	defer client.Close()
	var points []*qdrant.PointStruct
	for i, vector := range vectors {
		id := uint64(SystemID(fmt.Sprintf("%s_%d", name, i)))
		payload := map[string]any{
			"name":     name,
			"color":    color,
			"material": material,
		}
		if imageUrl != "" {
			payload["imageUrl"] = imageUrl
		}
		points = append(points, &qdrant.PointStruct{Id: qdrant.NewIDNum(id), Vectors: qdrant.NewVectorsDense(vector), Payload: qdrant.NewValueMap(payload)})
	}
	_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{CollectionName: "jewelry_inventory", Points: points})
	return err
}

func performVectorSearch(url string, vector []float32) ([]map[string]any, error) {
	return performVectorSearchWithLimit(url, vector, 3)
}

func performVectorSearchWithLimit(url string, vector []float32, limit uint64) ([]map[string]any, error) {
	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	}
	if err != nil {
		return nil, err
	}
	defer client.Close()
	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "jewelry_inventory",
		Query:          qdrant.NewQueryDense(vector),
		Limit:          qdrant.PtrOf(limit),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}
	var results []map[string]any
	for _, hit := range searchResult {
		payload := make(map[string]any)
		for k, v := range hit.Payload {
			payload[k] = v.GetStringValue()
		}
		results = append(results, map[string]any{"name": payload["name"], "imageUrl": payload["imageUrl"], "score": hit.Score})
	}
	return results, nil
}

func SystemID(name string) int {
	h := 0
	for _, c := range name {
		h = 31*h + int(c)
	}
	if h < 0 {
		return -h
	}
	return h
}
