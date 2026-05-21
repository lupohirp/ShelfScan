package tools

import (
	"encoding/json"
)

func (tm *ToolManager) getLayout(arguments json.RawMessage) (interface{}, error) {
	// For now, hardcoded response as in the original main.go
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": "Expected items: [Diamond Ring A, Gold Necklace B, Emerald Studs C]"},
		},
	}, nil
}
