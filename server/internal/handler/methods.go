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
	"shelfscan-api/internal/domain"
	subdomain "shelfscan-api/internal/domain/sub"
	"strings"
	"sync"

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
				filename := fmt.Sprintf("%d_%s", systemID(name), fileHeader.Filename)
				dst, err := os.Create("uploads/" + filename)
				if err == nil {
					defer dst.Close()
					io.Copy(dst, file)
					file.Seek(0, 0)
					savedURL = "/uploads/" + filename
				}
			}

			embedding, err := h.embeddingClient.GetEmbedding(file, fileHeader.Filename)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error generating embedding for %s: %v", fileHeader.Filename, err), http.StatusInternalServerError)
				return
			}
			embeddings = append(embeddings, embedding)
		}

		err = h.qdrantClient.SaveMultipleToQdrant(name, savedURL, color, material, embeddings)
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

		productMaxScores := make(map[string]float32)
		productImageURLs := make(map[string]string)
		allImageResults := make([]subdomain.ImageResult, len(imageFiles))

		var mainMu sync.Mutex
		var outerWg sync.WaitGroup

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

				img, _, err := image.Decode(bytes.NewReader(imgData))
				if err != nil {
					log.Printf("Error decoding image %d: %v", imgIdx, err)
					return
				}

				bounds := img.Bounds()
				width, height := bounds.Dx(), bounds.Dy()

				var geminiImgData []byte
				if width > 1600 || height > 1600 {
					resized := imaging.Fit(img, 1600, 1600, imaging.Linear)
					buf := new(bytes.Buffer)
					jpeg.Encode(buf, resized, &jpeg.Options{Quality: 80})
					geminiImgData = buf.Bytes()
				} else {
					buf := new(bytes.Buffer)
					jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
					geminiImgData = buf.Bytes()
				}

				ctx := context.Background()

				model := h.geminiClient.GetClient(ctx)
				if model == nil {
					log.Printf("Gemini model is nil, client initialization failed")
					return
				}

				var resp *genai.GenerateContentResponse
				for i := 0; i < 2; i++ {
					resp, err = model.GenerateContent(ctx, genai.Text(h.geminiClient.Prompt), genai.ImageData("jpeg", geminiImgData))
					if err == nil {
						break
					}
				}
				if err != nil {
					log.Printf("AI error for image %d: %v", imgIdx, err)
					return
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

				var detections []subdomain.Detection
				json.Unmarshal([]byte(responseText), &detections)
				allImageResults[imgIdx] = subdomain.ImageResult{Detections: detections}

				var wg sync.WaitGroup
				for _, det := range detections {
					wg.Add(1)
					go func(det subdomain.Detection) {
						defer wg.Done()

						box := det.Box
						if len(box) != 4 {
							box = det.Box2D
						}
						if len(box) != 4 {
							return
						}

						ymin, xmin, ymax, xmax := box[0]*height/1000, box[1]*width/1000, box[2]*height/1000, box[3]*width/1000
						if xmin >= xmax || ymin >= ymax || xmin < 0 || ymin < 0 || xmax > width || ymax > height {
							return
						}

						cropped := imaging.Crop(img, image.Rect(xmin, ymin, xmax, ymax))
						cropBuf := new(bytes.Buffer)
						jpeg.Encode(cropBuf, cropped, nil)

						emb, err := h.embeddingClient.GetEmbeddingBytes(cropBuf.Bytes(), "crop.jpg")
						if err != nil {
							return
						}

						hits, err := h.qdrantClient.PerformVectorSearch(emb)
						if err != nil {
							return
						}

						var bestMatchName string
						var bestMatchImgUrl string
						var bestMatchScore float32 = -1.0

						for _, hit := range hits {
							name, okName := hit["name"].(string)
							imgUrl, okImg := hit["imageUrl"].(string)
							if !okName || !okImg {
								continue
							}

							if !isCategoryMatch(det.Desc, name) {
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

							// Require a minimum raw visual similarity before applying any text boost
							if score < 0.65 {
								continue
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

						// Require a final confidence score of at least 0.75
						if bestMatchScore >= 0.75 {
							mainMu.Lock()
							if currentMax, ok := productMaxScores[bestMatchName]; !ok || bestMatchScore > currentMax {
								productMaxScores[bestMatchName] = bestMatchScore
								productImageURLs[bestMatchName] = bestMatchImgUrl
								log.Printf("Match accepted from img %d: %s (%f)", imgIdx, bestMatchName, bestMatchScore)
							}
							mainMu.Unlock()
						}
					}(det)
				}
				wg.Wait()
			}(imgIdx, fileHeader)
		}
		outerWg.Wait()

		inventory, _ := h.qdrantClient.ListInventory()
		foundResults := []subdomain.ProductResult{}
		missingItems := []subdomain.MissingItem{}

		for _, item := range inventory {
			payload, ok := item["payload"].(map[string]any)
			if !ok {
				continue
			}
			name, _ := payload["name"].(string)

			if score, found := productMaxScores[name]; found {
				foundResults = append(foundResults, subdomain.ProductResult{
					Name:     name,
					ImageURL: fixURL(productImageURLs[name], r.Host),
					Score:    score,
				})
			} else {
				imgUrl, _ := payload["imageUrl"].(string)
				missingItems = append(missingItems, subdomain.MissingItem{
					Name:     name,
					ImageURL: fixURL(imgUrl, r.Host),
				})
			}
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

func isCategoryMatch(detDesc string, productName string) bool {
	detDesc = strings.ToLower(detDesc)
	productName = strings.ToLower(productName)

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
			if strings.Contains(productName, n) {
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
