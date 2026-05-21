package handler

import (
	"net/http"

	"shelfscan-api/qdrant"
)

type Handler struct {
	corsMiddleware func(http.HandlerFunc) http.HandlerFunc
	qdrantClient   *qdrant.QdrantClient
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) WithCorsMiddleware(middleware func(http.HandlerFunc) http.HandlerFunc) *Handler {
	h.corsMiddleware = middleware
	return h
}

func (h *Handler) WithQdrantClient(client *qdrant.QdrantClient) *Handler {
	h.qdrantClient = client
	return h
}
