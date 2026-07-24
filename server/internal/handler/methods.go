package handler

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"shelfscan-api/internal/domain"
	"shelfscan-api/internal/gemini"
	"sort"
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
		appendMode := r.FormValue("append") == "true" || r.URL.Query().Get("append") == "true"
		existingAiDesc := strings.TrimSpace(r.FormValue("ai_description"))

		files := r.MultipartForm.File["images"]
		if len(files) == 0 {
			http.Error(w, "No images uploaded", http.StatusBadRequest)
			return
		}

		startIndex := 0
		if appendMode {
			n, err := h.qdrantClient.CountPointsBySKU(sku)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error counting existing points for SKU %s: %v", sku, err), http.StatusInternalServerError)
				return
			}
			startIndex = n
			log.Printf("Append mode: SKU %s has %d existing point(s); new indices will start at %d", sku, n, startIndex)
		}

		os.MkdirAll("uploads", 0755)
		var savedURLs []string
		var savedThumbURLs []string

		var embeddings [][]float32
		aiDesc := existingAiDesc
		for i, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Error opening file", http.StatusInternalServerError)
				return
			}

			if i == 0 && aiDesc == "" {
				imgBytes, readErr := io.ReadAll(file)
				if readErr == nil {
					prompt := "Descrivi brevemente l'articolo di gioielleria o orologio in questa immagine specificando esplicitamente la categoria (es. orologio, collana, anello, bracciale, orecchini), lo stile e i colori rilevanti in italiano."
					desc, err := h.geminiClient.DescribeImageWithModel(r.Context(), "models/gemini-3.6-flash", prompt, imgBytes)
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

			filename := fmt.Sprintf("%d_%s_%d_%s", systemID(name), sku, startIndex+i, fileHeader.Filename)
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

		err = h.qdrantClient.SaveMultipleToQdrantAt(name, sku, savedURLs, savedThumbURLs, color, material, aiDesc, embeddings, startIndex)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving to Qdrant: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		mode := "indexed"
		if appendMode {
			mode = fmt.Sprintf("appended (starting at index %d)", startIndex)
		}
		fmt.Fprintf(w, "Successfully uploaded and %s %d view(s) for %s", mode, len(embeddings), name)
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
	Desc              string         `json:"desc"`
	Box               []int          `json:"box"`
	CropFilename      string         `json:"crop_filename,omitempty"`
	BestMatch         string         `json:"best_match,omitempty"`
	BestMatchScore    float32        `json:"best_match_score,omitempty"`
	Accepted          bool           `json:"accepted"`
	Verified          bool           `json:"verified,omitempty"`
	VerificationNotes string         `json:"verification_notes,omitempty"`
	Candidates        []CandidateLog `json:"candidates"`
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
			CropURL  string
			Score    float32
		}
		var acceptedMatches []AcceptedMatch
		allImageResults := make([]domain.ImageResult, len(imageFiles))
		imageLogs := make([]ImageLog, len(imageFiles))
		var logMu sync.Mutex
		var mainMu sync.Mutex
		sem := make(chan struct{}, 2)

		isStreaming := strings.Contains(r.Header.Get("Accept"), "application/x-ndjson") || r.URL.Query().Get("stream") == "true"
		var flusher http.Flusher
		if isStreaming {
			w.Header().Set("Content-Type", "application/x-ndjson")
			w.Header().Set("Cache-Control", "no-cache, no-transform")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("X-Accel-Buffering", "no")
			if f, ok := w.(http.Flusher); ok {
				flusher = f
			}
		}

		var chunkMu sync.Mutex
		sendChunk := func(data any) {
			if isStreaming {
				chunkMu.Lock()
				defer chunkMu.Unlock()
				_ = json.NewEncoder(w).Encode(data)
				if flusher != nil {
					flusher.Flush()
				}
			}
		}

		for imgIdx, fileHeader := range imageFiles {
			log.Printf("Processing image %d/%d: %s", imgIdx+1, len(imageFiles), fileHeader.Filename)
			sendChunk(map[string]any{
				"type":     "image_start",
				"image":    imgIdx + 1,
				"total":    len(imageFiles),
				"filename": fileHeader.Filename,
			})
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error opening image %d: %v", imgIdx, err)
				continue
			}
			imgData, _ := io.ReadAll(file)
			file.Close()

			// Auto-rotate uploaded image bytes if EXIF orientation is present
			imgData = autoRotateBytes(imgData)

			// Save original (now guaranteed upright) image to log directory
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
				continue
			}

			bounds := img.Bounds()
			width, height := bounds.Dx(), bounds.Dy()

			var geminiImgData []byte
			if width > 1920 || height > 1080 {
				resized := imaging.Fit(img, 1920, 1080, imaging.Lanczos)
				buf := new(bytes.Buffer)
				jpeg.Encode(buf, resized, &jpeg.Options{Quality: 85})
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
				continue
			}
			log.Printf("Gemini Finish Reason for image %d (model: %s): %v", imgIdx, modelUsed, resp.Candidates[0].FinishReason)

			var responseText string
			for _, part := range resp.Candidates[0].Content.Parts {
				if text, ok := part.(genai.Text); ok {
					responseText += string(text)
				}
			}

			log.Printf("Gemini Response: %s", responseText)

			detections := parseDetections(responseText)

			logMu.Lock()
			imageLogs[imgIdx].GeminiModel = modelUsed
			imageLogs[imgIdx].GeminiRawResponse = responseText
			imageLogs[imgIdx].Detections = make([]DetectionLog, len(detections))
			logMu.Unlock()

			var detWg sync.WaitGroup
			detSem := make(chan struct{}, 5)

			for detIdx, det := range detections {
				detWg.Add(1)
				go func(detIdx int, det domain.Detection) {
					defer detWg.Done()
					detSem <- struct{}{}
					defer func() { <-detSem }()

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

						// Add 10% bounding box padding so straps/edges/pendants are never cut off
						boxH := ymax - ymin
						boxW := xmax - xmin
						padH := boxH / 10
						padW := boxW / 10

						cropYmin := ymin - padH
						if cropYmin < 0 {
							cropYmin = 0
						}
						cropYmax := ymax + padH
						if cropYmax > height {
							cropYmax = height
						}
						cropXmin := xmin - padW
						if cropXmin < 0 {
							cropXmin = 0
						}
						cropXmax := xmax + padW
						if cropXmax > width {
							cropXmax = width
						}

						if cropXmin >= cropXmax || cropYmin >= cropYmax {
							return
						}

						cropped := imaging.Crop(img, image.Rect(cropXmin, cropYmin, cropXmax, cropYmax))
						cropBuf := new(bytes.Buffer)
						jpeg.Encode(cropBuf, cropped, nil)

						// Save cropped image to log directory
						cropFilename := fmt.Sprintf("crop_%d_%d.jpg", imgIdx, detIdx)
						_ = os.WriteFile(filepath.Join(reqDir, cropFilename), cropBuf.Bytes(), 0644)

						// Also save to uploads directory so it can be served via HTTP
						cropUploadFilename := fmt.Sprintf("crop_%s_%d_%d.jpg", reqID, imgIdx, detIdx)
						if err := os.WriteFile(filepath.Join("uploads", cropUploadFilename), cropBuf.Bytes(), 0644); err != nil {
							log.Printf("Error writing crop file %s: %v", cropUploadFilename, err)
						}
						cropURL := "/uploads/" + cropUploadFilename

						detections[detIdx].CropURL = fixURL(cropURL, r.Host)

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

						type ViableCandidate struct {
							Name       string
							Sku        string
							ImgURL     string
							FinalScore float32
						}
						var viableCandidates []ViableCandidate

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
							sku, _ := hit["sku"].(string)

							var score float32
							if s, ok := hit["score"].(float32); ok {
								score = s
							} else if s, ok := hit["score"].(float64); ok {
								score = float32(s)
							}

							catMatch := isCategoryMatch(det.Desc, name, aiDesc)
							colConflict := hasColorConflict(det.Desc, name, hitColor)

							// Disable artificial keyword boost (which distorts visual vector ranking)
							var boost float32 = 0
							finalScore := score

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

							if catMatch && !colConflict && score >= 0.70 && finalScore >= 0.75 {
								viableCandidates = append(viableCandidates, ViableCandidate{
									Name:       name,
									Sku:        sku,
									ImgURL:     imgUrl,
									FinalScore: finalScore,
								})
							}
						}

						sort.Slice(viableCandidates, func(i, j int) bool {
							return viableCandidates[i].FinalScore > viableCandidates[j].FinalScore
						})

						accepted := false
						var verifiedCandidate ViableCandidate
						var verificationReason string

						var checkCandidates []gemini.VerificationCandidate
						var candidateMap []ViableCandidate
						seenSkus := make(map[string]bool)

						for _, candidate := range viableCandidates {
							if len(checkCandidates) >= 5 { // Prepare top 5 UNIQUE SKUs for batch verification
								break
							}
							if candidate.ImgURL == "" || seenSkus[candidate.Sku] {
								continue
							}
							seenSkus[candidate.Sku] = true

							dbImgFilename := filepath.Base(candidate.ImgURL)
							dbImgPath := filepath.Join("uploads", dbImgFilename)

							dbImgData, err := os.ReadFile(dbImgPath)
							if err != nil {
								log.Printf("[Verification Warning] Could not read database image file %s: %v", dbImgPath, err)
								continue
							}

							checkCandidates = append(checkCandidates, gemini.VerificationCandidate{
								Sku:     candidate.Sku,
								Name:    candidate.Name,
								ImgData: dbImgData,
							})
							candidateMap = append(candidateMap, candidate)
						}

						if len(checkCandidates) > 0 {
							category := getProductCategory(candidateMap[0].Name)
							log.Printf("[1-on-1 Parallel Verification] Testing %d candidate SKUs concurrently for det %d (%s)", len(checkCandidates), detIdx, det.Desc)

							type candRes struct {
								matched bool
								reason  string
								cand    ViableCandidate
							}
							resChan := make(chan candRes, len(checkCandidates))
							var cWg sync.WaitGroup

							for candIdx, cand := range checkCandidates {
								cWg.Add(1)
								go func(c gemini.VerificationCandidate, vCand ViableCandidate) {
									defer cWg.Done()
									matched, reason, err := h.geminiClient.VerifyMatch(context.Background(), cropBuf.Bytes(), c.ImgData, category)
									if err == nil {
										resChan <- candRes{matched: matched, reason: reason, cand: vCand}
									} else {
										log.Printf("[1-on-1 Warning] Gemini verification failed for SKU %s: %v", c.Sku, err)
										resChan <- candRes{matched: false, reason: err.Error(), cand: vCand}
									}
								}(cand, candidateMap[candIdx])
							}
							cWg.Wait()
							close(resChan)

							for r := range resChan {
								log.Printf("[1-on-1 Result] det %d (%s): match=%v, Reason: %s", detIdx, r.cand.Sku, r.matched, r.reason)
								if r.matched {
									if !accepted || r.cand.FinalScore > verifiedCandidate.FinalScore {
										accepted = true
										verifiedCandidate = r.cand
										verificationReason = r.reason
									}
								}
							}
						}

						logMu.Lock()
						if accepted {
							imageLogs[imgIdx].Detections[detIdx].BestMatch = verifiedCandidate.Name
							imageLogs[imgIdx].Detections[detIdx].BestMatchScore = verifiedCandidate.FinalScore
							imageLogs[imgIdx].Detections[detIdx].Accepted = true
							imageLogs[imgIdx].Detections[detIdx].Verified = true
							imageLogs[imgIdx].Detections[detIdx].VerificationNotes = verificationReason
						} else if len(viableCandidates) > 0 {
							imageLogs[imgIdx].Detections[detIdx].BestMatch = viableCandidates[0].Name
							imageLogs[imgIdx].Detections[detIdx].BestMatchScore = viableCandidates[0].FinalScore
							imageLogs[imgIdx].Detections[detIdx].Accepted = false
							imageLogs[imgIdx].Detections[detIdx].Verified = false
							imageLogs[imgIdx].Detections[detIdx].VerificationNotes = "All candidates rejected by Gemini side-by-side verification"
						}
						logMu.Unlock()

						if accepted {
							mainMu.Lock()
							acceptedMatches = append(acceptedMatches, AcceptedMatch{
								Name:     verifiedCandidate.Name,
								Sku:      verifiedCandidate.Sku,
								ImageURL: verifiedCandidate.ImgURL,
								CropURL:  cropURL,
								Score:    verifiedCandidate.FinalScore,
							})
							log.Printf("Match accepted from img %d: %s (%s, %f)", imgIdx, verifiedCandidate.Name, verifiedCandidate.Sku, verifiedCandidate.FinalScore)
							mainMu.Unlock()
							detections[detIdx].SKU = verifiedCandidate.Sku

							sendChunk(map[string]any{
								"type":     "item_matched",
								"sku":      verifiedCandidate.Sku,
								"name":     verifiedCandidate.Name,
								"imageUrl": fixURL(verifiedCandidate.ImgURL, r.Host),
								"cropUrl":  fixURL(cropURL, r.Host),
								"score":    verifiedCandidate.FinalScore,
							})
						} else {
							log.Printf("No candidate accepted by Gemini verification for detection %d (%s)", detIdx, det.Desc)
						}
				}(detIdx, det)
			}
			detWg.Wait()
			allImageResults[imgIdx] = domain.ImageResult{Detections: detections}
		}

		inventory, _ := h.qdrantClient.ListInventory()
		foundResults := []domain.ProductResult{}
		missingItems := []domain.MissingItem{}

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
			CropURL  string
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
					gm.Name = am.Name
					gm.CropURL = am.CropURL
				}
			} else {
				groupedMatches[am.Sku] = &GroupedMatch{
					Name:     am.Name,
					Sku:      am.Sku,
					ImageURL: am.ImageURL,
					CropURL:  am.CropURL,
					MaxScore: am.Score,
					Count:    1,
				}
			}
		}

		// Compile found results from groupedMatches
		for _, gm := range groupedMatches {
			foundResults = append(foundResults, domain.ProductResult{
				Name:     gm.Name,
				Sku:      gm.Sku,
				ImageURL: fixURL(gm.ImageURL, r.Host),
				CropURL:  fixURL(gm.CropURL, r.Host),
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
					missingItems = append(missingItems, domain.MissingItem{
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

		finalResponse := domain.AnalysisResponse{
			Found:        foundResults,
			Missing:      missingItems,
			ImageResults: allImageResults,
		}

		if isStreaming {
			sendChunk(map[string]any{
				"type": "complete",
				"data": finalResponse,
			})
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(finalResponse)
		}
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)

}

func isLocalHost(host string) bool {
	return strings.Contains(host, "localhost") ||
		strings.Contains(host, "127.0.0.1") ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.")
}

func fixURL(imgURL string, host string) string {
	if imgURL == "" {
		return ""
	}

	protocol := "https://"
	if isLocalHost(host) {
		protocol = "http://"
	}

	if strings.HasPrefix(imgURL, "http://") || strings.HasPrefix(imgURL, "https://") {
		if strings.Contains(imgURL, "localhost") && !isLocalHost(host) {
			imgURL = strings.Replace(imgURL, "localhost:8080", host, 1)
		}
		if idx := strings.Index(imgURL, "/uploads/"); idx != -1 {
			return protocol + host + imgURL[idx:]
		}
		if protocol == "https://" && strings.HasPrefix(imgURL, "http://") {
			return "https://" + imgURL[7:]
		}
		if protocol == "http://" && strings.HasPrefix(imgURL, "https://") {
			return "http://" + imgURL[8:]
		}
		return imgURL
	}

	if strings.HasPrefix(imgURL, "/") {
		return protocol + host + imgURL
	}
	return protocol + host + "/uploads/" + imgURL
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
	// Disabled brittle heuristic color conflict check.
	// Watches often have multiple colors (e.g. silver strap, white dial).
	// The VerifyMatch LLM will reliably catch any real color mismatches.
	return false
}

func parseDetections(responseText string) []domain.Detection {
	// 1. Try standard unmarshal of a direct array (backward compatibility)
	var detections []domain.Detection
	if err := json.Unmarshal([]byte(responseText), &detections); err == nil {
		return detections
	}

	// 2. Try standard unmarshal of the structured object response
	var objResp struct {
		ScaffaleVuoto bool                  `json:"scaffale_vuoto"`
		Spiegazione   string                `json:"spiegazione"`
		Articoli      []domain.Detection `json:"articoli"`
		Empty         bool               `json:"empty"`
		Reasoning     string             `json:"reasoning"`
		Items         []domain.Detection `json:"items"`
	}
	if err := json.Unmarshal([]byte(responseText), &objResp); err == nil {
		if len(objResp.Articoli) > 0 {
			return objResp.Articoli
		}
		if len(objResp.Items) > 0 {
			return objResp.Items
		}
		return []domain.Detection{}
	}

	// 3. Fallback: parse any completed Detection sub-objects in a possibly truncated/malformed JSON string
	var recovered []domain.Detection
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
					var det domain.Detection
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

func getOrientation(data []byte) int {
	if len(data) < 4 {
		return 1
	}
	if data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}

	idx := 2
	for idx < len(data)-1 {
		if data[idx] != 0xff {
			return 1
		}
		marker := data[idx+1]
		if marker == 0xd9 || marker == 0xda { // EOI or SOS
			break
		}
		if marker == 0xe1 { // APP1 (usually Exif)
			if idx+6+6 > len(data) {
				return 1
			}
			if string(data[idx+4:idx+8]) == "Exif" && data[idx+8] == 0 && data[idx+9] == 0 {
				tiffHeaderIdx := idx + 10
				if tiffHeaderIdx+8 > len(data) {
					return 1
				}
				byteOrder := string(data[tiffHeaderIdx : tiffHeaderIdx+2])
				var isLittleEndian bool
				if byteOrder == "II" {
					isLittleEndian = true
				} else if byteOrder == "MM" {
					isLittleEndian = false
				} else {
					return 1
				}

				var readUint16 func([]byte) uint16
				var readUint32 func([]byte) uint32
				if isLittleEndian {
					readUint16 = binary.LittleEndian.Uint16
					readUint32 = binary.LittleEndian.Uint32
				} else {
					readUint16 = binary.BigEndian.Uint16
					readUint32 = binary.BigEndian.Uint32
				}

				if readUint16(data[tiffHeaderIdx+2:tiffHeaderIdx+4]) != 42 {
					return 1
				}

				ifdOffset := readUint32(data[tiffHeaderIdx+4:tiffHeaderIdx+8])
				ifdIdx := tiffHeaderIdx + int(ifdOffset)
				if ifdIdx+2 > len(data) {
					return 1
				}

				numEntries := readUint16(data[ifdIdx : ifdIdx+2])
				dirIdx := ifdIdx + 2
				for i := 0; i < int(numEntries); i++ {
					entryIdx := dirIdx + i*12
					if entryIdx+12 > len(data) {
						break
					}
					tag := readUint16(data[entryIdx : entryIdx+2])
					if tag == 0x0112 { // Orientation Tag
						val := readUint16(data[entryIdx+8 : entryIdx+10])
						return int(val)
					}
				}
			}
		}
		if idx+4 > len(data) {
			break
		}
		segmentLength := int(binary.BigEndian.Uint16(data[idx+2 : idx+4]))
		idx += 2 + segmentLength
	}
	return 1
}

func getProductCategory(productName string) string {
	productName = strings.ToLower(productName)

	if strings.Contains(productName, "watch") || strings.Contains(productName, "orologio") || strings.Contains(productName, "cronografo") || strings.Contains(productName, "smartwatch") {
		return "orologio"
	}
	if strings.Contains(productName, "necklace") || strings.Contains(productName, "collana") || strings.Contains(productName, "pendant") || strings.Contains(productName, "pendente") || strings.Contains(productName, "girocollo") || strings.Contains(productName, "ciondolo") || strings.Contains(productName, "collier") || strings.Contains(productName, "catena") || strings.Contains(productName, "catenina") {
		return "collana"
	}
	if strings.Contains(productName, "earring") || strings.Contains(productName, "orecchini") || strings.Contains(productName, "orecchino") || strings.Contains(productName, "lobo") || strings.Contains(productName, "monachella") {
		return "orecchini"
	}
	if strings.Contains(productName, "bracelet") || strings.Contains(productName, "bracciale") || strings.Contains(productName, "braccialetto") || strings.Contains(productName, "bangle") {
		return "bracciale"
	}
	if strings.Contains(productName, "ring") || strings.Contains(productName, "anello") || strings.Contains(productName, "anelli") || strings.Contains(productName, "fede") || strings.Contains(productName, "fedina") || strings.Contains(productName, "solitario") || strings.Contains(productName, "veretta") {
		return "anello"
	}
	return "altro"
}

func autoRotateBytes(imgData []byte) []byte {
	orientation := getOrientation(imgData)
	if orientation <= 1 {
		return imgData
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return imgData
	}

	log.Printf("Applying auto-orientation rotation: %d", orientation)
	switch orientation {
	case 2:
		img = imaging.FlipH(img)
	case 3:
		img = imaging.Rotate180(img)
	case 4:
		img = imaging.FlipV(img)
	case 5:
		img = imaging.Rotate90(imaging.FlipH(img))
	case 6:
		img = imaging.Rotate270(img) // imaging rotates counter-clockwise, so Rotate270 is 90 degrees CW
	case 7:
		img = imaging.Rotate90(imaging.FlipV(img))
	case 8:
		img = imaging.Rotate90(img) // imaging rotates counter-clockwise, so Rotate90 is 90 degrees CCW (270 CW)
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90}); err == nil {
		return buf.Bytes()
	}
	return imgData
}

