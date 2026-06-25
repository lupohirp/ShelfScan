package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"shelfscan-api/internal/domain"
	subdomain "shelfscan-api/internal/domain/sub"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/generative-ai-go/genai"
)

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "healthy")
	})
	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) InventoryHandler(w http.ResponseWriter, r *http.Request) {

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			items, err := h.qdrantClient.ListInventory()
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
				if thumb, ok := payload["thumbUrl"].(string); ok {
					payload["thumbUrl"] = fixURL(thumb, r.Host)
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
			err := h.qdrantClient.DeleteFromQdrant(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPut {
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "Missing id", http.StatusBadRequest)
				return
			}
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				r.ParseForm()
			}
			name := r.FormValue("name")
			if name == "" {
				http.Error(w, "Missing jewelry name", http.StatusBadRequest)
				return
			}
			sku := r.FormValue("sku")
			if sku == "" {
				http.Error(w, "Missing jewelry SKU", http.StatusBadRequest)
				return
			}
			color := r.FormValue("color")
			material := r.FormValue("material")

			err = h.qdrantClient.UpdatePayload(id, name, sku, color, material)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		embedding, err := h.embeddingClient.GetEmbedding(file, handler.Filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Embedding error: %v", err), http.StatusInternalServerError)
			return
		}

		results, err := h.qdrantClient.PerformVectorSearch(embedding)
		if err != nil {
			http.Error(w, fmt.Sprintf("Search error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)

}

func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		sku := r.FormValue("sku")
		if sku == "" {
			http.Error(w, "Missing jewelry SKU", http.StatusBadRequest)
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
		var savedURLs []string
		var savedThumbURLs []string

		var embeddings [][]float32
		var aiDesc string
		for i, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Error opening file", http.StatusInternalServerError)
				return
			}

			if i == 0 {
				imgBytes, readErr := io.ReadAll(file)
				if readErr == nil {
					prompt := "Descrivi brevemente l'articolo di gioielleria o orologio in questa immagine specificando esplicitamente la categoria (es. orologio, collana, anello, bracciale, orecchini), lo stile e i colori rilevanti in italiano."
					desc, err := h.geminiClient.DescribeImageWithModel(r.Context(), "gemma-4-26b-a4b-it", prompt, imgBytes)
					if err != nil {
						log.Printf("Warning: failed to generate AI description with gemma4: %v", err)
					} else {
						aiDesc = strings.TrimSpace(desc)
						log.Printf("Generated AI Description with gemma4: %s", aiDesc)
					}
				} else {
					log.Printf("Warning: failed to read first image bytes for AI description: %v", readErr)
				}
				// Reset reader so we can continue processing/saving
				file.Seek(0, 0)
			}

			filename := fmt.Sprintf("%d_%s", systemID(name), fileHeader.Filename)
			thumbFilename := "thumb_" + filename

			// 1. Save original file
			origDst, err := os.Create("uploads/" + filename)
			if err == nil {
				io.Copy(origDst, file)
				origDst.Close()
			}
			file.Seek(0, 0)

			// 2. Decode and save thumbnail
			img, _, decodeErr := image.Decode(file)
			if decodeErr == nil {
				thumb := imaging.Fit(img, 400, 400, imaging.Linear)
				dst, err := os.Create("uploads/" + thumbFilename)
				if err == nil {
					jpeg.Encode(dst, thumb, &jpeg.Options{Quality: 80})
					dst.Close()
				}
			} else {
				// Fallback to copy if decode failed
				file.Seek(0, 0)
				dst, err := os.Create("uploads/" + thumbFilename)
				if err == nil {
					io.Copy(dst, file)
					dst.Close()
				}
			}
			file.Seek(0, 0)
			savedURLs = append(savedURLs, "/uploads/"+filename)
			savedThumbURLs = append(savedThumbURLs, "/uploads/"+thumbFilename)

			embedding, err := h.embeddingClient.GetEmbedding(file, fileHeader.Filename)
			file.Close()
			if err != nil {
				http.Error(w, fmt.Sprintf("Error generating embedding for %s: %v", fileHeader.Filename, err), http.StatusInternalServerError)
				return
			}
			embeddings = append(embeddings, embedding)
		}

		err = h.qdrantClient.SaveMultipleToQdrant(name, sku, savedURLs, savedThumbURLs, color, material, aiDesc, embeddings)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving to Qdrant: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully uploaded and indexed %d views for %s", len(embeddings), name)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)

}

func (h *Handler) UploadsHandler(w http.ResponseWriter, r *http.Request) {

	uploadsFS := http.FileServer(http.Dir("uploads"))

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/uploads/", uploadsFS).ServeHTTP(w, r)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)

}

type CandidateLog struct {
	Name          string  `json:"name"`
	Sku           string  `json:"sku"`
	RawScore      float32 `json:"raw_score"`
	CategoryMatch bool    `json:"category_match"`
	ColorConflict bool    `json:"color_conflict"`
	Boost         float32 `json:"boost"`
	FinalScore    float32 `json:"final_score"`
}

type DetectionLog struct {
	Desc           string         `json:"desc"`
	Box            []int          `json:"box"`
	CropFilename   string         `json:"crop_filename,omitempty"`
	BestMatch      string         `json:"best_match,omitempty"`
	BestMatchScore float32        `json:"best_match_score,omitempty"`
	Accepted       bool           `json:"accepted"`
	Candidates     []CandidateLog `json:"candidates"`
}

type ImageLog struct {
	ImageIdx          int            `json:"image_idx"`
	Filename          string         `json:"filename"`
	SavedOriginalFile string         `json:"saved_original_file"`
	GeminiModel       string         `json:"gemini_model"`
	GeminiRawResponse string         `json:"gemini_raw_response"`
	Detections        []DetectionLog `json:"detections"`
}

type RequestLog struct {
	RequestID string     `json:"request_id"`
	Timestamp string     `json:"timestamp"`
	Images    []ImageLog `json:"images"`
	Found     []any      `json:"found"`
	Missing   []any      `json:"missing"`
}

func (h *Handler) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Create a unique request directory under /app/logs
		reqID := fmt.Sprintf("%s_%03d", time.Now().Format("20060102_150405"), time.Now().Nanosecond()/1000000%1000)
		reqDir := filepath.Join("/app/logs", "req_"+reqID)
		if mkdirErr := os.MkdirAll(reqDir, 0755); mkdirErr != nil {
			log.Printf("Warning: failed to create request log dir %s: %v", reqDir, mkdirErr)
		}

		type AcceptedMatch struct {
			Name     string
			Sku      string
			ImageURL string
			Score    float32
		}
		var acceptedMatches []AcceptedMatch
		allImageResults := make([]subdomain.ImageResult, len(imageFiles))

		imageLogs := make([]ImageLog, len(imageFiles))
		var logMu sync.Mutex

		var mainMu sync.Mutex
		var outerWg sync.WaitGroup
		sem := make(chan struct{}, 2)

		for imgIdx, fileHeader := range imageFiles {
			outerWg.Add(1)
			go func(imgIdx int, fileHeader *multipart.FileHeader) {
				defer outerWg.Done()

				log.Printf("Processing image %d/%d: %s", imgIdx+1, len(imageFiles), fileHeader.Filename)
				file, err := fileHeader.Open()
				if err != nil {
					log.Printf("Error opening image %d: %v", imgIdx, err)
					return
				}
				imgData, _ := io.ReadAll(file)
				file.Close()

				// Save original image to log directory
				originalFilename := fmt.Sprintf("original_%d.jpg", imgIdx)
				_ = os.WriteFile(filepath.Join(reqDir, originalFilename), imgData, 0644)

				logMu.Lock()
				imageLogs[imgIdx] = ImageLog{
					ImageIdx:          imgIdx,
					Filename:          fileHeader.Filename,
					SavedOriginalFile: originalFilename,
				}
				logMu.Unlock()

				img, _, err := image.Decode(bytes.NewReader(imgData))
				if err != nil {
					log.Printf("Error decoding image %d: %v", imgIdx, err)
					return
				}

				bounds := img.Bounds()
				width, height := bounds.Dx(), bounds.Dy()

				var geminiImgData []byte
				if width > 1200 || height > 1200 {
					resized := imaging.Fit(img, 1200, 1200, imaging.Linear)
					buf := new(bytes.Buffer)
					jpeg.Encode(buf, resized, &jpeg.Options{Quality: 80})
					geminiImgData = buf.Bytes()
				} else {
					buf := new(bytes.Buffer)
					jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
					geminiImgData = buf.Bytes()
				}

				ctx := context.Background()

				resp, modelUsed, err := h.geminiClient.GenerateContentWithFallback(ctx, h.geminiClient.Prompt, geminiImgData)
				if err != nil {
					log.Printf("AI error for image %d (tried all models): %v", imgIdx, err)
					return
				}
				log.Printf("Gemini Finish Reason for image %d (model: %s): %v", imgIdx, modelUsed, resp.Candidates[0].FinishReason)

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

				log.Printf("Gemini Response: %s", responseText)

				detections := parseDetections(responseText)
				allImageResults[imgIdx] = subdomain.ImageResult{Detections: detections}

				logMu.Lock()
				imageLogs[imgIdx].GeminiModel = modelUsed
				imageLogs[imgIdx].GeminiRawResponse = responseText
				imageLogs[imgIdx].Detections = make([]DetectionLog, len(detections))
				logMu.Unlock()

				var wg sync.WaitGroup
				for detIdx, det := range detections {
					wg.Add(1)
					go func(detIdx int, det subdomain.Detection) {
						defer wg.Done()

						box := det.Box
						if len(box) != 4 {
							box = det.Box2D
						}
						if len(box) != 4 {
							return
						}

						logMu.Lock()
						imageLogs[imgIdx].Detections[detIdx] = DetectionLog{
							Desc: det.Desc,
							Box:  box,
						}
						logMu.Unlock()

						ymin, xmin, ymax, xmax := box[0]*height/1000, box[1]*width/1000, box[2]*height/1000, box[3]*width/1000
						if xmin >= xmax || ymin >= ymax || xmin < 0 || ymin < 0 || xmax > width || ymax > height {
							return
						}

						cropped := imaging.Crop(img, image.Rect(xmin, ymin, xmax, ymax))
						cropBuf := new(bytes.Buffer)
						jpeg.Encode(cropBuf, cropped, nil)

						// Save cropped image to log directory
						cropFilename := fmt.Sprintf("crop_%d_%d.jpg", imgIdx, detIdx)
						_ = os.WriteFile(filepath.Join(reqDir, cropFilename), cropBuf.Bytes(), 0644)

						logMu.Lock()
						imageLogs[imgIdx].Detections[detIdx].CropFilename = cropFilename
						logMu.Unlock()

						sem <- struct{}{}
						emb, err := h.embeddingClient.GetEmbeddingBytes(cropBuf.Bytes(), "crop.jpg")
						<-sem
						if err != nil {
							return
						}

						hits, err := h.qdrantClient.PerformVectorSearch(emb)
						if err != nil {
							return
						}

						var bestMatchName string
						var bestMatchImgUrl string
						var bestMatchSku string
						var bestMatchScore float32 = -1.0

						for _, hit := range hits {
							name, okName := hit["name"].(string)
							imgUrl, okImg := hit["imageUrl"].(string)
							if !okName || !okImg {
								continue
							}

							aiDesc, _ := hit["ai_description"].(string)

							if !isCategoryMatch(det.Desc, name, aiDesc) {
								continue
							}

							hitColor, _ := hit["color"].(string)
							hitMaterial, _ := hit["material"].(string)
							sku, _ := hit["sku"].(string)

							var score float32
							if s, ok := hit["score"].(float32); ok {
								score = s
							} else if s, ok := hit["score"].(float64); ok {
								score = float32(s)
							}

							catMatch := isCategoryMatch(det.Desc, name, aiDesc)
							colConflict := hasColorConflict(det.Desc, name, hitColor)

							descLower := strings.ToLower(det.Desc)
							var boost float32 = 0

							if hitColor != "" {
								parts := strings.Split(hitColor, ",")
								for i, part := range parts {
									part = strings.TrimSpace(strings.ToLower(part))
									if part != "" {
										weight := float32(0.10)
										if i == 0 {
											weight = float32(0.20)
										}
										if findSynonymIndex(descLower, part) != -1 {
											boost += weight
										} else {
											words := strings.Fields(part)
											for _, w := range words {
												w = strings.TrimFunc(w, func(r rune) bool {
													return r == ':' || r == '(' || r == ')' || r == '[' || r == ']' || r == ',' || r == '/'
												})
												if len(w) > 2 && findSynonymIndex(descLower, w) != -1 {
													boost += weight / 2
													break
												}
											}
										}
									}
								}
							}

							if hitMaterial != "" {
								parts := strings.Split(hitMaterial, ",")
								for _, part := range parts {
									part = strings.TrimSpace(strings.ToLower(part))
									if part != "" {
										if findSynonymIndex(descLower, part) != -1 {
											boost += 0.10
										} else {
											words := strings.Fields(part)
											for _, w := range words {
												w = strings.TrimFunc(w, func(r rune) bool {
													return r == ':' || r == '(' || r == ')' || r == '[' || r == ']' || r == ',' || r == '/'
												})
												if len(w) > 2 && findSynonymIndex(descLower, w) != -1 {
													boost += 0.05
													break
												}
											}
										}
									}
								}
							}

							finalScore := score + boost

							log.Printf("[Search Debug] Crop %q -> Candidate: %s, Raw Score: %f, Category Match: %t, Color Conflict: %t, Boost: %f, Final Score: %f",
								det.Desc, name, score, catMatch, colConflict, boost, finalScore)

							candLog := CandidateLog{
								Name:          name,
								Sku:           sku,
								RawScore:      score,
								CategoryMatch: catMatch,
								ColorConflict: colConflict,
								Boost:         boost,
								FinalScore:    finalScore,
							}

							logMu.Lock()
							imageLogs[imgIdx].Detections[detIdx].Candidates = append(imageLogs[imgIdx].Detections[detIdx].Candidates, candLog)
							logMu.Unlock()

							if !catMatch {
								continue
							}

							if colConflict {
								continue
							}

							// Require a minimum raw visual similarity prima di applicare il boost
							if score < 0.72 {
								continue
							}

							if finalScore > bestMatchScore {
								bestMatchScore = finalScore
								bestMatchName = name
								bestMatchImgUrl = imgUrl
								bestMatchSku = sku
							}
						}

						logMu.Lock()
						imageLogs[imgIdx].Detections[detIdx].BestMatch = bestMatchName
						imageLogs[imgIdx].Detections[detIdx].BestMatchScore = bestMatchScore
						imageLogs[imgIdx].Detections[detIdx].Accepted = (bestMatchScore >= 0.82)
						logMu.Unlock()

						// Require a final confidence score of at least 0.82
						if bestMatchScore >= 0.82 {
							mainMu.Lock()
							acceptedMatches = append(acceptedMatches, AcceptedMatch{
								Name:     bestMatchName,
								Sku:      bestMatchSku,
								ImageURL: bestMatchImgUrl,
								Score:    bestMatchScore,
							})
							log.Printf("Match accepted from img %d: %s (%s, %f)", imgIdx, bestMatchName, bestMatchSku, bestMatchScore)
							mainMu.Unlock()
						}
					}(detIdx, det)
				}
				wg.Wait()
			}(imgIdx, fileHeader)
		}
		outerWg.Wait()

		inventory, _ := h.qdrantClient.ListInventory()
		foundResults := []subdomain.ProductResult{}
		missingItems := []subdomain.MissingItem{}

		// Set of all found SKUs (matched by AI)
		foundSkus := make(map[string]bool)
		for _, am := range acceptedMatches {
			if am.Sku != "" {
				foundSkus[am.Sku] = true
			}
		}

		// Group acceptedMatches by SKU
		type GroupedMatch struct {
			Name     string
			Sku      string
			ImageURL string
			MaxScore float32
			Count    int
		}
		groupedMatches := make(map[string]*GroupedMatch)
		for _, am := range acceptedMatches {
			if am.Sku == "" {
				continue
			}
			if gm, ok := groupedMatches[am.Sku]; ok {
				gm.Count++
				if am.Score > gm.MaxScore {
					gm.MaxScore = am.Score
					gm.ImageURL = am.ImageURL
				}
			} else {
				groupedMatches[am.Sku] = &GroupedMatch{
					Name:     am.Name,
					Sku:      am.Sku,
					ImageURL: am.ImageURL,
					MaxScore: am.Score,
					Count:    1,
				}
			}
		}

		// Compile found results from groupedMatches
		for _, gm := range groupedMatches {
			foundResults = append(foundResults, subdomain.ProductResult{
				Name:     gm.Name,
				Sku:      gm.Sku,
				ImageURL: fixURL(gm.ImageURL, r.Host),
				Score:    gm.MaxScore,
				Count:    gm.Count,
			})
		}

		// Compile missing items by checking which SKUs in the inventory were NOT found.
		// Use a map to ensure we only list each missing SKU exactly once.
		addedMissingSkus := make(map[string]bool)
		for _, item := range inventory {
			payload, ok := item["payload"].(map[string]any)
			if !ok {
				continue
			}
			name, _ := payload["name"].(string)
			sku, _ := payload["sku"].(string)
			if sku == "" {
				continue
			}

			if !foundSkus[sku] {
				if !addedMissingSkus[sku] {
					addedMissingSkus[sku] = true
					imgUrl, _ := payload["imageUrl"].(string)
					missingItems = append(missingItems, subdomain.MissingItem{
						Name:     name,
						Sku:      sku,
						ImageURL: fixURL(imgUrl, r.Host),
					})
				}
			}
		}

		// Save request analysis log JSON
		var foundList []any
		for _, fr := range foundResults {
			foundList = append(foundList, fr)
		}
		var missingList []any
		for _, mi := range missingItems {
			missingList = append(missingList, mi)
		}

		reqLog := RequestLog{
			RequestID: reqID,
			Timestamp: time.Now().Format(time.RFC3339),
			Images:    imageLogs,
			Found:     foundList,
			Missing:   missingList,
		}

		reqLogData, marshalErr := json.MarshalIndent(reqLog, "", "  ")
		if marshalErr == nil {
			_ = os.WriteFile(filepath.Join(reqDir, "log.json"), reqLogData, 0644)
		} else {
			log.Printf("Warning: failed to marshal request log: %v", marshalErr)
		}

		log.Printf("Analysis complete across %d images, found %d products", len(imageFiles), len(foundResults))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domain.AnalysisResponse{
			Found:        foundResults,
			Missing:      missingItems,
			ImageResults: allImageResults,
		})
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)

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

