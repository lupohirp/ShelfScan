// One-shot migration: re-embeds every point in the jewelry_inventory
// Qdrant collection against the SigLIP2 sidecar, preserving point IDs
// and payloads. Used to cut over from the Gemini embedding latent space
// to the SigLIP2 latent space without losing catalog state.
//
// Usage inside the VPS docker network:
//
//	reembed-inventory \
//	    --qdrant-host qdrant --qdrant-port 6334 \
//	    --embed-url http://embeddings:8001 \
//	    --uploads-dir /app/uploads
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

const collection = "jewelry_inventory"

func main() {
	qdrantHost := flag.String("qdrant-host", "qdrant", "Qdrant gRPC host")
	qdrantPort := flag.Int("qdrant-port", 6334, "Qdrant gRPC port")
	embedURL := flag.String("embed-url", "http://embeddings:8001", "SigLIP2 sidecar base URL")
	uploadsDir := flag.String("uploads-dir", "/app/uploads", "Root directory holding the images referenced by imageUrl payloads")
	batchSize := flag.Int("batch", 8, "Images per /embed_batch request")
	sampleN := flag.Int("self-check", 3, "How many random points to re-verify at the end")
	dryRun := flag.Bool("dry-run", false, "Scroll + fetch files but skip embed/upsert")
	flag.Parse()

	client, err := qdrant.NewClient(&qdrant.Config{Host: *qdrantHost, Port: *qdrantPort})
	if err != nil {
		log.Fatalf("connect qdrant: %v", err)
	}
	defer client.Close()
	ctx := context.Background()

	points, err := scrollAll(ctx, client)
	if err != nil {
		log.Fatalf("scroll: %v", err)
	}
	log.Printf("scrolled %d points from %s", len(points), collection)

	hc := &http.Client{Timeout: 120 * time.Second}
	batches := groupInto(points, *batchSize)
	var ok, skipped, failed int
	for bi, batch := range batches {
		files, kept, drop := loadFiles(batch, *uploadsDir)
		skipped += drop
		if len(kept) == 0 {
			continue
		}
		if *dryRun {
			log.Printf("[batch %d/%d] would embed %d image(s)", bi+1, len(batches), len(kept))
			ok += len(kept)
			continue
		}
		vectors, err := embedBatch(hc, *embedURL, files, kept)
		if err != nil {
			log.Printf("[batch %d/%d] embed failed: %v — points skipped", bi+1, len(batches), err)
			failed += len(kept)
			continue
		}
		if err := upsert(ctx, client, kept, vectors); err != nil {
			log.Printf("[batch %d/%d] upsert failed: %v", bi+1, len(batches), err)
			failed += len(kept)
			continue
		}
		ok += len(kept)
		log.Printf("[batch %d/%d] re-embedded %d point(s)", bi+1, len(batches), len(kept))
	}
	log.Printf("re-embed complete: ok=%d skipped=%d failed=%d", ok, skipped, failed)

	if *dryRun || ok == 0 {
		return
	}
	if err := selfCheck(ctx, client, hc, *embedURL, *uploadsDir, points, *sampleN); err != nil {
		log.Fatalf("self-check FAIL: %v", err)
	}
	log.Printf("self-check PASS")
}

type point struct {
	ID      uint64
	Payload map[string]string
	// filled by loadFiles
	imagePath string
}

