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

	"github.com/disintegration/imaging"
	"github.com/google/generative-ai-go/genai"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/api/option"
)

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
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
		// If it's already absolute (legacy), try to fix the host if it's localhost
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
			// Fix URLs dynamically for the current host
			for _, item := range items {
				payload := item["payload"].(map[string]any)
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

		// 1. Get Embedding
		embedding, err := getEmbedding(embeddingsURL, file, handler.Filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Embedding error: %v", err), http.StatusInternalServerError)
			return
		}

		// 2. Search Qdrant
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

		files := r.MultipartForm.File["images"]
		if len(files) == 0 {
			http.Error(w, "No images uploaded", http.StatusBadRequest)
			return
		}

		// Create uploads directory if it doesn't exist
		os.MkdirAll("uploads", 0755)
		var savedURL string

		var embeddings [][]float32
		for i, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Error opening file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Save the first image as the representative thumbnail
			if i == 0 {
				filename := fmt.Sprintf("%d_%s", SystemID(name), fileHeader.Filename)
				dst, err := os.Create("uploads/" + filename)
				if err == nil {
					defer dst.Close()
					io.Copy(dst, file)
					file.Seek(0, 0) // Reset file pointer for embedding
					// Store only the relative path to be flexible with hostnames
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

		err = saveMultipleToQdrant(qdrantURL, name, savedURL, embeddings)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving to Qdrant: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully uploaded and indexed %d views for %s", len(embeddings), name)
	}))

	// Serve static uploads with CORS
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

		err := r.ParseMultipartForm(50 << 20)
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("image")
		if err != nil {
			log.Printf("Error getting image file: %v", err)
			http.Error(w, "Missing image", http.StatusBadRequest)
			return
		}
		defer file.Close()

		imgData, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Error reading image: %v", err)
			http.Error(w, "Error reading image", http.StatusInternalServerError)
			return
		}

		log.Printf("Image received, size: %d bytes", len(imgData))

		img, _, err := image.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Printf("Error decoding image: %v", err)
			http.Error(w, "Error decoding image", http.StatusInternalServerError)
			return
		}
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		log.Printf("Image decoded successfully, dimensions: %dx%d", width, height)

		// Re-encode to JPEG to ensure format consistency
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
			log.Printf("Error re-encoding image: %v", err)
			http.Error(w, "Error processing image", http.StatusInternalServerError)
			return
		}
		imgData = buf.Bytes()

		ctx := context.Background()
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = r.FormValue("apiKey")
		}
		if apiKey == "" {
			log.Printf("Error: Missing API Key")
			http.Error(w, "Missing API Key", http.StatusBadRequest)
			return
		}

		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Printf("Error creating GenAI client: %v", err)
			http.Error(w, fmt.Sprintf("AI SDK error: %v", err), http.StatusInternalServerError)
			return
		}
		defer client.Close()

		// Using gemma-4-26b-it as requested (native multimodal support)
		model := client.GenerativeModel("gemma-4-26b-a4b-it")

		prompt := `Analyze this jewelry display. Identify each distinct jewelry item. Return ONLY a valid JSON array of objects. Each object must have:
"desc": a short description of the item.
"box": an array of 4 numbers [ymin, xmin, ymax, xmax] representing the normalized bounding box (0 to 1000).`

		var resp *genai.GenerateContentResponse
		log.Printf("Sending request to Gemini/Gemma model...")
		
		// Simple retry for transient 500 errors
		for i := 0; i < 3; i++ {
			resp, err = model.GenerateContent(ctx, genai.Text(prompt), genai.ImageData("jpeg", imgData))
			if err == nil {
				break
			}
			if !strings.Contains(err.Error(), "500") {
				break
			}
			log.Printf("AI Generation transient error (attempt %d): %v", i+1, err)
		}

		if err != nil {
			log.Printf("AI Generation error: %v", err)
			http.Error(w, fmt.Sprintf("AI Generation error: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Received response from Gemini/Gemma")

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			log.Printf("Empty response from AI")
			http.Error(w, "Empty response from AI", http.StatusInternalServerError)
			return
		}

		var responseText string
		for _, part := range resp.Candidates[0].Content.Parts {
			if text, ok := part.(genai.Text); ok {
				responseText += string(text)
			}
		}

		// Improved extraction: find the JSON array block more carefully
		var extracted string
		// First try: look for the last markdown block
		if start := strings.LastIndex(responseText, "```"); start != -1 {
			temp := responseText[:start]
			if blockStart := strings.LastIndex(temp, "```"); blockStart != -1 {
				inner := responseText[blockStart+3 : start]
				inner = strings.TrimPrefix(inner, "json")
				extracted = strings.TrimSpace(inner)
			}
		}
		
		// Second try: look for the last [ ] block that looks like a JSON array of objects
		if extracted == "" {
			// Find all [ ... ] candidates
			lastOpen := strings.LastIndex(responseText, "[")
			lastClose := strings.LastIndex(responseText, "]")
			if lastOpen != -1 && lastClose != -1 && lastClose > lastOpen {
				candidate := responseText[lastOpen : lastClose+1]
				// Check if it contains "desc" and some form of box to be reasonably sure
				if strings.Contains(candidate, "\"desc\"") && (strings.Contains(candidate, "\"box\"") || strings.Contains(candidate, "\"box_2d\"")) {
					extracted = candidate
				}
			}
		}

		if extracted != "" {
			responseText = extracted
		}
		log.Printf("Extracted JSON: %s", responseText)

		type Detection struct {
			Desc  string `json:"desc"`
			Box   []int  `json:"box"`
			Box2D []int  `json:"box_2d"`
		}
		var detections []Detection
		if err := json.Unmarshal([]byte(responseText), &detections); err != nil {
			log.Printf("Failed to parse AI JSON: %v", err)
			http.Error(w, fmt.Sprintf("Failed to parse AI JSON: %v. Raw: %s", err, responseText), http.StatusInternalServerError)
			return
		}

		type ProductResult struct {
			Name     string  `json:"name"`
			ImageURL string  `json:"imageUrl"`
			Score    float32 `json:"score"`
		}
		
		foundMap := make(map[string]ProductResult)

		for i, det := range detections {
			box := det.Box
			if len(box) != 4 {
				box = det.Box2D
			}
			if len(box) != 4 {
				continue
			}

			ymin := box[0] * height / 1000
			xmin := box[1] * width / 1000
			ymax := box[2] * height / 1000
			xmax := box[3] * width / 1000

			if xmin >= xmax || ymin >= ymax || xmin < 0 || ymin < 0 || xmax > width || ymax > height {
				continue
			}

			log.Printf("Cropping image %d: [%d, %d, %d, %d]", i, xmin, ymin, xmax, ymax)
			cropped := imaging.Crop(img, image.Rect(xmin, ymin, xmax, ymax))
			cropBuf := new(bytes.Buffer)
			jpeg.Encode(cropBuf, cropped, nil)

			emb, err := getEmbeddingBytes(embeddingsURL, cropBuf.Bytes(), "crop.jpg")
			if err != nil {
				log.Printf("Error getting embedding for crop %d: %v", i, err)
				continue
			}

			hits, _ := performVectorSearch(qdrantURL, emb)
			if len(hits) > 0 {
				name := hits[0]["name"].(string)
				imgUrl, _ := hits[0]["imageUrl"].(string)
				imgUrl = fixURL(imgUrl, r.Host)
				var score float32
				if s, ok := hits[0]["score"].(float32); ok {
					score = s
				} else if s, ok := hits[0]["score"].(float64); ok {
					score = float32(s)
				}

				log.Printf("Crop %d matched: %s (score: %f)", i, name, score)
				// Use 0.35 as threshold. If multiple detections match the same product, 
				// we keep the one with the highest score.
				if score > 0.35 {
					if existing, ok := foundMap[name]; !ok || score > existing.Score {
						foundMap[name] = ProductResult{
							Name:     name,
							ImageURL: imgUrl,
							Score:    score,
						}
					}
				}
			}
		}

		// Get all inventory to find missing items
		inventory, _ := listInventory(qdrantURL)
		
		var foundResults []ProductResult
		for _, v := range foundMap {
			foundResults = append(foundResults, v)
		}

		type MissingItem struct {
			Name     string `json:"name"`
			ImageURL string `json:"imageUrl"`
		}
		var missingItems []MissingItem
		
		for _, item := range inventory {
			payload := item["payload"].(map[string]any)
			name := payload["name"].(string)
			imgUrl, _ := payload["imageUrl"].(string)
			imgUrl = fixURL(imgUrl, r.Host)
			
			if _, found := foundMap[name]; !found {
				missingItems = append(missingItems, MissingItem{
					Name:     name,
					ImageURL: imgUrl,
				})
			}
		}

		type AnalysisResponse struct {
			Found   []ProductResult `json:"found"`
			Missing []MissingItem   `json:"missing"`
		}

		log.Printf("Analysis complete, returning %d unique found and %d missing", len(foundResults), len(missingItems))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AnalysisResponse{
			Found:   foundResults,
			Missing: missingItems,
		})
	}))

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func listInventory(url string) ([]map[string]any, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "qdrant",
		Port: 6334,
	})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: "localhost",
			Port: 6334,
		})
		if err != nil {
			return nil, err
		}
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

		item := map[string]any{
			"id":      fmt.Sprintf("%d", p.Id.GetNum()),
			"payload": payload,
		}
		items = append(items, item)
	}

	return items, nil
}

