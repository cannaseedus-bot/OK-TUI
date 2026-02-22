from __future__ import annotations

import struct

import torch

from quantize import QuantizedTensor

FORMAT_INT4 = 1
FORMAT_INT8 = 2
FORMAT_FP16 = 3


def pack_int4(values: torch.Tensor) -> bytes:
    flat = values.reshape(-1).to(torch.int8).cpu().tolist()
    packed = bytearray()

    for i in range(0, len(flat), 2):
        a = flat[i] & 0x0F
        b = flat[i + 1] & 0x0F if i + 1 < len(flat) else 0
        packed.append((b << 4) | a)

    return bytes(packed)


def _payload(quantized: QuantizedTensor, qtype: str) -> tuple[int, bytes]:
    if qtype == "int4":
        return FORMAT_INT4, pack_int4(quantized.values)
    if qtype == "int8":
        return FORMAT_INT8, quantized.values.reshape(-1).to(torch.int8).cpu().numpy().tobytes()
    if qtype == "fp16":
        return FORMAT_FP16, quantized.values.reshape(-1).to(torch.float16).cpu().numpy().tobytes()
    raise ValueError(f"unsupported quantization type: {qtype}")


def pack_lane(name: str, quantized: QuantizedTensor, shape: tuple[int, ...], qtype: str) -> bytes:
    name_bytes = name.encode("utf-8")
    fmt, data = _payload(quantized, qtype)

    header = struct.pack(
        "<I I f I I",
        len(name_bytes),
        quantized.values.numel(),
        float(quantized.scale.item()),
        fmt,
        len(shape),
    )
    dims = struct.pack(f"<{len(shape)}I", *shape)

    return header + name_bytes + dims + data
