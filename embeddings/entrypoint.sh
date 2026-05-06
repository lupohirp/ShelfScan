#!/bin/bash
set -e

# Wait for Qdrant to be ready (if needed, though depends_on with healthcheck is better)
echo "Waiting for Qdrant to be ready..."
# We rely on docker-compose healthchecks, but a small sleep doesn't hurt for network stability
sleep 2

# Run the initialization script
echo "Initializing Qdrant collection..."
python init_qdrant.py

# Start the main application
echo "Starting Embeddings API..."
python main.py
