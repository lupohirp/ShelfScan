package handler

import (
	"database/sql"
	"net/http"

	"shelfscan-api/internal/embedding"
	"shelfscan-api/internal/gemini"
	"shelfscan-api/internal/mcp"
	"shelfscan-api/internal/qdrant"
)

type Handler struct {
	corsMiddleware  func(http.HandlerFunc) http.HandlerFunc
	qdrantClient    *qdrant.QdrantClient
	embeddingClient *embedding.EmbeddingClient
	mcpClient       *mcp.MCPClient
	geminiClient    *gemini.GeminiClient
	db              *sql.DB
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

func (h *Handler) WithEmbeddingClient(client *embedding.EmbeddingClient) *Handler {
	h.embeddingClient = client
	return h
}

func (h *Handler) WithMCPClient(mcpClient *mcp.MCPClient) *Handler {
	h.mcpClient = mcpClient
	return h
}

func (h *Handler) WithGeminiClient(geminiClient *gemini.GeminiClient) *Handler {
	h.geminiClient = geminiClient
	return h
}

func (h *Handler) WithDB(db *sql.DB) *Handler {
	h.db = db
	return h
}

