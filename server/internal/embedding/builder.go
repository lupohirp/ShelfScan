package embedding

type EmbeddingClient struct {
	host string
	port string
}

func NewEmbedding() *EmbeddingClient {
	return &EmbeddingClient{}
}

func (e *EmbeddingClient) WithHost(url string) *EmbeddingClient {
	e.host = url
	return e
}

func (e *EmbeddingClient) WithPort(port string) *EmbeddingClient {
	e.port = port
	return e
}
