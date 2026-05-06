FROM python:3.11-slim

WORKDIR /app

# Install curl for healthcheck
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

# Install CPU-only torch first to save space and avoid SSL issues with huge CUDA packages
RUN pip install --no-cache-dir -r requirements.txt

# Pre-download the model into the image
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('clip-ViT-B-32', device='cpu')"

COPY . .

# Make sure entrypoint is executable
RUN chmod +x entrypoint.sh

CMD ["./entrypoint.sh"]
