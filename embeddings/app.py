import io
import logging
import traceback

import numpy as np
import onnxruntime as ort
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from PIL import Image

MODEL_ID = "google/siglip2-base-patch16-224"
MODEL_PATH = "/app/vision_encoder.onnx"
IMAGE_SIZE = 224
# SigLIP2 preprocessing: rescale 1/255, then normalize with mean=std=0.5 → [-1, 1]
NORM_MEAN = np.float32(0.5)
NORM_STD = np.float32(0.5)
MAX_SIDE = 1024  # cap PIL decode footprint before the final 224 resize

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("shelfscan-embeddings")

session_options = ort.SessionOptions()
session_options.intra_op_num_threads = 1
session_options.inter_op_num_threads = 1
session = ort.InferenceSession(MODEL_PATH, sess_options=session_options, providers=["CPUExecutionProvider"])
INPUT_NAME = session.get_inputs()[0].name
OUTPUT_NAME = session.get_outputs()[0].name
logger.info("loaded ONNX model %s (input=%s, output=%s)", MODEL_ID, INPUT_NAME, OUTPUT_NAME)

app = FastAPI(title="shelfscan-embeddings", version="2.0")


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


def _preprocess(image: Image.Image) -> np.ndarray:
    resized = image.resize((IMAGE_SIZE, IMAGE_SIZE), Image.Resampling.BICUBIC)
    arr = np.asarray(resized, dtype=np.float32) / 255.0
    arr = (arr - NORM_MEAN) / NORM_STD
    arr = arr.transpose(2, 0, 1)  # HWC → CHW
    return arr


def _embed_one(image: Image.Image) -> list[float]:
    tensor = _preprocess(image)[np.newaxis, ...]  # 1×3×224×224
    try:
        outputs = session.run([OUTPUT_NAME], {INPUT_NAME: tensor})
        return outputs[0][0].tolist()
    finally:
        del tensor


@app.get("/health")
def health() -> dict:
    return {"status": "ok", "model": MODEL_ID, "runtime": "onnxruntime"}


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
