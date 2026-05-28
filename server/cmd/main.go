package main

import (
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strconv"

	"shelfscan-api/internal/embedding"
	"shelfscan-api/internal/gemini"
	handler "shelfscan-api/internal/handler"
	"shelfscan-api/internal/mcp"
	middleware "shelfscan-api/internal/middleware"
	qdrant "shelfscan-api/internal/qdrant"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	qdrantHost := os.Getenv("QDRANT_HOST")
	qdrantPort := os.Getenv("QDRANT_PORT")

	embeddingsHost := os.Getenv("EMBEDDINGS_HOST")
	if embeddingsHost == "" {
		embeddingsHost = os.Getenv("EMBEDDINGS_URL")
	}
	embeddingsPort := os.Getenv("EMBEDDINGS_PORT")

	prompt := os.Getenv("GEMINI_PROMPT")
	apiKey := os.Getenv("GEMINI_API_KEY")
	generativeModel := os.Getenv("GEMINI_MODEL")
	temperature := os.Getenv("GEMINI_TEMPERATURE")
	maxOutputTokens := os.Getenv("GEMINI_MAX_OUTPUT_TOKENS")

	if prompt == "" {
		prompt = `Analyze this jewelry display. List each distinct jewelry item.
	 Return ONLY a valid JSON array of objects. Each object must have:
	 "desc": a short description including color/material.
	 "box": an array [ymin, xmin, ymax, xmax] (normalized 0 to 1000).`
	}

	mcpUrl := os.Getenv("MCP_URL")

	corsMiddleware := middleware.NewMiddleware().
		WithAllowedOrigins("*").
		WithAllowedMethods("GET,POST,OPTIONS,DELETE,PUT").
		WithAllowedHeaders("Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization").
		DeclareCorsMiddleware()

	qdrantClient := qdrant.NewQdrantClient().
		WithHost(qdrantHost).
		WithPort(qdrantPort)

	embeddingClient := embedding.NewEmbedding().
		WithHost(embeddingsHost).
		WithPort(embeddingsPort)

	mcpClient := mcp.NewMCPClient().
		WithURL(mcpUrl)

	temp, err := strconv.ParseFloat(temperature, 64)
	if err != nil {
		temp = 0.1
	}
	tokens, err := strconv.Atoi(maxOutputTokens)
	if err != nil || tokens <= 0 {
		tokens = 2048
	}

	log.Printf("Starting server with GEMINI_MODEL=%s, GEMINI_MAX_OUTPUT_TOKENS=%d, GEMINI_TEMPERATURE=%f", generativeModel, tokens, temp)

	geminiClient := gemini.NewGeminiClient().
		WithGenerativeModel(generativeModel).
		WithApiKey(apiKey).
		WithPrompt(prompt).
		WithTemperature(temp).
		WithMaxOutputTokens(tokens)

	handlers := handler.NewHandler().
		WithCorsMiddleware(corsMiddleware).
		WithQdrantClient(qdrantClient).
		WithEmbeddingClient(embeddingClient).
		WithMCPClient(mcpClient).
		WithGeminiClient(geminiClient)

	http.HandleFunc("/health", handlers.HealthHandler)

	http.HandleFunc("/inventory", handlers.InventoryHandler)

	http.HandleFunc("/search", handlers.SearchHandler)

	http.HandleFunc("/upload", handlers.UploadHandler)

	http.HandleFunc("/uploads/", handlers.UploadsHandler)

	http.HandleFunc("/analyze", handlers.AnalyzeHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}

	// func

	// 	http.HandleFunc("/analyze", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
	// 		log.Printf("Received /analyze request")
	// 		if r.Method != http.MethodPost {
	// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 			return
	// 		}

	// 		err := r.ParseMultipartForm(100 << 20) // Up to 100MB for multiple photos
	// 		if err != nil {
	// 			log.Printf("Error parsing form: %v", err)
	// 			http.Error(w, err.Error(), http.StatusBadRequest)
	// 			return
	// 		}

	// 		imageFiles := r.MultipartForm.File["images"]
	// 		if len(imageFiles) == 0 {
	// 			// Fallback to single "image" field for backward compatibility
	// 			if _, _, err := r.FormFile("image"); err == nil {
	// 				imageFiles = r.MultipartForm.File["image"]
	// 			}
	// 		}

	// 		if len(imageFiles) == 0 {
	// 			http.Error(w, "Missing images", http.StatusBadRequest)
	// 			return
	// 		}

	// 		apiKey := os.Getenv("GEMINI_API_KEY")
	// 		if apiKey == "" {
	// 			apiKey = r.FormValue("apiKey")
	// 		}
	// 		if apiKey == "" {
	// 			log.Printf("Error: Missing API Key")
	// 			http.Error(w, "Missing API Key", http.StatusBadRequest)
	// 			return
	// 		}

	// 		ctx := context.Background()
	// 		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	// 		if err != nil {
	// 			log.Printf("Error creating GenAI client: %v", err)
	// 			http.Error(w, fmt.Sprintf("AI SDK error: %v", err), http.StatusInternalServerError)
	// 			return
	// 		}
	// 		defer client.Close()

	// 		model := client.GenerativeModel("gemma-4-26b-a4b-it")
	// 		model.SetTemperature(0.1)
	// 		model.SetMaxOutputTokens(1024)

	// 		mcpUrl := os.Getenv("MCP_URL")
	// 		if mcpUrl == "" {
	// 			mcpUrl = "ws://mcp:8081/ws"
	// 		}
	// 		embeddingsURL := os.Getenv("EMBEDDINGS_URL")
	// 		if embeddingsURL == "" {
	// 			embeddingsURL = "http://localhost:8001"
	// 		}
	// 		qdrantURL := os.Getenv("QDRANT_URL")
	// 		if qdrantURL == "" {
	// 			qdrantURL = "http://localhost:6333"
	// 		}

	// 		productMaxScores := make(map[string]float32)
	// 		productImageURLs := make(map[string]string)
	// 		var allImageResults []ImageResult

	// 		for imgIdx, fileHeader := range imageFiles {
	// 			log.Printf("Processing image %d/%d: %s", imgIdx+1, len(imageFiles), fileHeader.Filename)
	// 			file, err := fileHeader.Open()
	// 			if err != nil {
	// 				log.Printf("Error opening image %d: %v", imgIdx, err)
	// 				continue
	// 			}
	// 			imgData, _ := io.ReadAll(file)
	// 			file.Close()

	// 			img, _, err := image.Decode(bytes.NewReader(imgData))
	// 			if err != nil {
	// 				log.Printf("Error decoding image %d: %v", imgIdx, err)
	// 				continue
	// 			}

	// 			bounds := img.Bounds()
	// 			width, height := bounds.Dx(), bounds.Dy()

	// 			var geminiImgData []byte
	// 			if width > 1024 || height > 1024 {
	// 				resized := imaging.Fit(img, 1024, 1024, imaging.Lanczos)
	// 				buf := new(bytes.Buffer)
	// 				jpeg.Encode(buf, resized, &jpeg.Options{Quality: 80})
	// 				geminiImgData = buf.Bytes()
	// 			} else {
	// 				buf := new(bytes.Buffer)
	// 				jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
	// 				geminiImgData = buf.Bytes()
	// 			}

	// 			prompt := `Analyze this jewelry display. List each distinct jewelry item.
	// Return ONLY a valid JSON array of objects. Each object must have:
	// "desc": a short description including color/material.
	// "box": an array [ymin, xmin, ymax, xmax] (normalized 0 to 1000).`

	// 			var resp *genai.GenerateContentResponse
	// 			for i := 0; i < 2; i++ {
	// 				reqCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	// 				resp, err = model.GenerateContent(reqCtx, genai.Text(prompt), genai.ImageData("jpeg", geminiImgData))
	// 				cancel()
	// 				if err == nil {
	// 					break
	// 				}
	// 			}
	// 			if err != nil {
	// 				log.Printf("AI error for image %d: %v", imgIdx, err)
	// 				continue
	// 			}

	// 			var responseText string
	// 			for _, part := range resp.Candidates[0].Content.Parts {
	// 				if text, ok := part.(genai.Text); ok {
	// 					responseText += string(text)
	// 				}
	// 			}

	// 			var extracted string
	// 			if start := strings.LastIndex(responseText, "```"); start != -1 {
	// 				temp := responseText[:start]
	// 				if blockStart := strings.LastIndex(temp, "```"); blockStart != -1 {
	// 					inner := responseText[blockStart+3 : start]
	// 					inner = strings.TrimPrefix(inner, "json")
	// 					extracted = strings.TrimSpace(inner)
	// 				}
	// 			}
	// 			if extracted == "" {
	// 				lastOpen := strings.LastIndex(responseText, "[")
	// 				lastClose := strings.LastIndex(responseText, "]")
	// 				if lastOpen != -1 && lastClose != -1 && lastClose > lastOpen {
	// 					candidate := responseText[lastOpen : lastClose+1]
	// 					if strings.Contains(candidate, "\"desc\"") && (strings.Contains(candidate, "\"box\"") || strings.Contains(candidate, "\"box_2d\"")) {
	// 						extracted = candidate
	// 					}
	// 				}
	// 			}
	// 			if extracted != "" {
	// 				responseText = extracted
	// 			}

	// 			var detections []Detection
	// 			json.Unmarshal([]byte(responseText), &detections)
	// 			allImageResults = append(allImageResults, ImageResult{Detections: detections})

	// 			mcpClient, err := NewMCPClient(mcpUrl)
	// 			if err == nil {
	// 				for _, det := range detections {
	// 					box := det.Box
	// 					if len(box) != 4 {
	// 						box = det.Box2D
	// 					}
	// 					if len(box) != 4 {
	// 						continue
	// 					}

	// 					ymin, xmin, ymax, xmax := box[0]*height/1000, box[1]*width/1000, box[2]*height/1000, box[3]*width/1000
	// 					if xmin >= xmax || ymin >= ymax || xmin < 0 || ymin < 0 || xmax > width || ymax > height {
	// 						continue
	// 					}

	// 					cropped := imaging.Crop(img, image.Rect(xmin, ymin, xmax, ymax))
	// 					cropBuf := new(bytes.Buffer)
	// 					jpeg.Encode(cropBuf, cropped, nil)

	// 					emb, err := getEmbeddingBytes(embeddingsURL, cropBuf.Bytes(), "crop.jpg")
	// 					if err != nil {
	// 						continue
	// 					}

	// 					rawResults, err := mcpClient.CallVectorSearch(emb)
	// 					if err != nil {
	// 						continue
	// 					}

	// 					var hits []map[string]interface{}
	// 					json.Unmarshal([]byte(rawResults), &hits)

	// 					var bestMatchName string
	// 					var bestMatchImgUrl string
	// 					var bestMatchScore float32 = -1.0

	// 					for _, hit := range hits {
	// 						name, okName := hit["name"].(string)
	// 						imgUrl, okImg := hit["imageUrl"].(string)
	// 						if !okName || !okImg {
	// 							continue
	// 						}
	// 						hitColor, _ := hit["color"].(string)
	// 						hitMaterial, _ := hit["material"].(string)

	// 						var score float32
	// 						if s, ok := hit["score"].(float32); ok {
	// 							score = s
	// 						} else if s, ok := hit["score"].(float64); ok {
	// 							score = float32(s)
	// 						}

	// 						descLower := strings.ToLower(det.Desc)
	// 						var boost float32 = 0
	// 						if hitColor != "" && strings.Contains(descLower, strings.ToLower(hitColor)) {
	// 							boost += 0.20
	// 						}
	// 						if hitMaterial != "" && strings.Contains(descLower, strings.ToLower(hitMaterial)) {
	// 							boost += 0.10
	// 						}
	// 						finalScore := score + boost

	// 						if finalScore > bestMatchScore {
	// 							bestMatchScore = finalScore
	// 							bestMatchName = name
	// 							bestMatchImgUrl = imgUrl
	// 						}
	// 					}

	// 					if bestMatchScore > 0.70 {
	// 						if currentMax, ok := productMaxScores[bestMatchName]; !ok || bestMatchScore > currentMax {
	// 							productMaxScores[bestMatchName] = bestMatchScore
	// 							productImageURLs[bestMatchName] = bestMatchImgUrl
	// 							log.Printf("Match accepted from img %d: %s (%f)", imgIdx, bestMatchName, bestMatchScore)
	// 						}
	// 					}
	// 				}
	// 				mcpClient.Close()
	// 			}
	// 		}

	// 		inventory, _ := listInventory(qdrantURL)
	// 		foundResults := []ProductResult{}
	// 		missingItems := []MissingItem{}

	// 		for _, item := range inventory {
	// 			payload, ok := item["payload"].(map[string]any)
	// 			if !ok {
	// 				continue
	// 			}
	// 			name, _ := payload["name"].(string)

	// 			if score, found := productMaxScores[name]; found {
	// 				foundResults = append(foundResults, ProductResult{
	// 					Name:     name,
	// 					ImageURL: fixURL(productImageURLs[name], r.Host),
	// 					Score:    score,
	// 				})
	// 			} else {
	// 				imgUrl, _ := payload["imageUrl"].(string)
	// 				missingItems = append(missingItems, MissingItem{
	// 					Name:     name,
	// 					ImageURL: fixURL(imgUrl, r.Host),
	// 				})
	// 			}
	// 		}

	// 		log.Printf("Analysis complete across %d images, found %d products", len(imageFiles), len(foundResults))
	// 		w.Header().Set("Content-Type", "application/json")
	// 		json.NewEncoder(w).Encode(AnalysisResponse{
	// 			Found:        foundResults,
	// 			Missing:      missingItems,
	// 			ImageResults: allImageResults,
	// 		})
	// 	}))

	// 	if err := http.ListenAndServe(":"+port, nil); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// func listInventory(url string) ([]map[string]any, error) {
	// 	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	// 	if err != nil {
	// 		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	// 	}
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer client.Close()

	// 	points, err := client.Scroll(context.Background(), &qdrant.ScrollPoints{
	// 		CollectionName: "jewelry_inventory",
	// 		Limit:          qdrant.PtrOf(uint32(100)),
	// 		WithPayload:    qdrant.NewWithPayload(true),
	// 	})
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	var items []map[string]any
	// 	for _, p := range points {
	// 		payload := make(map[string]any)
	// 		for k, v := range p.Payload {
	// 			payload[k] = v.GetStringValue()
	// 		}
	// 		items = append(items, map[string]any{"id": fmt.Sprintf("%d", p.Id.GetNum()), "payload": payload})
	// 	}
	// 	return items, nil
	// }

	// func deleteFromQdrant(url string, idStr string) error {
	// 	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	// 	if err != nil {
	// 		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	// 	}
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer client.Close()
	// 	var id uint64
	// 	fmt.Sscanf(idStr, "%d", &id)
	// 	_, err = client.Delete(context.Background(), &qdrant.DeletePoints{
	// 		CollectionName: "jewelry_inventory",
	// 		Points:         qdrant.NewPointsSelector(qdrant.NewIDNum(id)),
	// 	})
	// 	return err
	// }

	// func getEmbedding(url string, file io.Reader, filename string) ([]float32, error) {
	// 	body := &bytes.Buffer{}
	// 	writer := multipart.NewWriter(body)
	// 	part, _ := writer.CreateFormFile("file", filename)
	// 	io.Copy(part, file)
	// 	writer.Close()
	// 	req, _ := http.NewRequest("POST", url+"/embed", body)
	// 	req.Header.Set("Content-Type", writer.FormDataContentType())
	// 	resp, err := http.DefaultClient.Do(req)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer resp.Body.Close()
	// 	var res EmbeddingResponse
	// 	json.NewDecoder(resp.Body).Decode(&res)
	// 	return res.Embedding, nil
	// }

	// func getEmbeddingBytes(url string, imgData []byte, filename string) ([]float32, error) {
	// 	body := &bytes.Buffer{}
	// 	writer := multipart.NewWriter(body)
	// 	part, _ := writer.CreateFormFile("file", filename)
	// 	io.Copy(part, bytes.NewReader(imgData))
	// 	writer.Close()
	// 	req, _ := http.NewRequest("POST", url+"/embed", body)
	// 	req.Header.Set("Content-Type", writer.FormDataContentType())
	// 	resp, err := http.DefaultClient.Do(req)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer resp.Body.Close()
	// 	var res EmbeddingResponse
	// 	json.NewDecoder(resp.Body).Decode(&res)
	// 	return res.Embedding, nil
	// }

	// func saveMultipleToQdrant(url string, name string, imageUrl string, color string, material string, vectors [][]float32) error {
	// 	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	// 	if err != nil {
	// 		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	// 	}
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer client.Close()
	// 	var points []*qdrant.PointStruct
	// 	for i, vector := range vectors {
	// 		id := uint64(SystemID(fmt.Sprintf("%s_%d", name, i)))
	// 		payload := map[string]any{
	// 			"name":     name,
	// 			"color":    color,
	// 			"material": material,
	// 		}
	// 		if imageUrl != "" {
	// 			payload["imageUrl"] = imageUrl
	// 		}
	// 		points = append(points, &qdrant.PointStruct{Id: qdrant.NewIDNum(id), Vectors: qdrant.NewVectorsDense(vector), Payload: qdrant.NewValueMap(payload)})
	// 	}
	// 	_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{CollectionName: "jewelry_inventory", Points: points})
	// 	return err
	// }

	// func performVectorSearch(url string, vector []float32) ([]map[string]any, error) {
	// 	return performVectorSearchWithLimit(url, vector, 3)
	// }

	// func performVectorSearchWithLimit(url string, vector []float32, limit uint64) ([]map[string]any, error) {
	// 	client, err := qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
	// 	if err != nil {
	// 		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	// 	}
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer client.Close()
	// 	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
	// 		CollectionName: "jewelry_inventory",
	// 		Query:          qdrant.NewQueryDense(vector),
	// 		Limit:          qdrant.PtrOf(limit),
	// 		WithPayload:    qdrant.NewWithPayload(true),
	// 	})
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	var results []map[string]any
	// 	for _, hit := range searchResult {
	// 		payload := make(map[string]any)
	// 		for k, v := range hit.Payload {
	// 			payload[k] = v.GetStringValue()
	// 		}
	// 		results = append(results, map[string]any{"name": payload["name"], "imageUrl": payload["imageUrl"], "score": hit.Score})
	// 	}
	// 	return results, nil
	// }

	//	func SystemID(name string) int {
	//		h := 0
	//		for _, c := range name {
	//			h = 31*h + int(c)
	//		}
	//		if h < 0 {
	//			return -h
	//		}
	//		return h
	//	}
}
