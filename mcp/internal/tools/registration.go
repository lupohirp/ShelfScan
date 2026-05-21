package tools

func (tm *ToolManager) init() {
	tm.tools["vector_search"] = Tool{
		Name:        "vector_search",
		Description: "Search for jewelry items in Qdrant using an embedding vector",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"embedding": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
			},
			"required": []string{"embedding"},
		},
		Execute: tm.vectorSearch,
	}

	tm.tools["get_layout"] = Tool{
		Name:        "get_layout",
		Description: "Get the expected layout for a specific shelf",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"shelf_id": map[string]interface{}{"type": "string"},
			},
			"required": []string{"shelf_id"},
		},
		Execute: tm.getLayout,
	}
}
