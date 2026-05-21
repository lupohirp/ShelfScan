package qdrant

type QdrantClient struct {
	host string
	port int
}

func (q *QdrantClient) NewQdrantClient() *QdrantClient {
	return &QdrantClient{}
}

func (q *QdrantClient) WithHost(host string) *QdrantClient {
	q.host = host
	return q
}

func (q *QdrantClient) WithPort(port int) *QdrantClient {
	q.port = port
	return q
}