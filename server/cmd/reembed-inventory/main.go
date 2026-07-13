// One-shot migration: re-embeds every point in the jewelry_inventory
// Qdrant collection against the embedding provider (Gemini), preserving
// point IDs and payloads. Used to re-align the collection after any
// change that invalidates the existing latent space.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"

	"shelfscan-api/internal/embedding"
)

const collection = "jewelry_inventory"

func main() {
	qdrantHost := flag.String("qdrant-host", "qdrant", "Qdrant gRPC host")
	qdrantPort := flag.Int("qdrant-port", 6334, "Qdrant gRPC port")
	uploadsDir := flag.String("uploads-dir", "/app/uploads", "Root directory holding the images referenced by imageUrl payloads")
	sampleN := flag.Int("self-check", 3, "How many random points to re-verify at the end")
	rate := flag.Duration("rate", 200*time.Millisecond, "Minimum interval between embedding calls (Gemini rate limit hygiene)")
	dryRun := flag.Bool("dry-run", false, "Scroll + fetch files but skip embed/upsert")
	flag.Parse()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" && !*dryRun {
		log.Fatalf("GEMINI_API_KEY not set")
	}
	embedder := embedding.NewEmbedding().WithApiKey(apiKey)

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

	var ok, skipped, failed int
	last := time.Now().Add(-*rate)
	for i, p := range points {
		fp := resolveFilePath(p.Payload["imageUrl"], *uploadsDir)
		if fp == "" {
			log.Printf("[%d/%d] point %d (sku=%s): missing imageUrl — skipped", i+1, len(points), p.ID, p.Payload["sku"])
			skipped++
			continue
		}
		data, err := os.ReadFile(fp)
		if err != nil {
			log.Printf("[%d/%d] point %d (sku=%s): cannot read %s: %v — skipped", i+1, len(points), p.ID, p.Payload["sku"], fp, err)
			skipped++
			continue
		}
		if *dryRun {
			log.Printf("[%d/%d] would embed point %d (sku=%s, %d bytes)", i+1, len(points), p.ID, p.Payload["sku"], len(data))
			ok++
			continue
		}
		if wait := time.Until(last.Add(*rate)); wait > 0 {
			time.Sleep(wait)
		}
		last = time.Now()
		vec, err := embedder.GetEmbeddingBytes(data, filepath.Base(fp))
		if err != nil {
			log.Printf("[%d/%d] point %d (sku=%s): embed failed: %v", i+1, len(points), p.ID, p.Payload["sku"], err)
			failed++
			continue
		}
		if err := upsert(ctx, client, p, vec); err != nil {
			log.Printf("[%d/%d] point %d (sku=%s): upsert failed: %v", i+1, len(points), p.ID, p.Payload["sku"], err)
			failed++
			continue
		}
		ok++
		log.Printf("[%d/%d] re-embedded point %d (sku=%s)", i+1, len(points), p.ID, p.Payload["sku"])
	}
	log.Printf("re-embed complete: ok=%d skipped=%d failed=%d", ok, skipped, failed)

	if *dryRun || ok == 0 {
		return
	}
	if err := selfCheck(ctx, client, embedder, *uploadsDir, points, *sampleN); err != nil {
		log.Fatalf("self-check FAIL: %v", err)
	}
	log.Printf("self-check PASS")
}

type point struct {
	ID      uint64
	Payload map[string]string
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

func upsert(ctx context.Context, client *qdrant.Client, p *point, vec []float32) error {
	payload := map[string]any{}
	for k, v := range p.Payload {
		payload[k] = v
	}
	_, err := client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collection,
		Points: []*qdrant.PointStruct{{
			Id:      qdrant.NewIDNum(p.ID),
			Vectors: qdrant.NewVectorsDense(vec),
			Payload: qdrant.NewValueMap(payload),
		}},
	})
	return err
}

func selfCheck(ctx context.Context, client *qdrant.Client, embedder *embedding.EmbeddingClient, uploadsDir string, all []*point, n int) error {
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
	rng := rand.New(rand.NewSource(1))
	rng.Shuffle(len(valid), func(i, j int) { valid[i], valid[j] = valid[j], valid[i] })

	for i := 0; i < n; i++ {
		p := valid[i]
		fp := resolveFilePath(p.Payload["imageUrl"], uploadsDir)
		data, err := os.ReadFile(fp)
		if err != nil {
			return fmt.Errorf("point %d: cannot re-read %s: %w", p.ID, fp, err)
		}
		vec, err := embedder.GetEmbeddingBytes(data, filepath.Base(fp))
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
		rank := -1
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
