package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"shelfscan-mcp/internal/handler"
	"shelfscan-mcp/internal/qdrant"
	"shelfscan-mcp/internal/tools"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	qdrantHost := os.Getenv("QDRANT_HOST")
	if qdrantHost == "" {
		qdrantHost = "qdrant"
	}

	qdrantPortStr := os.Getenv("QDRANT_PORT")
	if qdrantPortStr == "" {
		qdrantPortStr = "6334"
	}
	qdrantPort, _ := strconv.Atoi(qdrantPortStr)

	qdrantClient := qdrant.NewQdrantClient().
		WithHost(qdrantHost).
		WithPort(qdrantPort)

	toolManager := tools.NewToolManager(qdrantClient)

	handlers := handler.NewHandler().
		WithQdrantClient(qdrantClient).
		WithToolManager(toolManager)

	http.HandleFunc("/ws", handlers.WebSocketHandler)
	http.HandleFunc("/", handlers.RootHandler)

	log.Printf("MCP Streaming Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
