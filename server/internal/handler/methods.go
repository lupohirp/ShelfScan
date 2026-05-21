package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
