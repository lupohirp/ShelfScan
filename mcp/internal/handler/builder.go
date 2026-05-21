package handler

import (
	"net/http"
	"shelfscan-mcp/internal/qdrant"
	"shelfscan-mcp/internal/tools"

	"github.com/gorilla/websocket"
)

type Handler struct {
	qdrantClient *qdrant.QdrantClient
	toolManager  *tools.ToolManager
	upgrader     websocket.Upgrader
}

func NewHandler() *Handler {
	return &Handler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Handler) WithQdrantClient(client *qdrant.QdrantClient) *Handler {
	h.qdrantClient = client
	return h
}

func (h *Handler) WithToolManager(tm *tools.ToolManager) *Handler {
	h.toolManager = tm
	return h
}
