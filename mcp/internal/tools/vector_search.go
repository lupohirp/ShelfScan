package tools

import (
	"encoding/json"
)

func (tm *ToolManager) vectorSearch(arguments json.RawMessage) (interface{}, error) {
	var args struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, err
	}

	results, err := tm.qdrantClient.PerformVectorSearch(args.Embedding)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": results},
		},
	}, nil
}
