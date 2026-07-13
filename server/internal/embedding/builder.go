package embedding

type EmbeddingClient struct {
	apiKey string
}

func NewEmbedding() *EmbeddingClient {
	return &EmbeddingClient{}
}

func (e *EmbeddingClient) WithApiKey(key string) *EmbeddingClient {
	e.apiKey = key
	return e
}
