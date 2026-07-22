# 🏗️ ShelfScan Architectural Sequences

This document illustrates the data flow and interactions between the various components of the ShelfScan system using Mermaid diagrams.

## 1. Jewelry Inventory Onboarding (Admin Flow)
This sequence occurs when a new jewelry item is added to the system via the Admin PWA.

```mermaid
sequenceDiagram
    participant Admin as Admin PWA
    participant API as Go API (server)
    participant Gemini as Gemini Embeddings API
    participant Qdrant as Qdrant DB

    Admin->>API: POST /upload (Photo + Metadata)
    API->>Gemini: POST gemini-embedding-2:embedContent
    Gemini-->>API: 768-dim Vector
    API->>Qdrant: Upsert Point (Vector + Payload)
    Qdrant-->>API: Success
    API-->>Admin: 200 OK (Indexed)
```

## 2. Shelf Scanning & AI Verification (Main Flow)
This sequence shows the interaction during a shelf check, including EXIF rotation, Gemini object detection, vector search, and two-step side-by-side verification.

```mermaid
sequenceDiagram
    actor User
    participant App as Main PWA (client)
    participant API as Go API (server)
    participant Gemini as Google Gemini AI
    participant Qdrant as Qdrant DB

    User->>App: Leveling & Auto-Capture Image
    App->>API: POST /analyze (Image upload)
    
    API->>API: Auto-rotate EXIF orientation
    API->>Gemini: Detect items (gemini-3.6-flash + fallback rotation)
    Gemini-->>API: Detections + Normalized Bounding Boxes
    
    loop For each detected item
        API->>API: Apply 10% Bounding Box Padding & Crop
        API->>Gemini: Generate 768-dim embedding (gemini-embedding-2)
        Gemini-->>API: Vector
        API->>Qdrant: Perform Vector Search (Cosine)
        Qdrant-->>API: Candidate Matches (Scores & Payloads)
        
        opt Score >= 0.82
            API->>Gemini: Side-by-Side Category-Aware Verification
            Gemini-->>API: Match Boolean + Reasoning
        end
    end
    
    API-->>App: Analysis Response (Found & Missing items)
    App->>User: Display Results & Overlay
```

## 3. System Components Overview

| Service | Responsibility | Technology |
| :--- | :--- | :--- |
| **Main PWA** | UI for scanning, camera capture, Leveling guidance | React, Vite |
| **Admin PWA** | Inventory management, metadata entry | React, Vite |
| **Go API** | Image decoding, EXIF orientation, crop padding, Gemini & Qdrant orchestration | Go (Standard Library + imaging) |
| **MCP Server** | Model Context Protocol tools implementation | Go (WebSocket/JSON-RPC) |
| **Gemini AI** | Object detection, embeddings (768-dim), and two-step verification | Google Generative AI SDK / REST |
| **Qdrant** | High-performance vector storage and retrieval | Qdrant (Rust-based) |