func systemID(name string) int {
	h := 0
	for _, c := range name {
		h = 31*h + int(c)
	}
	if h < 0 {
		return -h
	}
	return h
}

func isCategoryMatch(detDesc string, productName string, aiDescription string) bool {
	detDesc = strings.ToLower(detDesc)
	productName = strings.ToLower(productName)
	aiDescription = strings.ToLower(aiDescription)

	// Smartwatch vs Analog watch distinction:
	// If the detection explicitly describes a smartwatch or digital watch,
	// it should not match a product that is a classic analog watch (doesn't contain smart/digital).
	isDetSmart := strings.Contains(detDesc, "smartwatch") || strings.Contains(detDesc, "smart watch") || strings.Contains(detDesc, "digital")
	isProdSmart := strings.Contains(productName, "smartwatch") || strings.Contains(productName, "smart watch") || strings.Contains(productName, "digital") || strings.Contains(aiDescription, "smartwatch") || strings.Contains(aiDescription, "smart watch") || strings.Contains(aiDescription, "digital")
	if isDetSmart && !isProdSmart {
		return false
	}

	categories := []struct {
		keywords []string
		names    []string
	}{
		{
			keywords: []string{"watch", "orolog", "nat ch", "face", "dial", "quadrante", "tempo", "smartwatch"},
			names:    []string{"watch", "orologio", "cronografo", "smartwatch"},
		},
		{
			keywords: []string{"necklace", "collana", "neck", "lace", "caten", "chain", "girocollo", "ciondolo", "collier", "pend"},
			names:    []string{"necklace", "collana", "pendant", "pendente", "girocollo", "ciondolo", "collier", "catena", "catenina"},
		},
		{
			keywords: []string{"earring", "orecchin", "ear", "ing", "lobo", "punti luce", "monachella", "pendenti"},
			names:    []string{"earring", "orecchini", "orecchino", "lobo", "monachella"},
		},
		{
			keywords: []string{"bracelet", "braccial", "brace", "polso", "chain", "caten", "bangle"},
			names:    []string{"bracelet", "bracciale", "braccialetto", "bangle"},
		},
		{
			keywords: []string{"ring", "anell", "fede", "solitario", "veretta"},
			names:    []string{"ring", "anello", "anelli", "fede", "fedina", "solitario", "veretta"},
		},
	}

	var prodCatIndex = -1
	for idx, cat := range categories {
		for _, n := range cat.names {
			if strings.Contains(productName, n) || (aiDescription != "" && strings.Contains(aiDescription, n)) {
				prodCatIndex = idx
				break
			}
		}
		if prodCatIndex != -1 {
			break
		}
	}

	if prodCatIndex == -1 {
		return true
	}

	cat := categories[prodCatIndex]
	for _, kw := range cat.keywords {
		if strings.Contains(detDesc, kw) {
			return true
		}
	}

	return false
}