func scrollAll(ctx context.Context, client *qdrant.Client) ([]*point, error) {
	var out []*point
	var offset *qdrant.PointId
	for {
		batch, next, err := client.ScrollAndOffset(ctx, &qdrant.ScrollPoints{
			CollectionName: collection,
			Limit:          qdrant.PtrOf(uint32(200)),
			WithPayload:    qdrant.NewWithPayload(true),
			Offset:         offset,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range batch {
			payload := map[string]string{}
			for k, v := range p.Payload {
				payload[k] = v.GetStringValue()
			}
			out = append(out, &point{ID: p.Id.GetNum(), Payload: payload})
		}
		if next == nil || len(batch) == 0 {
			break
		}
		offset = next
	}
	return out, nil
}

func groupInto[T any](in []T, size int) [][]T {
	if size < 1 {
		size = 1
	}
	var out [][]T
	for i := 0; i < len(in); i += size {
		end := i + size
		if end > len(in) {
			end = len(in)
		}
		out = append(out, in[i:end])
	}
	return out
}

func resolveFilePath(imageURL, uploadsDir string) string {
	if imageURL == "" {
		return ""
	}
	trimmed := imageURL
	if i := strings.Index(trimmed, "/uploads/"); i >= 0 {
		trimmed = trimmed[i+len("/uploads/"):]
	} else if strings.HasPrefix(trimmed, "uploads/") {
		trimmed = trimmed[len("uploads/"):]
	} else {
		trimmed = path.Base(trimmed)
	}
	return filepath.Join(uploadsDir, trimmed)
}

func loadFiles(batch []*point, uploadsDir string) (bodies [][]byte, kept []*point, dropped int) {
	for _, p := range batch {
		fp := resolveFilePath(p.Payload["imageUrl"], uploadsDir)
		if fp == "" {
			log.Printf("point %d (sku=%s): missing imageUrl — skipped", p.ID, p.Payload["sku"])
			dropped++
			continue
		}
		data, err := os.ReadFile(fp)
		if err != nil {
			log.Printf("point %d (sku=%s): cannot read %s: %v — skipped", p.ID, p.Payload["sku"], fp, err)
			dropped++
			continue
		}
		p.imagePath = fp
		bodies = append(bodies, data)
		kept = append(kept, p)
	}
	return bodies, kept, dropped
}

type batchResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func embedBatch(hc *http.Client, base string, bodies [][]byte, kept []*point) ([][]float32, error) {
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	for i, body := range bodies {
		filename := filepath.Base(kept[i].imagePath)
		part, err := mw.CreateFormFile("image", filename)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(body); err != nil {
			return nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(base, "/")+"/embed_batch", buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	var parsed batchResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if len(parsed.Embeddings) != len(bodies) {
		return nil, fmt.Errorf("sidecar returned %d vectors for %d images", len(parsed.Embeddings), len(bodies))
	}
	return parsed.Embeddings, nil
}

func upsert(ctx context.Context, client *qdrant.Client, kept []*point, vectors [][]float32) error {
	pts := make([]*qdrant.PointStruct, 0, len(kept))
	for i, p := range kept {
		payload := map[string]any{}
		for k, v := range p.Payload {
			payload[k] = v
		}
		pts = append(pts, &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(p.ID),
			Vectors: qdrant.NewVectorsDense(vectors[i]),
			Payload: qdrant.NewValueMap(payload),
		})
	}
	_, err := client.Upsert(ctx, &qdrant.UpsertPoints{CollectionName: collection, Points: pts})
	return err
}

func embedOne(hc *http.Client, base string, body []byte) ([]float32, error) {
	resp, err := hc.Post(strings.TrimRight(base, "/")+"/embed", "application/octet-stream", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	var parsed struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	return parsed.Embedding, nil
}

func selfCheck(ctx context.Context, client *qdrant.Client, hc *http.Client, embedBase, uploadsDir string, all []*point, n int) error {
	valid := make([]*point, 0, len(all))
	for _, p := range all {
		if resolveFilePath(p.Payload["imageUrl"], uploadsDir) != "" {
			valid = append(valid, p)
		}
	}
	if len(valid) == 0 {
		return fmt.Errorf("no points have a resolvable imageUrl")
	}
	if n > len(valid) {
		n = len(valid)
	}
	rng := rand.New(rand.NewSource(1)) // deterministic order for reproducible logs
	rng.Shuffle(len(valid), func(i, j int) { valid[i], valid[j] = valid[j], valid[i] })

	for i := 0; i < n; i++ {
		p := valid[i]
		fp := resolveFilePath(p.Payload["imageUrl"], uploadsDir)
		data, err := os.ReadFile(fp)
		if err != nil {
			return fmt.Errorf("point %d: cannot re-read %s: %w", p.ID, fp, err)
		}
		vec, err := embedOne(hc, embedBase, data)
		if err != nil {
			return fmt.Errorf("point %d: re-embed failed: %w", p.ID, err)
		}
		hits, err := client.Query(ctx, &qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryDense(vec),
			Limit:          qdrant.PtrOf(uint64(5)),
			WithPayload:    qdrant.NewWithPayload(false),
		})
		if err != nil {
			return fmt.Errorf("point %d: qdrant query failed: %w", p.ID, err)
		}
		var rank int = -1
		for idx, h := range hits {
			if h.Id.GetNum() == p.ID {
				rank = idx + 1
				break
			}
		}
		if rank == -1 {
			return fmt.Errorf("point %d (sku=%s): not in top-5 after re-embed", p.ID, p.Payload["sku"])
		}
		log.Printf("self-check %d/%d: point %d (sku=%s) recovered at rank %d", i+1, n, p.ID, p.Payload["sku"], rank)
	}
	return nil
}
