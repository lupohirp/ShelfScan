# 💍 ShelfScan

**ShelfScan** is an AI-powered automated inventory verification system for jewelry stores. It processes high-resolution shelf photos using computer vision, multimodal LLM visual analysis (Google Gemini), and vector search (Qdrant) to detect missing or displaced items on jewelry display shelves with high precision.

---

## 🌟 Key Features & Pipeline

1. **EXIF Auto-Orientation & Preprocessing**:
   Automatically detects EXIF orientation metadata in uploaded photos and rotates them to be upright prior to detection and cropping.

2. **AI Object Detection with Model Fallback**:
   Uses Google Gemini (`models/gemini-3.6-flash`, `gemini-3.1-flash-lite`, `gemini-3.5-flash-lite`, `gemini-3-flash` with automatic fallback rotation) to detect real jewelry/watch items and extract normalized bounding boxes.

3. **10% Bounding Box Padding**:
   Adds a 10% safety margin around detected item bounding boxes prior to cropping, ensuring straps, bezels, pendants, and outer edges of jewelry are never cut off.

4. **Multimodal Vector Indexing & Search**:
   Generates 768-dimensional visual embeddings via Google's `gemini-embedding-2` model and performs cosine similarity search against a self-hosted Qdrant Vector DB.

5. **Dynamic Category-Aware Side-by-Side Verification**:
   Performs a 2nd-step verification using Gemini Vision by comparing the cropped shelf image side-by-side with official inventory catalog images, dynamically tailoring the checklist for watches, rings, earrings, bracelets, or necklaces to eliminate false positives.

6. **4K Dual PWA System**:
   - **Main Client App**: Scanning interface for store staff with real-time DeviceOrientation guidance.
   - **Admin Client App**: Inventory and catalog management system.

---

## 🏗️ Architecture

The system is built as a microservices-oriented stack:

- **Frontend**: Two React + TypeScript PWAs built with Vite.
- **Backend API (Go)**: High-performance Go service handling image decoding, EXIF orientation, crop generation, Gemini orchestration, Qdrant vector search, and two-step visual verification.
- **MCP Server (Go)**: Model Context Protocol implementation for tool integration.
- **Vector DB**: Qdrant (Self-hosted gRPC/REST vector storage).
- **Embeddings Provider**: Google Gemini (`models/gemini-embedding-2:embedContent`, 768 dimensions).
- **Backup Service**: `offen/docker-volume-backup` sidecar container.

---

## 🚀 Quick Start (Docker)

### 1. Environment Setup
Create a `.env` file in the root directory (refer to `.env.example`):
```env
GEMINI_API_KEY=YOUR_GEMINI_API_KEY
GEMINI_MODEL=models/gemini-3.6-flash
GEMINI_TEMPERATURE=0.3
GEMINI_MAX_OUTPUT_TOKENS=8192
```

### 2. Build and Launch
```bash
# Start all services with Docker Compose
docker compose up -d --build
```

---

## 🛠️ Local Ports & Services

- **Main Scanning PWA**: [http://localhost:5175](http://localhost:5175)
- **Admin PWA**: [http://localhost:5176](http://localhost:5176)
- **Go API**: [http://localhost:8085](http://localhost:8085)
- **MCP Server**: [http://localhost:8087](http://localhost:8087)
- **Qdrant Vector DB**: [http://localhost:6335](http://localhost:6335) (gRPC: `6336`)

---

## 💾 Backup & Restore

A daily backup sidecar container (`offen/docker-volume-backup`) backs up the SQLite database, media uploads (`/app/uploads`), and Qdrant storage.

- **Schedule**: Daily at 3:00 AM (`0 3 * * *`).
- **Retention**: Last 14 days.
- **Location**: Saved inside `./backups` on the host.

---

## 🚢 CI/CD & Deployment

Deployments are automated via GitHub Actions (`.github/workflows/deploy-on-tag.yml`).

Pushing a version tag triggers the container build and deployment:
- Development server: `git tag 0.0.X.dev && git push origin 0.0.X.dev`
- Reembed migration: `git tag 0.0.X.reembed.dev && git push origin 0.0.X.reembed.dev`
- Production server: `git tag v0.0.X.prod && git push origin v0.0.X.prod`
