package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/qdrant/go-client/qdrant"
)

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow only localhost for development
		if origin == "http://localhost:5173" || origin == "http://localhost:5174" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
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
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		err := r.ParseMultipartForm(50 << 20) // 50MB for multiple images
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

		var embeddings [][]float32
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Error opening file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Forward to embedding service
			embedding, err := getEmbedding(embeddingsURL, file, fileHeader.Filename)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error generating embedding for %s: %v", fileHeader.Filename, err), http.StatusInternalServerError)
				return
			}
			embeddings = append(embeddings, embedding)
		}

		// Save all views to Qdrant
		err = saveMultipleToQdrant(qdrantURL, name, embeddings)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving to Qdrant: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully uploaded and indexed %d views for %s", len(embeddings), name)
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
			payload[k] = v.GetStringValue() // For now we only use name which is a string
		}

		item := map[string]any{
			"id":      p.Id.GetNum(),
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

func saveMultipleToQdrant(url string, name string, vectors [][]float32) error {
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
		// Use a composite ID based on name and index to allow multiple views
		id := uint64(SystemID(fmt.Sprintf("%s_%d", name, i)))
		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(id),
			Vectors: qdrant.NewVectorsDense(vector),
			Payload: qdrant.NewValueMap(map[string]any{
				"name": name,
			}),
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
			"name":  payload["name"],
			"score": hit.Score,
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