func findSynonymIndex(text, synonym string) int {
	idx := 0
	for {
		i := strings.Index(text[idx:], synonym)
		if i == -1 {
			return -1
		}
		actualIdx := idx + i
		// Check start boundary
		startOk := true
		if actualIdx > 0 {
			prevChar := text[actualIdx-1]
			if (prevChar >= 'a' && prevChar <= 'z') || (prevChar >= '0' && prevChar <= '9') {
				startOk = false
			}
		}
		// Check end boundary
		endOk := true
		endIdx := actualIdx + len(synonym)
		if endIdx < len(text) {
			nextChar := text[endIdx]
			if (nextChar >= 'a' && nextChar <= 'z') || (nextChar >= '0' && nextChar <= '9') {
				// For exact color codes/names, we require absolute boundary
				if synonym == "tan" || synonym == "gold" || synonym == "pink" || synonym == "brown" || synonym == "black" || synonym == "white" || synonym == "oro" || synonym == "nero" || synonym == "rosa" || synonym == "marrone" || synonym == "beige" {
					endOk = false
				}
			}
		}
		if startOk && endOk {
			return actualIdx
		}
		idx = actualIdx + 1
	}
}

func hasColorConflict(detDesc string, productName string, hitColor string) bool {
	detDesc = strings.ToLower(detDesc)
	hitColor = strings.ToLower(hitColor)
	productName = strings.ToLower(productName)

	if hitColor == "" && productName == "" {
		return false
	}

	colors := []struct {
		name     string
		synonyms []string
	}{
		{name: "gold", synonyms: []string{"gold", "oro", "dorat"}},
		{name: "silver", synonyms: []string{"silver", "argent"}},
		{name: "black", synonyms: []string{"black", "nero", "scur"}},
		{name: "white", synonyms: []string{"white", "bianc", "chiar"}},
		{name: "pink", synonyms: []string{"pink", "rosa", "rosat"}},
		{name: "brown", synonyms: []string{"brown", "marrone", "beige", "tan"}},
	}

	var hitColorCat = ""
	for _, c := range colors {
		for _, syn := range c.synonyms {
			if findSynonymIndex(hitColor, syn) != -1 {
				hitColorCat = c.name
				break
			}
		}
		if hitColorCat != "" {
			break
		}
	}

	if hitColorCat == "" {
		for _, c := range colors {
			for _, syn := range c.synonyms {
				if findSynonymIndex(productName, syn) != -1 {
					hitColorCat = c.name
					break
				}
			}
			if hitColorCat != "" {
				break
			}
		}
	}

	if hitColorCat == "" {
		return false
	}

	// Find the FIRST color category mentioned in the detection description
	var firstColorCat = ""
	var firstColorIndex = -1
	for _, c := range colors {
		for _, syn := range c.synonyms {
			idx := findSynonymIndex(detDesc, syn)
			if idx != -1 {
				if firstColorIndex == -1 || idx < firstColorIndex {
					firstColorIndex = idx
					firstColorCat = c.name
				}
			}
		}
	}

	// If the description mentions a primary color, it must match the product's color
	if firstColorCat != "" && firstColorCat != hitColorCat {
		return true // Conflict!
	}

	return false
}

