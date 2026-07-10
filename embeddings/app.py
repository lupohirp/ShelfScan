import gc
import io
import logging
import traceback

import torch
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from PIL import Image
from transformers import AutoModel, AutoProcessor

MODEL_ID = "google/siglip2-base-patch16-224"
MAX_SIDE = 512  # cap PIL decode footprint before the processor resizes to 224

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("shelfscan-embeddings")

processor = AutoProcessor.from_pretrained(MODEL_ID)
model = AutoModel.from_pretrained(MODEL_ID)
model.eval()
logger.info("model loaded: %s", MODEL_ID)

app = FastAPI(title="shelfscan-embeddings", version="1.0")


@app.exception_handler(Exception)
async def unhandled_exception_handler(_: Request, exc: Exception) -> JSONResponse:
    logger.error("unhandled exception:\n%s", traceback.format_exc())
    return JSONResponse(status_code=500, content={"detail": f"{type(exc).__name__}: {exc}"})


def _decode(data: bytes) -> Image.Image:
    if not data:
        raise HTTPException(status_code=400, detail="empty image body")
    try:
        image = Image.open(io.BytesIO(data)).convert("RGB")
    except Exception as exc:
        raise HTTPException(status_code=400, detail=f"invalid image: {exc}") from exc
    if max(image.size) > MAX_SIDE:
        image.thumbnail((MAX_SIDE, MAX_SIDE), Image.Resampling.BILINEAR)
    return image


def _embed_one(image: Image.Image) -> list[float]:
    inputs = processor(images=[image], return_tensors="pt")
    try:
        with torch.no_grad():
            features = model.get_image_features(**inputs)
        vec = features[0].tolist()
    finally:
        del inputs
        if "features" in locals():
            del features
        gc.collect()
    return vec


@app.get("/health")
def health() -> dict:
    return {"status": "ok", "model": MODEL_ID}


@app.post("/embed")
async def embed(request: Request) -> dict:
    body = await request.body()
    image = _decode(body)
    try:
        vec = _embed_one(image)
    finally:
        image.close()
    return {"embedding": vec}


@app.post("/embed_batch")
async def embed_batch(request: Request) -> dict:
    form = await request.form()
    files = form.getlist("image")
    if not files:
        raise HTTPException(status_code=400, detail="no images (form field: image)")

    vectors: list[list[float]] = []
    for idx, f in enumerate(files):
        data = await f.read()
        image = _decode(data)
        try:
            vectors.append(_embed_one(image))
        except Exception:
            logger.error("embed failed on batch item %d (filename=%s)", idx, getattr(f, "filename", "?"))
            raise
        finally:
            image.close()
    return {"embeddings": vectors}
