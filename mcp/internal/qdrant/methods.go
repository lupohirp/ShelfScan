package qdrant

import (
	"context"
	"encoding/json"

	"github.com/qdrant/go-client/qdrant"
)

func (q *QdrantClient) PerformVectorSearch(vector []float32) (string, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: q.host,
		Port: q.port,
	})
	if err != nil {
		return "", err
	}
	defer client.Close()

	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "jewelry_inventory",
		Query:          qdrant.NewQueryDense(vector),
		Limit:          qdrant.PtrOf(uint64(5)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return "", err
	}

	var results []map[string]any
	for _, hit := range searchResult {
		payload := make(map[string]any)
		for k, v := range hit.Payload {
			payload[k] = v.GetStringValue()
		}
		results = append(results, map[string]any{
			"name":     payload["name"],
			"sku":      payload["sku"],
			"imageUrl": payload["imageUrl"],
			"color":    payload["color"],
			"material": payload["material"],
			"score":    hit.Score,
		})
	}

	output, _ := json.Marshal(results)
	return string(output), nil
}