func parseDetections(responseText string) []subdomain.Detection {
	// 1. Try standard unmarshal of a direct array (backward compatibility)
	var detections []subdomain.Detection
	if err := json.Unmarshal([]byte(responseText), &detections); err == nil {
		return detections
	}

	// 2. Try standard unmarshal of the structured object response
	var objResp struct {
		ScaffaleVuoto bool                  `json:"scaffale_vuoto"`
		Spiegazione   string                `json:"spiegazione"`
		Articoli      []subdomain.Detection `json:"articoli"`
		Empty         bool                  `json:"empty"`
		Reasoning     string                `json:"reasoning"`
		Items         []subdomain.Detection `json:"items"`
	}
	if err := json.Unmarshal([]byte(responseText), &objResp); err == nil {
		if len(objResp.Articoli) > 0 {
			return objResp.Articoli
		}
		if len(objResp.Items) > 0 {
			return objResp.Items
		}
		return []subdomain.Detection{}
	}

	// 3. Fallback: parse any completed Detection sub-objects in a possibly truncated/malformed JSON string
	var recovered []subdomain.Detection
	type bracePos struct {
		index int
		depth int
	}
	var stack []bracePos
	depth := 0

	for i := 0; i < len(responseText); i++ {
		char := responseText[i]
		if char == '{' {
			stack = append(stack, bracePos{index: i, depth: depth})
			depth++
		} else if char == '}' {
			depth--
			if depth < 0 {
				depth = 0
			}
			if len(stack) > 0 {
				startBrace := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				objStr := responseText[startBrace.index : i+1]
				// Only parse if it looks like a Detection object (contains desc and box)
				if strings.Contains(objStr, "\"desc\"") && (strings.Contains(objStr, "\"box\"") || strings.Contains(objStr, "\"box_2d\"")) {
					var det subdomain.Detection
					if err := json.Unmarshal([]byte(objStr), &det); err == nil {
						// Ensure we got a valid box
						box := det.Box
						if len(box) != 4 {
							box = det.Box2D
						}
						if len(box) == 4 {
							recovered = append(recovered, det)
						}
					}
				}
			}
		}
	}

	if len(recovered) > 0 {
		log.Printf("Notice: Used fallback JSON parser. Recovered %d items.", len(recovered))
	} else {
		log.Printf("Warning: Failed to parse JSON and fallback parser found 0 items.")
	}

	return recovered
}
