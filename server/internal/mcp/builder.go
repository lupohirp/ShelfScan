package mcp

import "github.com/gorilla/websocket"

type MCPClient struct {
	conn *websocket.Conn
	url  string
}

func NewMCPClient() *MCPClient {
	return &MCPClient{}
}

func (c *MCPClient) WithURL(url string) *MCPClient {
	c.url = url
	return c
}
