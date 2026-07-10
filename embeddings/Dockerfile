FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    HF_HOME=/opt/hf-cache \
    TRANSFORMERS_OFFLINE=0

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir --index-url https://download.pytorch.org/whl/cpu torch==2.7.1 \
 && pip install --no-cache-dir \
        "fastapi==0.115.*" \
        "uvicorn[standard]==0.32.*" \
        "python-multipart==0.0.*" \
        "transformers==4.51.*" \
        "Pillow==11.*"

RUN python -c "from transformers import AutoProcessor, AutoModel; \
AutoProcessor.from_pretrained('google/siglip2-base-patch16-224'); \
AutoModel.from_pretrained('google/siglip2-base-patch16-224')"

ENV TRANSFORMERS_OFFLINE=1

COPY app.py .

EXPOSE 8001

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD curl -fsS http://localhost:8001/health || exit 1

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8001", "--workers", "1"]
