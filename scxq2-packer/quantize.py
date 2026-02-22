from __future__ import annotations

from dataclasses import dataclass

import torch


@dataclass(frozen=True)
class QuantizedTensor:
    values: torch.Tensor
    scale: torch.Tensor


def quantize_int4(tensor: torch.Tensor) -> QuantizedTensor:
    tensor = tensor.detach().to(torch.float32)
    max_val = torch.amax(torch.abs(tensor))
    if max_val.item() == 0:
        scale = torch.tensor(1.0, dtype=torch.float32)
    else:
        scale = max_val / 7.0

    quantized = torch.round(tensor / scale)
    quantized = torch.clamp(quantized, -8, 7).to(torch.int8)
    return QuantizedTensor(values=quantized, scale=scale)


def quantize_int8(tensor: torch.Tensor) -> QuantizedTensor:
    tensor = tensor.detach().to(torch.float32)
    max_val = torch.amax(torch.abs(tensor))
    if max_val.item() == 0:
        scale = torch.tensor(1.0, dtype=torch.float32)
    else:
        scale = max_val / 127.0

    quantized = torch.round(tensor / scale)
    quantized = torch.clamp(quantized, -128, 127).to(torch.int8)
    return QuantizedTensor(values=quantized, scale=scale)


def quantize_fp16(tensor: torch.Tensor) -> QuantizedTensor:
    return QuantizedTensor(values=tensor.detach().to(torch.float16), scale=torch.tensor(1.0, dtype=torch.float32))
