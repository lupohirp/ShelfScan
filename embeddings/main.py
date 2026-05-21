import os
import torch
from fastapi import FastAPI, UploadFile, File
from sentence_transformers import SentenceTransformer
from PIL import Image
import io
import gc

# Limit threads to improve performance while controlling memory overhead
torch.set_num_threads(4)

app = FastAPI()

# Load CLIP model on CPU
model = SentenceTransformer('clip-ViT-B-32', device='cpu')
model.eval()

@app.get("/health")
async def health():
    return {"status": "healthy"}

@app.post("/embed")
async def embed_image(file: UploadFile = File(...)):
    contents = await file.read()
    image = Image.open(io.BytesIO(contents))
    
    # Generate embedding without gradient calculation
    with torch.no_grad():
        embedding = model.encode(image, convert_to_numpy=True)
    
    # Clean up image from memory explicitly
    del image
    gc.collect()
    
    return {"embedding": embedding.tolist()}

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8001))
    uvicorn.run(app, host="0.0.0.0", port=port)
