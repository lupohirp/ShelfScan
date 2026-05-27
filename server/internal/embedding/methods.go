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
	req, err := http.NewRequest("POST", e.getUrl()+"/embed", body)
	if err != nil {
		return nil, err
	}
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
	req, err := http.NewRequest("POST", e.getUrl()+"/embed", body)
	if err != nil {
		return nil, err
	}
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
	host := e.host
	if host == "" {
		host = "http://localhost"
	}
	if len(host) < 7 || (host[:7] != "http://" && (len(host) < 8 || host[:8] != "https://")) {
		host = "http://" + host
	}
	hasPort := false
	startIndex := 7
	if host[:8] == "https://" {
		startIndex = 8
	}
	for i := startIndex; i < len(host); i++ {
		if host[i] == ':' {
			hasPort = true
			break
		}
	}
	if hasPort || e.port == "" {
		return host
	}
	return host + ":" + e.port
}
