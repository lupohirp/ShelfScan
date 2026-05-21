package qdrant

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

func (q *QdrantClient) getClient() (*qdrant.Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{Host: q.host, Port: q.port})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	}
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client, nil

}

func (q *QdrantClient) ListInventory() ([]map[string]any, error) {
	client, err := q.getClient()

	points, err := client.Scroll(context.Background(), &qdrant.ScrollPoints{
		CollectionName: "jewelry_inventory",
		Limit:          qdrant.PtrOf(uint32(100)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	var items []map[string]any
	for _, p := range points {
		payload := make(map[string]any)
		for k, v := range p.Payload {
			payload[k] = v.GetStringValue()
		}
		items = append(items, map[string]any{"id": fmt.Sprintf("%d", p.Id.GetNum()), "payload": payload})
	}
	return items, nil
}

func (q *QdrantClient) DeleteFromQdrant(idStr string) error {
	client, err := q.getClient()
	if err != nil {
		return err
	}
	defer client.Close()
	var id uint64
	fmt.Sscanf(idStr, "%d", &id)
	_, err = client.Delete(context.Background(), &qdrant.DeletePoints{
		CollectionName: "jewelry_inventory",
		Points:         qdrant.NewPointsSelector(qdrant.NewIDNum(id)),
	})
	return err
}

func (q *QdrantClient) SaveMultipleToQdrant(name string, imageUrl string, color string, material string, vectors [][]float32) error {
	client, err := q.getClient()
	var points []*qdrant.PointStruct
	for i, vector := range vectors {
		id := uint64(systemID(fmt.Sprintf("%s_%d", name, i)))
		payload := map[string]any{
			"name":     name,
			"color":    color,
			"material": material,
		}
		if imageUrl != "" {
			payload["imageUrl"] = imageUrl
		}
		points = append(points, &qdrant.PointStruct{Id: qdrant.NewIDNum(id), Vectors: qdrant.NewVectorsDense(vector), Payload: qdrant.NewValueMap(payload)})
	}
	_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{CollectionName: "jewelry_inventory", Points: points})
	return err
}

func (q *QdrantClient) PerformVectorSearch(vector []float32) ([]map[string]any, error) {
	return q.performVectorSearchWithLimit(vector, 3)
}

func (q *QdrantClient) performVectorSearchWithLimit(vector []float32, limit uint64) ([]map[string]any, error) {
	client, err := q.getClient()

	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "jewelry_inventory",
		Query:          qdrant.NewQueryDense(vector),
		Limit:          qdrant.PtrOf(limit),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}
	var results []map[string]any
	for _, hit := range searchResult {
		payload := make(map[string]any)
		for k, v := range hit.Payload {
			payload[k] = v.GetStringValue()
		}
		results = append(results, map[string]any{"name": payload["name"], "imageUrl": payload["imageUrl"], "score": hit.Score})
	}
	return results, nil
}

func systemID(name string) int {
	h := 0
	for _, c := range name {
		h = 31*h + int(c)
	}
	if h < 0 {
		return -h
	}
	return h
}
