package qdrant

import "strconv"

type QdrantClient struct {
	host string
	port int
}

func NewQdrantClient() *QdrantClient {
	return &QdrantClient{}
}

func (q *QdrantClient) WithHost(host string) *QdrantClient {
	q.host = host
	return q
}

func (q *QdrantClient) WithPort(port string) *QdrantClient {
	if p, err := strconv.Atoi(port); err == nil {
		q.port = p
	}
	return q
}
