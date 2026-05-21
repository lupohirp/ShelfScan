package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func (c *MCPClient) Connect() (*MCPClient, error) {
	if c.url == "" {
		return nil, fmt.Errorf("URL is required")
	}
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return c, nil
}

func (c *MCPClient) Close() {
	if c.conn == nil {
		return
	}
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close()
}

func (c *MCPClient) CallVectorSearch(embedding []float32) (string, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "vector_search",
			"arguments": map[string]interface{}{
				"embedding": embedding,
			},
		},
	}
	err := c.conn.WriteJSON(req)
	if err != nil {
		return "", err
	}

	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return "", err
	}

	var res struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(message, &res); err != nil {
		return "", err
	}
	if len(res.Result.Content) == 0 {
		return "", fmt.Errorf("no content in result")
	}
	return res.Result.Content[0].Text, nil
}
