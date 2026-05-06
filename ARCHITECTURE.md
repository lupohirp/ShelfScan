# 🏗️ ShelfScan Architectural Sequences

This document illustrates the data flow and interactions between the various components of the ShelfScan system using Mermaid diagrams.

## 1. Jewelry Inventory Onboarding (Admin Flow)
This sequence occurs when a new jewelry item is added to the system via the Admin PWA.

```mermaid
sequenceDiagram
    participant Admin as Admin PWA
    participant API as Go API (server)
    participant ML as Embeddings Service
    participant Qdrant as Qdrant DB

    Admin->>API: POST /upload (Photo + Name)
    API->>ML: POST /embed (Image)
    Note over ML: CLIP-ViT-B-32 (CPU)
    ML-->>API: 512-dim Vector
    API->>Qdrant: Upsert Point (Vector + Payload)
    Qdrant-->>API: Success
    API-->>Admin: 200 OK (Indexed)
```

## 2. Shelf Scanning & AI Verification (Main Flow)
This sequence shows the interaction during a shelf check, including the leveling logic and AI orchestration.

```mermaid
sequenceDiagram
    actor User
    participant App as Main PWA (client)
    participant Gemini as Google Gemini AI
    participant MCP as Go MCP Server
    participant Qdrant as Qdrant DB

    User->>App: Opens Camera
    App->>App: Leveling Logic (Pitch/Roll < 5°)
    App->>App: Stability Timer (1.5s)
    App->>App: Auto-Capture Image
    
    App->>Gemini: Image + Prompt (Google AI SDK)
    
    activate Gemini
    Note right of Gemini: Identify items & detect gaps
    
    Gemini->>MCP: Call vector_search(embedding)
    MCP->>Qdrant: Search (Similarity)
    Qdrant-->>MCP: Nearest Metadata
    MCP-->>Gemini: Jewelry Identification
    
    Gemini->>MCP: Call get_layout(shelf_id)
    MCP-->>Gemini: List of expected items
    
    Note right of Gemini: Perform Diff Analysis
    
    Gemini-->>App: Analysis Report (JSON/Text)
    deactivate Gemini
    
    App->>User: Display Missing/Displaced Items
```

## 3. System Components Overview

| Service | Responsibility | Technology |
| :--- | :--- | :--- |
| **Main PWA** | UI for scanning, 4K Camera, Leveling, Gemini Integration | React, Vite, Tailwind |
| **Admin PWA** | Inventory management, metadata entry | React, Vite |
| **Go API** | Bridge for inventory indexing, Qdrant orchestration | Go (Standard Library + gRPC) |
| **MCP Server** | Providing tools to the AI Agent (Search, Layout) | Go (JSON-RPC) |
| **Embeddings** | Image vectorization using CLIP models | Python, FastAPI, PyTorch (CPU) |
| **Qdrant** | High-performance vector storage and retrieval | Qdrant (Rust-based) |
