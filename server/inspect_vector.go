package main

import (
	"context"
	"fmt"
	"reflect"
	"github.com/qdrant/go-client/qdrant"
)

func main() {
	client, err := qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	id := uint64(2826376487158208505)
	resp, err := client.Get(context.Background(), &qdrant.GetPoints{
		CollectionName: "jewelry_inventory",
		Ids:            []*qdrant.PointId{qdrant.NewIDNum(id)},
		WithVectors:    qdrant.NewWithVectors(true),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		panic(err)
	}
	if len(resp) == 0 {
		fmt.Println("Point not found!")
		return
	}

	point := resp[0]
	v := point.Vectors
	if v == nil {
		fmt.Println("Vectors nil")
		return
	}

	fmt.Println("GetVector():", v.GetVector())
	fmt.Println("GetVectors():", v.GetVectors())

	if v.GetVectors() != nil {
		namedVectors := v.GetVectors()
		fmt.Println("Named vectors map:", namedVectors.Vectors)
		for k, val := range namedVectors.Vectors {
			fmt.Printf("Key: %q, Data length: %d\n", k, len(val.GetData()))
		}
	}
	
	// Let's print using reflect
	opt := v.GetVectorsOptions()
	fmt.Println("VectorsOptions reflect type:", reflect.TypeOf(opt))
}
