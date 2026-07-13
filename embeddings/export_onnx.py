"""Export the SigLIP2 vision encoder + projection head to a single ONNX file.

Runs once at Docker build time in the builder stage. Runtime image never
touches torch or transformers — it just loads the produced .onnx with
onnxruntime.
"""

import sys
from pathlib import Path

import torch
from transformers import AutoModel

MODEL_ID = "google/siglip2-base-patch16-224"
OUT_PATH = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("/export/vision_encoder.onnx")
OUT_PATH.parent.mkdir(parents=True, exist_ok=True)


class VisionEncoder(torch.nn.Module):
    """Wraps AutoModel.get_image_features so ONNX export sees a clean forward."""

    def __init__(self, model: torch.nn.Module) -> None:
        super().__init__()
        self.model = model

    def forward(self, pixel_values: torch.Tensor) -> torch.Tensor:
        return self.model.get_image_features(pixel_values=pixel_values)


def main() -> None:
    model = AutoModel.from_pretrained(MODEL_ID)
    model.eval()
    wrapper = VisionEncoder(model)
    dummy = torch.zeros(1, 3, 224, 224, dtype=torch.float32)
    torch.onnx.export(
        wrapper,
        (dummy,),
        str(OUT_PATH),
        input_names=["pixel_values"],
        output_names=["image_features"],
        dynamic_axes={
            "pixel_values": {0: "batch"},
            "image_features": {0: "batch"},
        },
        opset_version=17,
        do_constant_folding=True,
    )
    print(f"exported {MODEL_ID} vision encoder to {OUT_PATH} ({OUT_PATH.stat().st_size / 1e6:.1f} MB)")


if __name__ == "__main__":
    main()
