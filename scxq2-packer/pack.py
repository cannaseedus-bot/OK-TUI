from __future__ import annotations

import argparse
import json
import struct
from pathlib import Path
from typing import Iterator

import torch

from lane_pack import pack_lane
from merkle import merkle_root
from quantize import QuantizedTensor, quantize_fp16, quantize_int4, quantize_int8

MAGIC = b"SCXQ2"
VERSION = 1


def _read_model(path: Path) -> dict[str, torch.Tensor]:
    suffix = path.suffix.lower()

    if suffix == ".safetensors":
        from safetensors.torch import load_file

        return load_file(str(path), device="cpu")

    if suffix in {".bin", ".pt", ".pth"}:
        payload = torch.load(path, map_location="cpu")
        if isinstance(payload, dict) and "state_dict" in payload:
            payload = payload["state_dict"]
        if isinstance(payload, dict):
            return {k: v for k, v in payload.items() if isinstance(v, torch.Tensor)}

    if suffix == ".npy":
        import numpy as np

        arr = np.load(path)
        return {path.stem: torch.from_numpy(arr)}

    raise ValueError(f"unsupported model format for {path.name}")


def _quantize(tensor: torch.Tensor, qtype: str) -> QuantizedTensor:
    if qtype == "int4":
        return quantize_int4(tensor)
    if qtype == "int8":
        return quantize_int8(tensor)
    if qtype == "fp16":
        return quantize_fp16(tensor)
    raise ValueError(f"unsupported quantization mode: {qtype}")


def _iter_lanes(state_dict: dict[str, torch.Tensor], qtype: str) -> Iterator[bytes]:
    for name in sorted(state_dict.keys()):
        tensor = state_dict[name]
        if not isinstance(tensor, torch.Tensor) or tensor.numel() == 0:
            continue
        print(f"packing {name}")
        q = _quantize(tensor, qtype)
        yield pack_lane(name=name, quantized=q, shape=tuple(tensor.shape), qtype=qtype)


def pack_model(model_path: Path, output_path: Path, qtype: str) -> None:
    state_dict = _read_model(model_path)
    lanes = list(_iter_lanes(state_dict, qtype))
    root = merkle_root(lanes)

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("wb") as f:
        f.write(MAGIC)
        f.write(struct.pack("<I", VERSION))
        f.write(root)
        f.write(struct.pack("<I", len(lanes)))

        for lane in lanes:
            f.write(struct.pack("<I", len(lane)))
            f.write(lane)

        meta = {"quantization": qtype, "tensor_count": len(lanes)}
        meta_bytes = json.dumps(meta, sort_keys=True).encode("utf-8")
        f.write(struct.pack("<I", len(meta_bytes)))
        f.write(meta_bytes)

    print(f"wrote {output_path}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Pack model weights into SCXQ2 format")
    parser.add_argument("model_path", type=Path)
    parser.add_argument("output_path", type=Path)
    parser.add_argument("--quantization", choices=["int4", "int8", "fp16"], default="int4")
    args = parser.parse_args()

    pack_model(args.model_path, args.output_path, args.quantization)


if __name__ == "__main__":
    main()
