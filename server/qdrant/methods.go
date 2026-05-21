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
