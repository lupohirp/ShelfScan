package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"shelfscan-mcp/internal/domain"

	"github.com/gorilla/websocket"
)

func (h *Handler) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("MCP Server is running. Use /ws for WebSocket streaming."))
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket client connected")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break
		}

		var req domain.MCPRequest
		if err := json.Unmarshal(message, &req); err != nil {
			log.Printf("Unmarshal error: %v", err)
			continue
		}

		var res domain.MCPResponse
		res.JSONRPC = "2.0"
		res.ID = req.ID

		switch req.Method {
		case "tools/list":
			res.Result = map[string]interface{}{
				"tools": h.toolManager.GetTools(),
			}
		case "tools/call":
			var params domain.ToolCallParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				res.Error = map[string]interface{}{"code": -32602, "message": "Invalid params"}
			} else {
				result, err := h.toolManager.CallTool(params.Name, params.Arguments)
				if err != nil {
					res.Error = map[string]interface{}{"code": -32000, "message": err.Error()}
				} else if result == nil {
					res.Error = map[string]interface{}{"code": -32601, "message": "Tool not found"}
				} else {
					res.Result = result
				}
			}
		default:
			res.Error = map[string]interface{}{"code": -32601, "message": "Method not found"}
		}

		responseMsg, _ := json.Marshal(res)
		if err := conn.WriteMessage(websocket.TextMessage, responseMsg); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}
