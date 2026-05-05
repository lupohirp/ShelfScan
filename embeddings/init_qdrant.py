import os
from qdrant_client import QdrantClient
from qdrant_client.http import models

def init_qdrant():
    qdrant_url = os.getenv("QDRANT_URL", "http://localhost:6333")
    client = QdrantClient(url=qdrant_url)
    
    collection_name = "jewelry_inventory"
    
    # Check if collection exists
    collections = client.get_collections().collections
    exists = any(c.name == collection_name for c in collections)
    
    if not exists:
        print(f"Creating collection: {collection_name}")
        client.recreate_collection(
            collection_name=collection_name,
            vectors_config=models.VectorParams(size=512, distance=models.Distance.COSINE),
        )
        print("Collection created successfully.")
    else:
        print(f"Collection {collection_name} already exists.")

if __name__ == "__main__":
    init_qdrant()
