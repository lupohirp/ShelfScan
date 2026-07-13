# 📊 ShelfScan API Sequence Diagrams

This document contains detailed sequence diagrams for the primary API endpoints and component interactions in the ShelfScan system.

---

## 1. Automated Shelf Analysis (`POST /analyze`)
This API endpoint receives one or more high-resolution shelf images, uses Gemini to locate items, crops detected bounding boxes, requests embeddings for each cropped item using the Gemini Embedding 2 API, queries Qdrant for matches, and returns an inventory comparison report (found vs. missing products).

```mermaid
sequenceDiagram
    autonumber
    actor Client as PWA Client
    participant API as Go API (server)
    participant GeminiAPI as Google Gemini API
    participant Qdrant as Qdrant DB

    Client->>API: POST /analyze {images: [...]}
    activate API
    Note over API: Create log folder req_<timestamp><br/>for trace logging
    
    par For Each Image (Concurrent)
        API->>API: Save original image to req dir
        API->>API: Resize image to max 1200x1200px
        API->>GeminiAPI: POST /models/gemini-...:generateContent {image}
        GeminiAPI-->>API: JSON Response (bounding boxes + descriptions)
        API->>API: Parse items list (desc, box)
        
        par For Each Detected Box (Concurrent)
            API->>API: Crop detected box from original image
            API->>API: Save crop image to uploads/ & req dir
            
            API->>GeminiAPI: POST /models/gemini-embedding-2:embedContent {crop_image}
            Note over GeminiAPI: Gemini Embeddings 2
            GeminiAPI-->>API: 768-dim embedding vector
            
            API->>Qdrant: Search collections/jewelry_inventory/points {vector}
            Qdrant-->>API: Top-N match hits
            API->>API: Score & filter hits (category conflict, color conflict)
            API->>API: Save best match above threshold
        end
    end
    
    API->>Qdrant: Scroll collections/jewelry_inventory/points
    Qdrant-->>API: Full jewelry catalog
    API->>API: Perform diff analysis (catalog vs. matched items)
    API->>API: Log final request statistics JSON
    API-->>Client: 200 OK {found: [...], missing: [...], imageResults: [...]}
    deactivate API
```

---

## 2. Image-Based Vector Search (`POST /search`)
This endpoint accepts a photo of a single jewelry item, generates its vector using Gemini Embeddings 2, and returns the closest matches from the Qdrant vector database.

```mermaid
sequenceDiagram
    autonumber
    actor Client as PWA Client / Admin
    participant API as Go API (server)
    participant GeminiAPI as Google Gemini API
    participant Qdrant as Qdrant DB

    Client->>API: POST /search {image}
    activate API
    
    API->>GeminiAPI: POST /models/gemini-embedding-2:embedContent {image}
    activate GeminiAPI
    GeminiAPI-->>API: 768-dim vector
    deactivate GeminiAPI
    
    API->>Qdrant: Search collections/jewelry_inventory/points {vector}
    activate Qdrant
    Qdrant-->>API: Nearest hits (SKUs, scores, payloads)
    deactivate Qdrant
    
    API-->>Client: 200 OK [{sku, name, score, imageUrl}]
    deactivate API
```

---

## 3. Product Inventory Onboarding (`POST /upload`)
This endpoint indexes one or more images of a jewelry item, generates their embeddings using Gemini Embeddings 2, and stores them in Qdrant with associated metadata.

```mermaid
sequenceDiagram
    autonumber
    actor Admin as Admin PWA
    participant API as Go API (server)
    participant GeminiAPI as Google Gemini API
    participant Qdrant as Qdrant DB

    Admin->>API: POST /upload {sku, name, color, material, images, append_mode}
    activate API
    API->>API: Save original images to static uploads/ folder
    
    par For Each Image (Concurrent)
        API->>GeminiAPI: POST /models/gemini-embedding-2:embedContent {image}
        GeminiAPI-->>API: 768-dim vector
    end
    
    API->>Qdrant: PUT /collections/jewelry_inventory/points {points: [vectors + payload]}
    Qdrant-->>API: 200 Success
    API-->>Admin: 200 OK {status: "success"}
    deactivate API
```

---

## 4. MCP Tools Orchestration & Gemini Function Calling (Main Flow)
This diagram illustrates how the PWA Client uses Model Context Protocol (MCP) tools during a live camera scan. The client exposes MCP tools as functions to Gemini via the Google AI SDK, which Gemini can execute by invoking tools on the Go MCP Server via WebSocket.

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant App as Main PWA (client)
    participant GeminiAPI as Google Gemini API
    participant MCP as Go MCP Server (WebSocket)
    participant Qdrant as Qdrant DB

    User->>App: Opens camera & aligns shelf
    App->>App: Auto-captures steady photo
    
    App->>GeminiAPI: POST /models/gemini-...:generateContent {image, tools_declaration}
    activate GeminiAPI
    Note over GeminiAPI: Gemini detects gap in layout,<br/>needs to look up shelf inventory
    
    GeminiAPI-->>App: Call Tool: get_layout(shelf_id)
    
    App->>MCP: WebSocket Send (JSON-RPC tools/call: get_layout)
    activate MCP
    Note over MCP: Query layout from SQLite/Store metadata
    MCP-->>App: WebSocket Recv (JSON-RPC result: items list)
    deactivate MCP
    
    App->>GeminiAPI: Return Tool Response (items list)
    
    Note over GeminiAPI: Gemini needs to perform vector search<br/>on a cropped item vector
    GeminiAPI-->>App: Call Tool: vector_search(embedding)
    
    App->>MCP: WebSocket Send (JSON-RPC tools/call: vector_search)
    activate MCP
    MCP->>Qdrant: PerformVectorSearch(embedding)
    Qdrant-->>MCP: Similarity matches
    MCP-->>App: WebSocket Recv (JSON-RPC result: matches)
    deactivate MCP
    
    App->>GeminiAPI: Return Tool Response (matches)
    
    Note over GeminiAPI: Final reasoning & report creation
    GeminiAPI-->>App: Analysis Report (found/missing products JSON)
    deactivate GeminiAPI
    
    App->>User: Render discrepancies UI (bubble overlays)
```
