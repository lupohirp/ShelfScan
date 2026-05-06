#!/bin/bash
set -e

echo "Waiting for Qdrant at ${QDRANT_URL}..."

# Use Python to wait for the port to be open
python3 << END
import socket
import time
import os
from urllib.parse import urlparse

url = os.getenv('QDRANT_URL', 'http://qdrant:6333')
parsed = urlparse(url)
host = parsed.hostname
port = parsed.port or 6333

print(f"Checking connection to {host}:{port}...")
while True:
    try:
        with socket.create_connection((host, port), timeout=1):
            print("Qdrant is up and reachable!")
            break
    except (socket.timeout, ConnectionRefusedError, socket.gaierror):
        print("Waiting for Qdrant...")
        time.sleep(2)
END

# Run the initialization script
echo "Initializing Qdrant collection..."
python3 init_qdrant.py

# Start the main application
echo "Starting Embeddings API on 0.0.0.0:8001..."
# Use python3 -u to avoid buffered output and ensure uvicorn is correctly invoked
python3 -u main.py
