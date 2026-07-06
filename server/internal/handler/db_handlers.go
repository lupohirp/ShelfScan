package handler

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"shelfscan-api/internal/db"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type VisitPayload struct {
	ID              string `json:"id"`
	Store           struct {
		ID string `json:"id"`
	} `json:"store"`
	Status          string `json:"status"`
	Coverage        int    `json:"coverage"`
	CreatedAt       string `json:"createdAt"`
	FinalizedAt     string `json:"finalizedAt"`
	Agent           string `json:"agent"`
	FoundProducts   []struct {
		SKU      string `json:"sku"`
		Name     string `json:"name"`
		Category string `json:"category"`
		ImageURL string `json:"imageUrl"`
		CropURL  string `json:"cropUrl"`
	} `json:"foundProducts"`
	MissingProducts []struct {
		SKU      string `json:"sku"`
		Name     string `json:"name"`
		Category string `json:"category"`
		ImageURL string `json:"imageUrl"`
		CropURL  string `json:"cropUrl"`
	} `json:"missingProducts"`
	AnalyzedImages []struct {
		CapturedImage string `json:"capturedImage"` // base64 data URL or upload path
	} `json:"analyzedImages"`
}

func (h *Handler) StoresHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			type StorePayload struct {
				Code         string `json:"code"`
				Name         string `json:"name"`
				Province     string `json:"province"`
				ProvinceName string `json:"province_name"`
				Address      string `json:"address"`
				Region       string `json:"region"`
				RegionCode   string `json:"region_code"`
				City         string `json:"city"`
				AgentName    string `json:"agent_name"`
			}

			var p StorePayload
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if p.Code == "" || p.Name == "" || p.Province == "" || p.ProvinceName == "" || p.City == "" {
				http.Error(w, "Missing required fields (code, name, province, province_name, city)", http.StatusBadRequest)
				return
			}

			if p.Region == "" {
				p.Region = "Lombardia"
			}
			if p.RegionCode == "" {
				p.RegionCode = "LOM"
			}
			if p.AgentName == "" {
				p.AgentName = "Marco Ferraro"
			}

			cleanedName := db.CleanStoreName(p.Name)
			cleanedProvince := strings.ToUpper(strings.TrimSpace(p.Province))
			cleanedProvinceName := strings.ToUpper(strings.TrimSpace(p.ProvinceName))
			cleanedAddress := strings.ToUpper(strings.TrimSpace(p.Address))
			cleanedRegion := strings.ToUpper(strings.TrimSpace(p.Region))
			cleanedRegionCode := strings.ToUpper(strings.TrimSpace(p.RegionCode))
			cleanedCity := strings.ToUpper(strings.TrimSpace(p.City))
			cleanedAgentName := db.GetOfficialAgentName(p.AgentName)

			res, err := h.db.Exec(`
				INSERT INTO stores (code, name, province, province_name, address, region, region_code, city, agent_name)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, p.Code, cleanedName, cleanedProvince, cleanedProvinceName, cleanedAddress, cleanedRegion, cleanedRegionCode, cleanedCity, cleanedAgentName)
			if err != nil {
				log.Printf("Error inserting store: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			lastID, _ := res.LastInsertId()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id":            strconv.FormatInt(lastID, 10),
				"code":          p.Code,
				"name":          cleanedName,
				"province":      cleanedProvince,
				"province_name": cleanedProvinceName,
				"address":       cleanedAddress,
				"region":        cleanedRegion,
				"region_code":   cleanedRegionCode,
				"city":          cleanedCity,
				"agent_name":    cleanedAgentName,
				"status":        "created",
			})
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query().Get("q")
		region := r.URL.Query().Get("region")
		province := r.URL.Query().Get("province")
		city := r.URL.Query().Get("city")
		agent := r.URL.Query().Get("agent")

		queryStr := "SELECT id, code, name, province, province_name, address, region, region_code, city, agent_name FROM stores WHERE 1=1"
		var args []any

		if q != "" {
			queryStr += " AND (name LIKE ? OR code LIKE ? OR city LIKE ? OR province LIKE ? OR province_name LIKE ? OR address LIKE ? OR region LIKE ? OR region_code LIKE ?)"
			likeQ := "%" + q + "%"
			args = append(args, likeQ, likeQ, likeQ, likeQ, likeQ, likeQ, likeQ, likeQ)
		}
		if region != "" {
			queryStr += " AND (region = ? OR region_code = ?)"
			args = append(args, region, region)
		}
		if province != "" {
			if len(province) == 2 {
				queryStr += " AND province = ?"
				args = append(args, strings.ToUpper(province))
			} else {
				queryStr += " AND province_name LIKE ?"
				args = append(args, "%"+province+"%")
			}
		}
		if city != "" {
			queryStr += " AND city LIKE ?"
			args = append(args, "%"+city+"%")
		}
		if agent != "" {
			queryStr += " AND agent_name = ?"
			args = append(args, agent)
		}

		queryStr += " ORDER BY name ASC LIMIT 100"

		rows, err := h.db.Query(queryStr, args...)
		if err != nil {
			log.Printf("Error querying stores: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type StoreJSON struct {
			ID           string `json:"id"`
			Code         string `json:"code"`
			Name         string `json:"name"`
			Province     string `json:"province"`
			ProvinceName string `json:"province_name"`
			Address      string `json:"address"`
			Region       string `json:"region"`
			RegionCode   string `json:"region_code"`
			City         string `json:"city"`
			AgentName    string `json:"agent_name"`
		}

		stores := []StoreJSON{}
		for rows.Next() {
			var s StoreJSON
			var id int
			err := rows.Scan(&id, &s.Code, &s.Name, &s.Province, &s.ProvinceName, &s.Address, &s.Region, &s.RegionCode, &s.City, &s.AgentName)
			if err != nil {
				log.Printf("Error scanning store: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.ID = strconv.Itoa(id)
			stores = append(stores, s)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stores)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) StoresDetailHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 3 || parts[2] == "" {
				http.Error(w, "Invalid store ID", http.StatusBadRequest)
				return
			}
			storeIDStr := parts[2]
			storeID, err := strconv.Atoi(storeIDStr)
			if err != nil {
				http.Error(w, "Invalid store ID", http.StatusBadRequest)
				return
			}

			tx, err := h.db.Begin()
			if err != nil {
				log.Printf("Error starting delete transaction: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()

			_, err = tx.Exec(`
				DELETE FROM visit_products 
				WHERE visit_id IN (SELECT id FROM visits WHERE store_id = ?)
			`, storeID)
			if err != nil {
				log.Printf("Error deleting store visit products: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(`
				DELETE FROM visit_scans 
				WHERE visit_id IN (SELECT id FROM visits WHERE store_id = ?)
			`, storeID)
			if err != nil {
				log.Printf("Error deleting store visit scans: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec("DELETE FROM visits WHERE store_id = ?", storeID)
			if err != nil {
				log.Printf("Error deleting store visits: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			res, err := tx.Exec("DELETE FROM stores WHERE id = ?", storeID)
			if err != nil {
				log.Printf("Error deleting store: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rowsAffected, _ := res.RowsAffected()
			if rowsAffected == 0 {
				http.Error(w, "Store not found", http.StatusNotFound)
				return
			}

			if err := tx.Commit(); err != nil {
				log.Printf("Error committing delete transaction: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract ID from URL path (e.g. /stores/123)
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 || parts[2] == "" {
			http.Error(w, "Invalid store ID", http.StatusBadRequest)
			return
		}
		storeIDStr := parts[2]
		storeID, err := strconv.Atoi(storeIDStr)
		if err != nil {
			http.Error(w, "Invalid store ID", http.StatusBadRequest)
			return
		}

		// Query store
		var s struct {
			ID         string `json:"id"`
			Code       string `json:"code"`
			Name       string `json:"name"`
			Province   string `json:"province"`
			Address    string `json:"address"`
			Region     string `json:"region"`
			RegionCode string `json:"region_code"`
			City       string `json:"city"`
			AgentName  string `json:"agent_name"`
		}
		var id int
		err = h.db.QueryRow("SELECT id, code, name, province, address, region, region_code, city, agent_name FROM stores WHERE id = ?", storeID).
			Scan(&id, &s.Code, &s.Name, &s.Province, &s.Address, &s.Region, &s.RegionCode, &s.City, &s.AgentName)
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Error querying store details: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.ID = strconv.Itoa(id)

		// Query visits history
		rows, err := h.db.Query(`
			SELECT id, status, coverage, missing_count, agent, created_at, finalized_at
			FROM visits
			WHERE store_id = ?
			ORDER BY created_at DESC
		`, storeID)
		if err != nil {
			log.Printf("Error querying store visits: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type VisitJSON struct {
			ID          string     `json:"id"`
			Status      string     `json:"status"`
			Coverage    int        `json:"coverage"`
			Missing     int        `json:"missing_count"`
			Agent       string     `json:"agent"`
			CreatedAt   time.Time  `json:"created_at"`
			FinalizedAt *time.Time `json:"finalized_at,omitempty"`
			Photos      []any      `json:"photos"`
			Exposed     []any      `json:"exposed"`
			MissingProds []any     `json:"missing"`
		}

		visits := []VisitJSON{}
		visitIDs := []string{}
		for rows.Next() {
			var v VisitJSON
			err := rows.Scan(&v.ID, &v.Status, &v.Coverage, &v.Missing, &v.Agent, &v.CreatedAt, &v.FinalizedAt)
			if err != nil {
				log.Printf("Error scanning visit: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			v.Photos = []any{}
			v.Exposed = []any{}
			v.MissingProds = []any{}
			visits = append(visits, v)
			visitIDs = append(visitIDs, v.ID)
		}

		// Map to store photos by visit_id
		photosMap := make(map[string][]any)
		if len(visitIDs) > 0 {
			photoRows, err := h.db.Query(`
				SELECT s.visit_id, s.photo_url, s.photo_tone
				FROM visit_scans s
				JOIN visits v ON s.visit_id = v.id
				WHERE v.store_id = ?
			`, storeID)
			if err == nil {
				defer photoRows.Close()
				for photoRows.Next() {
					var visitID, url, tone string
					if err := photoRows.Scan(&visitID, &url, &tone); err == nil {
						photosMap[visitID] = append(photosMap[visitID], map[string]string{
							"url":  url,
							"tone": tone,
						})
					}
				}
			}
		}

		// Map to store exposed/missing products by visit_id
		exposedMap := make(map[string][]any)
		missingMap := make(map[string][]any)
		if len(visitIDs) > 0 {
			prodRows, err := h.db.Query(`
				SELECT p.visit_id, p.sku, p.name, p.category, COALESCE(p.image_url, ''), COALESCE(p.crop_url, ''), p.is_exposed
				FROM visit_products p
				JOIN visits v ON p.visit_id = v.id
				WHERE v.store_id = ?
			`, storeID)
			if err == nil {
				defer prodRows.Close()
				for prodRows.Next() {
					var visitID, sku, name, cat, imgUrl, cropUrl string
					var isExposed bool
					if err := prodRows.Scan(&visitID, &sku, &name, &cat, &imgUrl, &cropUrl, &isExposed); err == nil {
						prodItem := map[string]string{
							"sku":       sku,
							"name":      name,
							"cat":       cat,
							"image_url": imgUrl,
							"crop_url":  cropUrl,
						}
						if isExposed {
							exposedMap[visitID] = append(exposedMap[visitID], prodItem)
						} else {
							missingMap[visitID] = append(missingMap[visitID], prodItem)
						}
					}
				}
			}
		}

		// Merge database results in memory
		for i := range visits {
			vID := visits[i].ID
			if photos, ok := photosMap[vID]; ok {
				visits[i].Photos = photos
			}
			if exposed, ok := exposedMap[vID]; ok {
				visits[i].Exposed = exposed
			}
			if missing, ok := missingMap[vID]; ok {
				visits[i].MissingProds = missing
			}
		}

		response := map[string]any{
			"store":  s,
			"visits": visits,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) VisitsHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload VisitPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			log.Printf("Error decoding visit payload: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		storeID, err := strconv.Atoi(payload.Store.ID)
		if err != nil {
			log.Printf("Invalid store ID: %v", err)
			http.Error(w, "Invalid store ID", http.StatusBadRequest)
			return
		}

		// Parse times
		createdAt, err := time.Parse(time.RFC3339, payload.CreatedAt)
		if err != nil {
			createdAt = time.Now()
		}
		var finalizedAt *time.Time
		if payload.FinalizedAt != "" {
			t, err := time.Parse(time.RFC3339, payload.FinalizedAt)
			if err == nil {
				finalizedAt = &t
			}
		}
		if finalizedAt == nil && payload.Status == "finalized" {
			t := time.Now()
			finalizedAt = &t
		}

		agentName := payload.Agent
		if agentName == "" {
			err = h.db.QueryRow("SELECT agent_name FROM stores WHERE id = ?", storeID).Scan(&agentName)
			if err != nil {
				agentName = "Unknown Agent"
			}
		}

		tx, err := h.db.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 1. Insert or Replace visit
		missingCount := len(payload.MissingProducts)
		_, err = tx.Exec(`
			INSERT OR REPLACE INTO visits (id, store_id, status, coverage, missing_count, agent, created_at, finalized_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, payload.ID, storeID, payload.Status, payload.Coverage, missingCount, agentName, createdAt, finalizedAt)
		if err != nil {
			log.Printf("Error inserting visit: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 2. Clear existing scans and products for this visit
		_, err = tx.Exec("DELETE FROM visit_scans WHERE visit_id = ?", payload.ID)
		if err != nil {
			log.Printf("Error deleting existing scans: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec("DELETE FROM visit_products WHERE visit_id = ?", payload.ID)
		if err != nil {
			log.Printf("Error deleting existing products: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 3. Process and Insert scans
		_ = os.MkdirAll("uploads", 0755)
		for idx, img := range payload.AnalyzedImages {
			photoURL := ""
			if strings.HasPrefix(img.CapturedImage, "data:image/") {
				// Base64 Data URL, decode and save to file
				commaIdx := strings.Index(img.CapturedImage, ",")
				if commaIdx != -1 {
					b64Data := img.CapturedImage[commaIdx+1:]
					dec, err := base64.StdEncoding.DecodeString(b64Data)
					if err == nil {
						safeID := strings.NewReplacer("..", "", "/", "", "\\", "").Replace(payload.ID)
						filename := fmt.Sprintf("visit_%s_%d.jpg", safeID, idx)
						filepath := "uploads/" + filename
						err = os.WriteFile(filepath, dec, 0644)
						if err == nil {
							photoURL = "/uploads/" + filename
						} else {
							log.Printf("Error saving scan image: %v", err)
						}
					} else {
						log.Printf("Error decoding base64 image: %v", err)
					}
				}
			} else {
				photoURL = img.CapturedImage
			}

			if photoURL != "" {
				scanID := fmt.Sprintf("%s_%d", payload.ID, idx)
				_, err = tx.Exec(`
					INSERT INTO visit_scans (id, visit_id, photo_url, photo_tone)
					VALUES (?, ?, ?, ?)
				`, scanID, payload.ID, photoURL, "neutral")
				if err != nil {
					log.Printf("Error inserting visit scan: %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// 4. Insert found products
		for _, prod := range payload.FoundProducts {
			_, err = tx.Exec(`
				INSERT INTO visit_products (visit_id, sku, name, category, image_url, crop_url, is_exposed)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, payload.ID, prod.SKU, prod.Name, prod.Category, prod.ImageURL, prod.CropURL, true)
			if err != nil {
				log.Printf("Error inserting found product: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// 5. Insert missing products
		for _, prod := range payload.MissingProducts {
			_, err = tx.Exec(`
				INSERT INTO visit_products (visit_id, sku, name, category, image_url, crop_url, is_exposed)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, payload.ID, prod.SKU, prod.Name, prod.Category, prod.ImageURL, prod.CropURL, false)
			if err != nil {
				log.Printf("Error inserting missing product: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("Error committing transaction: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "success", "visit_id": payload.ID})
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}


func (h *Handler) StatsRegionsHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := h.db.Query(`
			SELECT
				s.region,
				s.region_code,
				COUNT(DISTINCT s.id) AS total_stores,
				COUNT(DISTINCT v.id) AS visit_count,
				COALESCE(AVG(v.coverage), 0.0) AS avg_coverage
			FROM stores s
			LEFT JOIN visits v ON s.id = v.store_id AND v.status = 'finalized'
			GROUP BY s.region, s.region_code
			ORDER BY avg_coverage DESC
		`)
		if err != nil {
			log.Printf("Error querying region stats: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type RegionStat struct {
			Region      string  `json:"region"`
			RegionCode  string  `json:"region_code"`
			TotalStores int     `json:"total_stores"`
			VisitCount  int     `json:"visit_count"`
			AvgCoverage float64 `json:"avg_coverage"`
		}

		stats := []RegionStat{}
		for rows.Next() {
			var rs RegionStat
			err := rows.Scan(&rs.Region, &rs.RegionCode, &rs.TotalStores, &rs.VisitCount, &rs.AvgCoverage)
			if err != nil {
				log.Printf("Error scanning region stat: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			stats = append(stats, rs)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) StatsTopProductsHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := h.db.Query(`
			SELECT
				sku,
				name,
				SUM(CASE WHEN is_exposed = 1 THEN 1 ELSE 0 END) AS exposed_count,
				COUNT(DISTINCT visit_id) AS total_visits,
				(CAST(SUM(CASE WHEN is_exposed = 1 THEN 1 ELSE 0 END) AS REAL) / COUNT(DISTINCT visit_id)) * 100 AS presence_rate
			FROM visit_products vp
			JOIN visits v ON vp.visit_id = v.id
			WHERE v.status = 'finalized'
			GROUP BY sku, name
			ORDER BY presence_rate DESC, exposed_count DESC
			LIMIT 10
		`)
		if err != nil {
			log.Printf("Error querying top products: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type TopProductStat struct {
			SKU          string  `json:"sku"`
			Name         string  `json:"name"`
			ExposedCount int     `json:"exposed_count"`
			TotalVisits  int     `json:"total_visits"`
			PresenceRate float64 `json:"presence_rate"`
		}

		stats := []TopProductStat{}
		for rows.Next() {
			var tp TopProductStat
			err := rows.Scan(&tp.SKU, &tp.Name, &tp.ExposedCount, &tp.TotalVisits, &tp.PresenceRate)
			if err != nil {
				log.Printf("Error scanning top product: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			stats = append(stats, tp)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) StatsRecentVisitsHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := h.db.Query(`
			SELECT
				v.id,
				s.name AS store_name,
				s.city,
				s.region,
				v.coverage,
				v.finalized_at,
				v.agent
			FROM visits v
			JOIN stores s ON v.store_id = s.id
			WHERE v.status = 'finalized'
			ORDER BY v.finalized_at DESC
			LIMIT 10
		`)
		if err != nil {
			log.Printf("Error querying recent visits: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type RecentVisit struct {
			ID          string    `json:"id"`
			StoreName   string    `json:"store_name"`
			City        string    `json:"city"`
			Region      string    `json:"region"`
			Coverage    int       `json:"coverage"`
			FinalizedAt time.Time `json:"finalized_at"`
			Agent       string    `json:"agent"`
		}

		visits := []RecentVisit{}
		for rows.Next() {
			var rv RecentVisit
			err := rows.Scan(&rv.ID, &rv.StoreName, &rv.City, &rv.Region, &rv.Coverage, &rv.FinalizedAt, &rv.Agent)
			if err != nil {
				log.Printf("Error scanning recent visit: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			visits = append(visits, rv)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(visits)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) StatsStoresHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := h.db.Query(`
			SELECT
				s.id,
				s.code,
				s.name,
				s.province,
				s.province_name,
				s.address,
				s.region,
				s.region_code,
				s.city,
				s.agent_name,
				v.id,
				v.coverage,
				v.missing_count,
				v.finalized_at
			FROM stores s
			LEFT JOIN (
				SELECT store_id, id, coverage, missing_count, finalized_at,
				       ROW_NUMBER() OVER (PARTITION BY store_id ORDER BY finalized_at DESC) as rn
				FROM visits
				WHERE status = 'finalized'
			) v ON s.id = v.store_id AND v.rn = 1
			ORDER BY s.name ASC
		`)
		if err != nil {
			log.Printf("Error querying stores stats: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type StoreStat struct {
			ID           string  `json:"id"`
			Code         string  `json:"code"`
			Name         string  `json:"name"`
			Province     string  `json:"province"`
			ProvinceName string  `json:"province_name"`
			Address      string  `json:"address"`
			Region       string  `json:"region"`
			RegionCode   string  `json:"region_code"`
			City         string  `json:"city"`
			AgentName    string  `json:"agent_name"`
			VisitID      *string `json:"visit_id,omitempty"`
			Coverage     *int    `json:"coverage,omitempty"`
			MissingCount *int    `json:"missing_count,omitempty"`
			FinalizedAt  *string `json:"finalized_at,omitempty"`
		}

		stores := []StoreStat{}
		for rows.Next() {
			var s StoreStat
			var id int
			var visitID sql.NullString
			var coverage sql.NullInt64
			var missingCount sql.NullInt64
			var finalizedAt sql.NullString

			err := rows.Scan(&id, &s.Code, &s.Name, &s.Province, &s.ProvinceName, &s.Address, &s.Region, &s.RegionCode, &s.City, &s.AgentName,
				&visitID, &coverage, &missingCount, &finalizedAt)
			if err != nil {
				log.Printf("Error scanning store stat: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.ID = strconv.Itoa(id)

			if visitID.Valid {
				s.VisitID = &visitID.String
			}
			if coverage.Valid {
				cov := int(coverage.Int64)
				s.Coverage = &cov
			}
			if missingCount.Valid {
				mc := int(missingCount.Int64)
				s.MissingCount = &mc
			}
			if finalizedAt.Valid {
				s.FinalizedAt = &finalizedAt.String
			}

			stores = append(stores, s)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stores)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) AgentsHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			type AgentPayload struct {
				Zona          string `json:"zona"`
				Agente        string `json:"agente"`
				Note          string `json:"note"`
				Tel           string `json:"tel"`
				Email         string `json:"email"`
				EmailPersonal string `json:"email_personal"`
				Password      string `json:"password"`
			}

			var p AgentPayload
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if p.Agente == "" || p.Zona == "" || p.Password == "" {
				http.Error(w, "Missing required fields (agente, zona, password)", http.StatusBadRequest)
				return
			}

			cleanedZona := strings.ToUpper(strings.TrimSpace(p.Zona))
			cleanedAgente := strings.ToUpper(strings.TrimSpace(p.Agente))

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Error hashing password: %v", err)
				http.Error(w, "Error processing password", http.StatusInternalServerError)
				return
			}

			res, err := h.db.Exec(`
				INSERT INTO agents (zona, agente, note, tel, email, email_personal, password)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, cleanedZona, cleanedAgente, p.Note, p.Tel, p.Email, p.EmailPersonal, string(hashedPassword))
			if err != nil {
				log.Printf("Error inserting agent: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			lastID, _ := res.LastInsertId()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id":             strconv.FormatInt(lastID, 10),
				"zona":           cleanedZona,
				"agente":         cleanedAgente,
				"note":           p.Note,
				"tel":            p.Tel,
				"email":          p.Email,
				"email_personal": p.EmailPersonal,
				"status":         "created",
			})
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query().Get("q")
		zona := r.URL.Query().Get("zona")

		queryStr := "SELECT id, zona, agente, note, tel, email, email_personal FROM agents WHERE 1=1"
		var args []any

		if q != "" {
			queryStr += " AND (agente LIKE ? OR zona LIKE ? OR email_personal LIKE ?)"
			likeQ := "%" + q + "%"
			args = append(args, likeQ, likeQ, likeQ)
		}
		if zona != "" {
			queryStr += " AND zona LIKE ?"
			args = append(args, "%" + zona + "%")
		}

		queryStr += " ORDER BY agente ASC"

		rows, err := h.db.Query(queryStr, args...)
		if err != nil {
			log.Printf("Error querying agents: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type AgentJSON struct {
			ID            string `json:"id"`
			Zona          string `json:"zona"`
			Agente        string `json:"agente"`
			Note          string `json:"note"`
			Tel           string `json:"tel"`
			Email         string `json:"email"`
			EmailPersonal string `json:"email_personal"`
		}

		agents := []AgentJSON{}
		for rows.Next() {
			var id int
			var a AgentJSON
			err := rows.Scan(&id, &a.Zona, &a.Agente, &a.Note, &a.Tel, &a.Email, &a.EmailPersonal)
			if err != nil {
				log.Printf("Error scanning agent: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			a.ID = strconv.Itoa(id)
			agents = append(agents, a)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) AgentsDetailHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 3 || parts[2] == "" {
				http.Error(w, "Invalid agent ID", http.StatusBadRequest)
				return
			}
			agentIDStr := parts[2]
			agentID, err := strconv.Atoi(agentIDStr)
			if err != nil {
				http.Error(w, "Invalid agent ID", http.StatusBadRequest)
				return
			}

			res, err := h.db.Exec("DELETE FROM agents WHERE id = ?", agentID)
			if err != nil {
				log.Printf("Error deleting agent: %v", err)
				http.Error(w, "Failed to delete agent", http.StatusInternalServerError)
				return
			}
			rowsAffected, _ := res.RowsAffected()
			if rowsAffected == 0 {
				http.Error(w, "Agent not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "deleted",
			})
			return
		}

		if r.Method == http.MethodPut {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 3 || parts[2] == "" {
				http.Error(w, "Invalid agent ID", http.StatusBadRequest)
				return
			}
			agentIDStr := parts[2]
			agentID, err := strconv.Atoi(agentIDStr)
			if err != nil {
				http.Error(w, "Invalid agent ID", http.StatusBadRequest)
				return
			}

			type UpdateAgentPayload struct {
				Zona          string `json:"zona"`
				Agente        string `json:"agente"`
				Note          string `json:"note"`
				Tel           string `json:"tel"`
				Email         string `json:"email"`
				EmailPersonal string `json:"email_personal"`
				Password      string `json:"password"`
			}

			var p UpdateAgentPayload
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if p.Agente == "" || p.Zona == "" {
				http.Error(w, "Missing required fields (agente, zona)", http.StatusBadRequest)
				return
			}

			cleanedZona := strings.ToUpper(strings.TrimSpace(p.Zona))
			cleanedAgente := strings.ToUpper(strings.TrimSpace(p.Agente))

			var res sql.Result
			if p.Password != "" {
				var hashedPassword []byte
				hashedPassword, err = bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
				if err != nil {
					log.Printf("Error hashing password: %v", err)
					http.Error(w, "Error processing password", http.StatusInternalServerError)
					return
				}
				res, err = h.db.Exec(`
					UPDATE agents 
					SET zona = ?, agente = ?, note = ?, tel = ?, email = ?, email_personal = ?, password = ?
					WHERE id = ?
				`, cleanedZona, cleanedAgente, p.Note, p.Tel, p.Email, p.EmailPersonal, string(hashedPassword), agentID)
			} else {
				res, err = h.db.Exec(`
					UPDATE agents 
					SET zona = ?, agente = ?, note = ?, tel = ?, email = ?, email_personal = ?
					WHERE id = ?
				`, cleanedZona, cleanedAgente, p.Note, p.Tel, p.Email, p.EmailPersonal, agentID)
			}

			if err != nil {
				log.Printf("Error updating agent: %v", err)
				http.Error(w, "Failed to update agent", http.StatusInternalServerError)
				return
			}

			rowsAffected, _ := res.RowsAffected()
			if rowsAffected == 0 {
				http.Error(w, "Agent not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"id":             strconv.Itoa(agentID),
				"zona":           cleanedZona,
				"agente":         cleanedAgente,
				"note":           p.Note,
				"tel":            p.Tel,
				"email":          p.Email,
				"email_personal": p.EmailPersonal,
				"status":         "updated",
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		type LoginPayload struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var p LoginPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(p.Email)
		password := p.Password

		if email == "" || password == "" {
			http.Error(w, "Email e password richiesti", http.StatusBadRequest)
			return
		}

		var id int
		var dbEmail, dbAgente, dbZona, dbPassword string
		err := h.db.QueryRow(`
			SELECT id, email, agente, zona, password 
			FROM agents 
			WHERE LOWER(email) = LOWER(?)
		`, email).Scan(&id, &dbEmail, &dbAgente, &dbZona, &dbPassword)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Credenziali non valide", http.StatusUnauthorized)
				return
			}
			log.Printf("Login error querying DB: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Verify password using bcrypt if hashed, or direct comparison if plaintext legacy
		var validPassword bool
		if strings.HasPrefix(dbPassword, "$2a$") || strings.HasPrefix(dbPassword, "$2b$") || strings.HasPrefix(dbPassword, "$2y$") {
			validPassword = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password)) == nil
		} else {
			validPassword = (dbPassword == password)
		}

		if !validPassword {
			http.Error(w, "Credenziali non valide", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":     strconv.Itoa(id),
			"email":  dbEmail,
			"agente": dbAgente,
			"zona":   dbZona,
		})
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

func parseFloat(val string) float64 {
	if val == "" {
		return 0.0
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0
	}
	return f
}

func (h *Handler) CustomizationsHandler(w http.ResponseWriter, r *http.Request) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Parse multipart form
			err := r.ParseMultipartForm(50 << 20) // up to 50MB
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Get fields
			agentIdStr := r.FormValue("agent_id")
			agentId, _ := strconv.Atoi(agentIdStr)
			agentName := r.FormValue("agent_name")
			customerCode := r.FormValue("customer_code")
			customerBusinessName := r.FormValue("customer_business_name")
			customerStoreName := r.FormValue("customer_store_name")
			customerAddress := r.FormValue("customer_address")
			customerCap := r.FormValue("customer_cap")
			customerCity := r.FormValue("customer_city")
			customerEmail := r.FormValue("customer_email")
			customerPhone := r.FormValue("customer_phone")
			annualSellInEstimate := r.FormValue("annual_sell_in_estimate")

			// Cust 1
			cust1Subject := r.FormValue("cust1_subject")
			cust1Type := r.FormValue("cust1_type")
			cust1Width := parseFloat(r.FormValue("cust1_width_cm"))
			cust1Height := parseFloat(r.FormValue("cust1_height_cm"))
			cust1Material := r.FormValue("cust1_material")

			// Cust 2
			cust2Subject := r.FormValue("cust2_subject")
			cust2Type := r.FormValue("cust2_type")
			var cust2Width, cust2Height float64
			if r.FormValue("cust2_width_cm") != "" {
				cust2Width = parseFloat(r.FormValue("cust2_width_cm"))
			}
			if r.FormValue("cust2_height_cm") != "" {
				cust2Height = parseFloat(r.FormValue("cust2_height_cm"))
			}
			cust2Material := r.FormValue("cust2_material")

			// Cust 3
			cust3Subject := r.FormValue("cust3_subject")
			cust3Type := r.FormValue("cust3_type")
			var cust3Width, cust3Height float64
			if r.FormValue("cust3_width_cm") != "" {
				cust3Width = parseFloat(r.FormValue("cust3_width_cm"))
			}
			if r.FormValue("cust3_height_cm") != "" {
				cust3Height = parseFloat(r.FormValue("cust3_height_cm"))
			}
			cust3Material := r.FormValue("cust3_material")

			startDate := r.FormValue("start_date")
			endDate := r.FormValue("end_date")

			printingCostResponsibility := r.FormValue("printing_cost_responsibility")
			assemblyCostResponsibility := r.FormValue("assembly_cost_responsibility")

			shippingAddress := r.FormValue("shipping_address")
			shippingCivic := r.FormValue("shipping_civic")
			shippingCity := r.FormValue("shipping_city")
			shippingProvince := r.FormValue("shipping_province")
			shippingCap := r.FormValue("shipping_cap")

			// Get photo file
			file, header, err := r.FormFile("photo")
			if err != nil {
				http.Error(w, "Photo file is required", http.StatusBadRequest)
				return
			}
			defer file.Close()

			// Validate that the file is an image
			buff := make([]byte, 512)
			if _, err := file.Read(buff); err != nil {
				http.Error(w, "Failed to read file headers", http.StatusInternalServerError)
				return
			}
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				http.Error(w, "Failed to reset file pointer", http.StatusInternalServerError)
				return
			}
			contentType := http.DetectContentType(buff)
			if !strings.HasPrefix(contentType, "image/") {
				http.Error(w, "Only image files are allowed", http.StatusBadRequest)
				return
			}

			// Save file
			os.MkdirAll("uploads", 0755)
			filename := fmt.Sprintf("customization_%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
			filePath := "uploads/" + filename

			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file contents", http.StatusInternalServerError)
				return
			}

			// Save to DB
			now := time.Now()
			res, err := h.db.Exec(`
				INSERT INTO customizations (
					agent_id, agent_name, customer_code, customer_business_name, customer_store_name,
					customer_address, customer_cap, customer_city, customer_email, customer_phone,
					annual_sell_in_estimate,
					cust1_subject, cust1_type, cust1_width_cm, cust1_height_cm, cust1_material,
					cust2_subject, cust2_type, cust2_width_cm, cust2_height_cm, cust2_material,
					cust3_subject, cust3_type, cust3_width_cm, cust3_height_cm, cust3_material,
					start_date, end_date,
					printing_cost_responsibility, assembly_cost_responsibility,
					shipping_address, shipping_civic, shipping_city, shipping_province, shipping_cap,
					photo_url, created_at
				) VALUES (
					?, ?, ?, ?, ?,
					?, ?, ?, ?, ?,
					?,
					?, ?, ?, ?, ?,
					?, ?, ?, ?, ?,
					?, ?, ?, ?, ?,
					?, ?,
					?, ?,
					?, ?, ?, ?, ?,
					?, ?
				)
			`, agentId, agentName, customerCode, customerBusinessName, customerStoreName,
				customerAddress, customerCap, customerCity, customerEmail, customerPhone,
				annualSellInEstimate,
				cust1Subject, cust1Type, cust1Width, cust1Height, cust1Material,
				cust2Subject, cust2Type, cust2Width, cust2Height, cust2Material,
				cust3Subject, cust3Type, cust3Width, cust3Height, cust3Material,
				startDate, endDate,
				printingCostResponsibility, assemblyCostResponsibility,
				shippingAddress, shippingCivic, shippingCity, shippingProvince, shippingCap,
				"/uploads/"+filename, now)

			if err != nil {
				log.Printf("Error inserting customization: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			lastID, _ := res.LastInsertId()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id":     lastID,
				"status": "created",
			})
			return
		}

		if r.Method == http.MethodGet {
			// Query DB
			rows, err := h.db.Query(`
				SELECT 
					id, agent_id, agent_name, customer_code, customer_business_name, customer_store_name,
					customer_address, customer_cap, customer_city, customer_email, customer_phone,
					annual_sell_in_estimate,
					cust1_subject, cust1_type, cust1_width_cm, cust1_height_cm, cust1_material,
					cust2_subject, cust2_type, cust2_width_cm, cust2_height_cm, cust2_material,
					cust3_subject, cust3_type, cust3_width_cm, cust3_height_cm, cust3_material,
					start_date, end_date,
					printing_cost_responsibility, assembly_cost_responsibility,
					shipping_address, shipping_civic, shipping_city, shipping_province, shipping_cap,
					photo_url, created_at
				FROM customizations
				ORDER BY created_at DESC
			`)
			if err != nil {
				log.Printf("Error querying customizations: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			type CustomizationResponse struct {
				ID                         int       `json:"id"`
				AgentID                    int       `json:"agent_id"`
				AgentName                  string    `json:"agent_name"`
				CustomerCode               string    `json:"customer_code"`
				CustomerBusinessName       string    `json:"customer_business_name"`
				CustomerStoreName          string    `json:"customer_store_name"`
				CustomerAddress            string    `json:"customer_address"`
				CustomerCap                string    `json:"customer_cap"`
				CustomerCity               string    `json:"customer_city"`
				CustomerEmail              string    `json:"customer_email"`
				CustomerPhone              string    `json:"customer_phone"`
				AnnualSellInEstimate       string    `json:"annual_sell_in_estimate"`
				Cust1Subject               string    `json:"cust1_subject"`
				Cust1Type                  string    `json:"cust1_type"`
				Cust1WidthCm               float64   `json:"cust1_width_cm"`
				Cust1HeightCm              float64   `json:"cust1_height_cm"`
				Cust1Material              string    `json:"cust1_material"`
				Cust2Subject               *string   `json:"cust2_subject"`
				Cust2Type                  *string   `json:"cust2_type"`
				Cust2WidthCm               *float64  `json:"cust2_width_cm"`
				Cust2HeightCm              *float64  `json:"cust2_height_cm"`
				Cust2Material              *string   `json:"cust2_material"`
				Cust3Subject               *string   `json:"cust3_subject"`
				Cust3Type                  *string   `json:"cust3_type"`
				Cust3WidthCm               *float64  `json:"cust3_width_cm"`
				Cust3HeightCm              *float64  `json:"cust3_height_cm"`
				Cust3Material              *string   `json:"cust3_material"`
				StartDate                  string    `json:"start_date"`
				EndDate                    string    `json:"end_date"`
				PrintingCostResponsibility string    `json:"printing_cost_responsibility"`
				AssemblyCostResponsibility string    `json:"assembly_cost_responsibility"`
				ShippingAddress            string    `json:"shipping_address"`
				ShippingCivic              string    `json:"shipping_civic"`
				ShippingCity               string    `json:"shipping_city"`
				ShippingProvince           string    `json:"shipping_province"`
				ShippingCap                string    `json:"shipping_cap"`
				PhotoURL                   string    `json:"photo_url"`
				CreatedAt                  time.Time `json:"created_at"`
			}

			results := []CustomizationResponse{}
			for rows.Next() {
				var cr CustomizationResponse
				var startD, endD sql.NullString
				var agentIdVal sql.NullInt64
				var cust2Sub, cust2Type, cust2Mat sql.NullString
				var cust3Sub, cust3Type, cust3Mat sql.NullString
				var cust2W, cust2H, cust3W, cust3H sql.NullFloat64

				err := rows.Scan(
					&cr.ID, &agentIdVal, &cr.AgentName, &cr.CustomerCode, &cr.CustomerBusinessName, &cr.CustomerStoreName,
					&cr.CustomerAddress, &cr.CustomerCap, &cr.CustomerCity, &cr.CustomerEmail, &cr.CustomerPhone,
					&cr.AnnualSellInEstimate,
					&cr.Cust1Subject, &cr.Cust1Type, &cr.Cust1WidthCm, &cr.Cust1HeightCm, &cr.Cust1Material,
					&cust2Sub, &cust2Type, &cust2W, &cust2H, &cust2Mat,
					&cust3Sub, &cust3Type, &cust3W, &cust3H, &cust3Mat,
					&startD, &endD,
					&cr.PrintingCostResponsibility, &cr.AssemblyCostResponsibility,
					&cr.ShippingAddress, &cr.ShippingCivic, &cr.ShippingCity, &cr.ShippingProvince, &cr.ShippingCap,
					&cr.PhotoURL, &cr.CreatedAt,
				)
				if err != nil {
					log.Printf("Error scanning customization row: %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if agentIdVal.Valid {
					cr.AgentID = int(agentIdVal.Int64)
				}
				if cust2Sub.Valid {
					cr.Cust2Subject = &cust2Sub.String
				}
				if cust2Type.Valid {
					cr.Cust2Type = &cust2Type.String
				}
				if cust2Mat.Valid {
					cr.Cust2Material = &cust2Mat.String
				}
				if cust2W.Valid {
					cr.Cust2WidthCm = &cust2W.Float64
				}
				if cust2H.Valid {
					cr.Cust2HeightCm = &cust2H.Float64
				}
				if cust3Sub.Valid {
					cr.Cust3Subject = &cust3Sub.String
				}
				if cust3Type.Valid {
					cr.Cust3Type = &cust3Type.String
				}
				if cust3Mat.Valid {
					cr.Cust3Material = &cust3Mat.String
				}
				if cust3W.Valid {
					cr.Cust3WidthCm = &cust3W.Float64
				}
				if cust3H.Valid {
					cr.Cust3HeightCm = &cust3H.Float64
				}
				if startD.Valid {
					cr.StartDate = startD.String
				}
				if endD.Valid {
					cr.EndDate = endD.String
				}

				results = append(results, cr)
			}

			if err = rows.Err(); err != nil {
				log.Printf("Error during rows iteration: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(results)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	if h.corsMiddleware != nil {
		h.corsMiddleware(handlerFunc)(w, r)
		return
	}
	handlerFunc(w, r)
}

