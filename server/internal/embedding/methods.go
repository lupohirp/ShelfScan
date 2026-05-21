package embedding

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	domain "shelfscan-api/internal/domain/sub"
)

func (e *EmbeddingClient) GetEmbedding(file io.Reader, filename string) ([]float32, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, file)
	writer.Close()
	req, _ := http.NewRequest("POST", e.getUrl()+"/embed", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var res domain.EmbeddingResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Embedding, nil
}

func (e *EmbeddingClient) GetEmbeddingBytes(imgData []byte, filename string) ([]float32, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, bytes.NewReader(imgData))
	writer.Close()
	req, _ := http.NewRequest("POST", e.getUrl()+"/embed", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var res domain.EmbeddingResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Embedding, nil
}

func (e *EmbeddingClient) getUrl() string {
	return e.host + ":" + e.port

}
