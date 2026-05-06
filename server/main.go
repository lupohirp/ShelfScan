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

	http.HandleFunc("/upload", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Error retrieving the file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		name := r.FormValue("name")
		if name == "" {
			name = handler.Filename
		}

		// Forward to embedding service
		embedding, err := getEmbedding(embeddingsURL, file, handler.Filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error generating embedding: %v", err), http.StatusInternalServerError)
			return
		}

		// Save to Qdrant
		err = saveToQdrant(qdrantURL, name, embedding)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving to Qdrant: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully uploaded and indexed %s", name)
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

func saveToQdrant(url string, name string, vector []float32) error {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "qdrant", // From docker-compose service name
		Port: 6334,     // gRPC port
	})
	if err != nil {
		// Fallback for local testing if not in docker
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: "localhost",
			Port: 6334,
		})
		if err != nil {
			return err
		}
	}
	defer client.Close()

	upsertPoints := &qdrant.UpsertPoints{
		CollectionName: "jewelry_inventory",
		Points: []*qdrant.PointStruct{
			{
				Id: qdrant.NewIDNum(uint64(SystemID(name))),
				Vectors: qdrant.NewVectorsDense(vector),
				Payload: qdrant.NewValueMap(map[string]any{
					"name": name,
				}),
			},
		},
	}

	_, err = client.Upsert(context.Background(), upsertPoints)
	return err
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
