package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/qdrant/go-client/qdrant"
)

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/", handleMCP)

	log.Printf("MCP Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var res MCPResponse
	res.JSONRPC = "2.0"
	res.ID = req.ID

	switch req.Method {
	case "tools/list":
		res.Result = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "vector_search",
					"description": "Search for jewelry items in Qdrant using an embedding vector",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"embedding": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
						},
						"required": []string{"embedding"},
					},
				},
				{
					"name":        "get_layout",
					"description": "Get the expected layout for a specific shelf",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"shelf_id": map[string]interface{}{"type": "string"},
						},
						"required": []string{"shelf_id"},
					},
				},
			},
		}
	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			res.Error = map[string]interface{}{"code": -32602, "message": "Invalid params"}
			break
		}

		switch params.Name {
		case "vector_search":
			var args struct {
				Embedding []float32 `json:"embedding"`
			}
			json.Unmarshal(params.Arguments, &args)
			results, err := performVectorSearch(args.Embedding)
			if err != nil {
				res.Error = map[string]interface{}{"code": -32000, "message": err.Error()}
			} else {
				res.Result = map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": results},
					},
				}
			}
		case "get_layout":
			// Mock layout
			res.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "Expected items: [Diamond Ring A, Gold Necklace B, Emerald Studs C]"},
				},
			}
		default:
			res.Error = map[string]interface{}{"code": -32601, "message": "Tool not found"}
		}
	default:
		res.Error = map[string]interface{}{"code": -32601, "message": "Method not found"}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func performVectorSearch(vector []float32) (string, error) {
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
			return "", err
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
		return "", err
	}

	output, _ := json.Marshal(searchResult)
	return string(output), nil
}
