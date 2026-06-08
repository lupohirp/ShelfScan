package qdrant

import (
	"context"
	"fmt"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

func (q *QdrantClient) getClient() (*qdrant.Client, error) {
	host := q.host
	if len(host) > 7 && host[:7] == "http://" {
		host = host[7:]
	} else if len(host) > 8 && host[:8] == "https://" {
		host = host[8:]
	}

	port := q.port
	if port == 6333 {
		port = 6334
	}

	client, err := qdrant.NewClient(&qdrant.Config{Host: host, Port: port})
	if err != nil {
		client, err = qdrant.NewClient(&qdrant.Config{Host: "qdrant", Port: 6334})
		if err != nil {
			client, err = qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
		}
	}
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (q *QdrantClient) ListInventory() ([]map[string]any, error) {
	client, err := q.getClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

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

func (q *QdrantClient) SaveMultipleToQdrant(name string, sku string, imageUrl string, color string, material string, vectors [][]float32) error {
	client, err := q.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	var points []*qdrant.PointStruct
	for i, vector := range vectors {
		id := uint64(systemID(fmt.Sprintf("%s_%d", name, i)))
		payload := map[string]any{
			"name":     name,
			"sku":      sku,
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
	return q.performVectorSearchWithLimit(vector, 10)
}

func (q *QdrantClient) performVectorSearchWithLimit(vector []float32, limit uint64) ([]map[string]any, error) {
	client, err := q.getClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

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
		results = append(results, map[string]any{
			"name":     payload["name"],
			"sku":      payload["sku"],
			"imageUrl": payload["imageUrl"],
			"color":    payload["color"],
			"material": payload["material"],
			"score":    hit.Score,
		})
	}
	return results, nil
}

func (q *QdrantClient) UpdatePayload(idStr string, name string, sku string, color string, material string) error {
	client, err := q.getClient()
	if err != nil {
		return err
	}
	defer client.Close()
	var id uint64
	fmt.Sscanf(idStr, "%d", &id)

	resp, err := client.Get(context.Background(), &qdrant.GetPoints{
		CollectionName: "jewelry_inventory",
		Ids:            []*qdrant.PointId{qdrant.NewIDNum(id)},
		WithVectors:    qdrant.NewWithVectors(true),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch point: %w", err)
	}
	if len(resp) == 0 {
		return fmt.Errorf("point not found: %s", idStr)
	}

	point := resp[0]
	oldPayload := make(map[string]any)
	for k, v := range point.Payload {
		oldPayload[k] = v.GetStringValue()
	}
	oldName := oldPayload["name"].(string)

	newPayload := map[string]any{
		"name":     name,
		"sku":      sku,
		"color":    color,
		"material": material,
	}
	if img, ok := oldPayload["imageUrl"]; ok {
		newPayload["imageUrl"] = img
	}

	if oldName == name {
		_, err = client.SetPayload(context.Background(), &qdrant.SetPayloadPoints{
			CollectionName: "jewelry_inventory",
			PointsSelector: qdrant.NewPointsSelector(qdrant.NewIDNum(id)),
			Payload:        qdrant.NewValueMap(newPayload),
		})
		return err
	}

	_, err = client.Delete(context.Background(), &qdrant.DeletePoints{
		CollectionName: "jewelry_inventory",
		Points:         qdrant.NewPointsSelector(qdrant.NewIDNum(id)),
	})
	if err != nil {
		return err
	}

	newID := uint64(systemID(fmt.Sprintf("%s_%s", name, idStr)))

	_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: "jewelry_inventory",
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewIDNum(newID),
				Vectors: qdrant.NewVectorsDense(point.Vectors.GetVector().GetDense().GetData()),
				Payload: qdrant.NewValueMap(newPayload),
			},
		},
	})
	return err
}

func (q *QdrantClient) InitCollectionIfNeeded(size uint64) error {
	client, err := q.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := context.Background()
	info, err := client.GetCollectionInfo(ctx, "jewelry_inventory")
	if err == nil {
		var currentSize uint64
		if info.Config != nil && info.Config.Params != nil && info.Config.Params.VectorsConfig != nil {
			if params := info.Config.Params.VectorsConfig.GetParams(); params != nil {
				currentSize = params.GetSize()
			}
		}
		if currentSize == size {
			log.Printf("Collection jewelry_inventory already exists with correct vector size %d.", size)
			return nil
		}
		log.Printf("Collection jewelry_inventory has size %d, but expected %d. Recreating...", currentSize, size)
		err = client.DeleteCollection(ctx, "jewelry_inventory")
		if err != nil {
			return fmt.Errorf("failed to delete collection: %w", err)
		}
	} else {
		log.Printf("Collection jewelry_inventory does not exist. Creating...")
	}

	err = client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: "jewelry_inventory",
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     size,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	log.Printf("Collection jewelry_inventory created successfully with size %d and Cosine distance.", size)
	return nil
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
