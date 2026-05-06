# 💍 ShelfScan

**ShelfScan** is an advanced AI-powered inventory verification system designed for jewelry stores. It uses high-resolution photography, computer vision (CLIP embeddings), and Large Language Models (Gemini) to automatically detect missing or displaced items on jewelry shelves.

## 🌟 Key Features

- **4K Dual PWA System**: 
  - **Main App**: Used by staff to scan shelves with automated leveling (DeviceOrientation) and auto-capture.
  - **Admin App**: Used to manage and index jewelry inventory with metadata and photos.
- **AI-Powered Analysis**: Uses Google Gemini to analyze visual discrepancies and identify items.
- **Vector Intelligence**: Employs Qdrant Vector DB and CLIP embeddings for sub-second visual search and identification.
- **Real-time Feedback**: Visual "bubble level" overlay ensures perfect photo alignment for maximum AI accuracy.

## 🏗️ Architecture

The project is built with a modern, microservices-oriented stack:

- **Frontend**: Two React + TypeScript PWAs (Vite, TailwindCSS).
- **Backend API (Go)**: High-performance service for inventory management and Qdrant orchestration.
- **MCP Server (Go)**: Model Context Protocol implementation for vector search and layout analysis.
- **Embedding Service (Python)**: Sidecar service using `sentence-transformers/clip-ViT-B-32` for image vectorization.
- **Vector DB**: Qdrant (Self-hosted).

## 🚀 Quick Start (Docker)

The easiest way to get ShelfScan running is via Docker Compose.

### 1. Environment Setup
Create a `.env` file in the root directory:
```env
VITE_GEMINI_API_KEY=your_google_gemini_api_key_here
```

### 2. Build and Launch
```bash
# Build with progress visibility
docker compose build --progress=plain

# Start all services
docker compose up -d
```

### 3. Initialize Database
```bash
docker compose exec embeddings python init_qdrant.py
```

## 🛠️ Local Development & Debugging

For the best experience, open the project using the VS Code Workspace file:
`ShelfScan.code-workspace`

### Available Ports:
- **Main PWA**: [http://localhost:5173](http://localhost:5173)
- **Admin PWA**: [http://localhost:5174](http://localhost:5174)
- **Go API**: [http://localhost:8080](http://localhost:8080)
- **MCP Server**: [http://localhost:8081](http://localhost:8081)
- **Qdrant Dashboard**: [http://localhost:6333/dashboard](http://localhost:6333/dashboard)

### Debugging in VS Code
The repository includes a pre-configured `launch.json`. You can debug the **Full Backend** (Go + Python) by pressing **F5** in the Run & Debug side panel.

## 📁 Directory Structure

- `client/`: Main scanning PWA.
- `admin-client/`: Inventory management PWA.
- `server/`: Go backend for data and vector indexing.
- `mcp/`: Go implementation of Model Context Protocol tools.
- `embeddings/`: Python ML service for CLIP embeddings.

## 📝 License
This project is for internal use. See the project specification document for further details.
