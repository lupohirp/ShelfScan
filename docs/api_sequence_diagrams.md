# 📊 ShelfScan API Sequence Diagrams

This document contains detailed sequence diagrams for the primary API endpoints implemented in the Go API backend.

---

## 1. Automated Shelf Analysis (`POST /analyze`)
This API endpoint receives one or more high-resolution shelf images, uses Gemini to locate items, crops detected bounding boxes, requests embeddings for each cropped item, queries Qdrant for matches, and returns an inventory comparison report (found vs. missing products).

```mermaid
sequenceDiagram
    autonumber
    actor Client as PWA Client
    participant API as Go API (server)
    participant Gemini as Gemini API
    participant ML as Embedding Service (Python)
    participant Qdrant as Qdrant DB

    Client->>API: POST /analyze {images: [...]}
    activate API
    Note over API: Create log folder req_<timestamp><br/>for trace logging
    
    par For Each Image (Concurrent)
        API->>API: Save original image to req dir
        API->>API: Resize image to max 1200x1200px
        API->>Gemini: POST /v1/models/gemini-...:generateContent {image}
        Gemini-->>API: JSON Response (bounding boxes + descriptions)
        API->>API: Parse items list (desc, box)
        
        par For Each Detected Box (Concurrent)
            API->>API: Crop detected box from original image
            API->>API: Save crop image to uploads/ & req dir
            API->>ML: POST /embed {crop_image}
            ML-->>API: 512-dim embedding vector
            API->>Qdrant: POST /collections/jewelry_inventory/points/query {vector}
            Qdrant-->>API: Top-N match hits
            API->>API: Score & filter hits (category conflict, color conflict)
            API->>API: Save best match above threshold
        end
    end
    
    API->>Qdrant: GET /collections/jewelry_inventory/points/scroll
    Qdrant-->>API: Full jewelry catalog
    API->>API: Perform diff analysis (catalog vs. matched items)
    API->>API: Log final request statistics JSON
    API-->>Client: 200 OK {found: [...], missing: [...], imageResults: [...]}
    deactivate API
```

---

## 2. Image-Based Vector Search (`POST /search`)
This endpoint accepts a photo of a single jewelry item and returns the closest matches from the Qdrant vector database.

```mermaid
sequenceDiagram
    autonumber
    actor Client as PWA Client / Admin
    participant API as Go API (server)
    participant ML as Embedding Service (Python)
    participant Qdrant as Qdrant DB

    Client->>API: POST /search {image}
    activate API
    API->>ML: POST /embed {image}
    activate ML
    Note over ML: Sentence-Transformers CLIP
    ML-->>API: 512-dim vector
    deactivate ML
    API->>Qdrant: POST /collections/jewelry_inventory/points/query {vector}
    activate Qdrant
    Qdrant-->>API: Nearest hits (SKUs, scores, payloads)
    deactivate Qdrant
    API-->>Client: 200 OK [{sku, name, score, imageUrl}]
    deactivate API
```

---

## 3. Product Inventory Onboarding (`POST /upload`)
This endpoint indexes one or more images of a jewelry item, generates their embeddings, and stores them in Qdrant with associated metadata.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin PWA
    participant API as Go API (server)
    participant ML as Embedding Service (Python)
    participant Qdrant as Qdrant DB

    Admin->>API: POST /upload {sku, name, color, material, images, append_mode}
    activate API
    API->>API: Save original images to static uploads/ folder
    
    par For Each Image (Concurrent)
        API->>ML: POST /embed {image}
        ML-->>API: 512-dim vector
    end
    
    API->>Qdrant: PUT /collections/jewelry_inventory/points {points: [vectors + payload]}
    Qdrant-->>API: 200 Success
    API-->>Admin: 200 OK {status: "success"}
    deactivate API
```
