import io

import torch
from fastapi import FastAPI, HTTPException, Request
from PIL import Image
from transformers import AutoModel, AutoProcessor

MODEL_ID = "google/siglip2-base-patch16-224"

processor = AutoProcessor.from_pretrained(MODEL_ID)
model = AutoModel.from_pretrained(MODEL_ID)
model.eval()

app = FastAPI(title="shelfscan-embeddings", version="1.0")


def _decode(data: bytes) -> Image.Image:
    if not data:
        raise HTTPException(status_code=400, detail="empty image body")
    try:
        return Image.open(io.BytesIO(data)).convert("RGB")
    except Exception as exc:
        raise HTTPException(status_code=400, detail=f"invalid image: {exc}") from exc


def _embed(images: list[Image.Image]) -> list[list[float]]:
    inputs = processor(images=images, return_tensors="pt")
    with torch.no_grad():
        features = model.get_image_features(**inputs)
    return features.tolist()


@app.get("/health")
def health() -> dict:
    return {"status": "ok", "model": MODEL_ID}


@app.post("/embed")
async def embed(request: Request) -> dict:
    body = await request.body()
    image = _decode(body)
    return {"embedding": _embed([image])[0]}


@app.post("/embed_batch")
async def embed_batch(request: Request) -> dict:
    form = await request.form()
    files = form.getlist("image")
    if not files:
        raise HTTPException(status_code=400, detail="no images (form field: image)")
    images = [_decode(await f.read()) for f in files]
    return {"embeddings": _embed(images)}
