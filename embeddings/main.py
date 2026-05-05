import os
from fastapi import FastAPI, UploadFile, File
from sentence_transformers import SentenceTransformer
from PIL import Image
import io

app = FastAPI()

# Load CLIP model
model = SentenceTransformer('clip-ViT-B-32')

@app.get("/health")
async def health():
    return {"status": "healthy"}

@app.post("/embed")
async def embed_image(file: UploadFile = File(...)):
    contents = await file.read()
    image = Image.open(io.BytesIO(contents))
    
    # Generate embedding
    embedding = model.encode(image)
    
    return {"embedding": embedding.tolist()}

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8001))
    uvicorn.run(app, host="0.0.0.0", port=port)
