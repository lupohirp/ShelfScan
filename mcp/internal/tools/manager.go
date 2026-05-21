package tools

import (
	"encoding/json"
	"shelfscan-mcp/internal/qdrant"
)

type Tool struct {
	Name        string
	Description string
	InputSchema interface{}
	Execute     func(args json.RawMessage) (interface{}, error)
}

type ToolManager struct {
	qdrantClient *qdrant.QdrantClient
	tools        map[string]Tool
}

func NewToolManager(qdrantClient *qdrant.QdrantClient) *ToolManager {
	tm := &ToolManager{
		qdrantClient: qdrantClient,
		tools:        make(map[string]Tool),
	}
	tm.init()
	return tm
}

func (tm *ToolManager) GetTools() []map[string]interface{} {
	var list []map[string]interface{}
	for _, t := range tm.tools {
		list = append(list, map[string]interface{}{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	return list
}

func (tm *ToolManager) CallTool(name string, args json.RawMessage) (interface{}, error) {
	tool, ok := tm.tools[name]
	if !ok {
		return nil, nil // or error
	}
	return tool.Execute(args)
}