func deleteFromQdrant(url string, idStr string) error {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "qdrant",
		Port: 6334,
	})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: "localhost",
			Port: 6334,
		})
		if err != nil {
			return err
		}
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
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", url+"/embed", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var res EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Embedding, nil
}

func getEmbeddingBytes(url string, imgData []byte, filename string) ([]float32, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, bytes.NewReader(imgData))
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", url+"/embed", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var res EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Embedding, nil
}

func saveMultipleToQdrant(url string, name string, imageUrl string, vectors [][]float32) error {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "qdrant",
		Port: 6334,
	})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: "localhost",
			Port: 6334,
		})
		if err != nil {
			return err
		}
	}
	defer client.Close()

	var points []*qdrant.PointStruct
	for i, vector := range vectors {
		id := uint64(SystemID(fmt.Sprintf("%s_%d", name, i)))
		payload := map[string]any{
			"name": name,
		}
		if imageUrl != "" {
			payload["imageUrl"] = imageUrl
		}

		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(id),
			Vectors: qdrant.NewVectorsDense(vector),
			Payload: qdrant.NewValueMap(payload),
		})
	}

	upsertPoints := &qdrant.UpsertPoints{
		CollectionName: "jewelry_inventory",
		Points:         points,
	}

	_, err = client.Upsert(context.Background(), upsertPoints)
	return err
}

func performVectorSearch(url string, vector []float32) ([]map[string]any, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "qdrant",
		Port: 6334,
	})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: "localhost",
			Port: 6334,
		})
		if err != nil {
			return nil, err
		}
	}
	defer client.Close()

	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "jewelry_inventory",
		Query:          qdrant.NewQueryDense(vector),
		Limit:          qdrant.PtrOf(uint64(3)),
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
		results = append(results, map[string]any{
			"name":     payload["name"],
			"imageUrl": payload["imageUrl"],
			"score":    hit.Score,
		})
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
