#!/usr/bin/env python3
"""Build Phi-3-mini-4k-instruct SCXQ2 shards for browser WebGPU runtimes."""

from __future__ import annotations

import hashlib
import json
import struct
from pathlib import Path

import torch
from huggingface_hub import snapshot_download
from transformers import AutoModelForCausalLM
from tqdm import tqdm

MODEL_ID = "microsoft/Phi-3-mini-4k-instruct"
OUTPUT_DIR = Path("phi3-scxq2")
MAX_SHARD = 256 * 1024 * 1024
MAGIC = b"SCXQ2"
VERSION = 1


def quantize_int4(tensor: torch.Tensor) -> tuple[torch.Tensor, float]:
    """Symmetric per-tensor INT4 quantization."""
    max_val = tensor.abs().max()
    if max_val.item() == 0:
        return torch.zeros_like(tensor, dtype=torch.int8), 1.0

    scale = float(max_val.item() / 7.0)
    q = torch.round(tensor / scale)
    q = torch.clamp(q, -8, 7)
    return q.to(torch.int8), scale


def pack_int4(values: torch.Tensor) -> bytes:
    """Pack signed INT4 values into bytes (two nibbles per byte)."""
    flat = values.flatten().tolist()
    out = bytearray()

    for i in range(0, len(flat), 2):
        a = int(flat[i]) & 0xF
        b = int(flat[i + 1]) & 0xF if i + 1 < len(flat) else 0
        out.append((b << 4) | a)

    return bytes(out)


def pack_tensor(name: str, tensor: torch.Tensor) -> bytes:
    """Serialize one quantized tensor as [header][name][payload]."""
    q, scale = quantize_int4(tensor)
    data = pack_int4(q)
    name_bytes = name.encode("utf-8")
    header = struct.pack("<I I f", len(name_bytes), len(data), scale)
    return header + name_bytes + data


def merkle_root(nodes: list[bytes]) -> bytes:
    """Compute deterministic SHA-256 Merkle root over tensor blocks."""
    if not nodes:
        return hashlib.sha256(b"").digest()

    hashes = [hashlib.sha256(node).digest() for node in nodes]

    while len(hashes) > 1:
        merged: list[bytes] = []
        for i in range(0, len(hashes), 2):
            a = hashes[i]
            b = hashes[i + 1] if i + 1 < len(hashes) else a
            merged.append(hashlib.sha256(a + b).digest())
        hashes = merged

    return hashes[0]


def split_shards(packed_tensors: list[bytes]) -> list[list[bytes]]:
    """Split serialized tensors into shard-sized chunks."""
    shards: list[list[bytes]] = []
    current: list[bytes] = []
    size = 0

    for tensor_blob in packed_tensors:
        if current and size + len(tensor_blob) > MAX_SHARD:
            shards.append(current)
            current = []
            size = 0

        current.append(tensor_blob)
        size += len(tensor_blob)

    if current:
        shards.append(current)

    return shards


def save_shard(shard: list[bytes], path: Path) -> None:
    """Write one SCXQ2 shard with header + Merkle root."""
    root = merkle_root(shard)

    with path.open("wb") as f:
        f.write(MAGIC)
        f.write(struct.pack("<I", VERSION))
        f.write(root)
        for tensor_blob in shard:
            f.write(tensor_blob)


def save_index(shards: list[list[bytes]]) -> None:
    index = {
        "magic": "SCXQ2_INDEX",
        "version": VERSION,
        "model": MODEL_ID,
        "shards": [f"shard_{i:03}.scxq2" for i in range(len(shards))],
        "hidden": 3072,
        "layers": 32,
    }

    (OUTPUT_DIR / "model.scxq2.index").write_text(json.dumps(index, indent=2), encoding="utf-8")


def save_transformers_meta(shards: list[list[bytes]]) -> None:
    meta = {
        "format": "scxq2",
        "hidden_size": 3072,
        "num_hidden_layers": 32,
        "num_attention_heads": 32,
        "shards": [f"shard_{i:03}.scxq2" for i in range(len(shards))],
    }

    (OUTPUT_DIR / "model.transformers.json").write_text(json.dumps(meta, indent=2), encoding="utf-8")


def build() -> None:
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    print("Downloading Phi-3...")
    local_path = snapshot_download(MODEL_ID)

    model = AutoModelForCausalLM.from_pretrained(
        local_path,
        torch_dtype=torch.float16,
        low_cpu_mem_usage=True,
    )

    packed: list[bytes] = []
    print("Quantizing...")
    for name, tensor in tqdm(sorted(model.state_dict().items())):
        if tensor.is_floating_point() and tensor.numel() > 0:
            packed.append(pack_tensor(name, tensor.cpu()))

    print("Splitting shards...")
    shards = split_shards(packed)

    print("Saving shards...")
    for i, shard in enumerate(shards):
        save_shard(shard, OUTPUT_DIR / f"shard_{i:03}.scxq2")

    save_index(shards)
    save_transformers_meta(shards)

    tokenizer_src = Path(local_path) / "tokenizer.json"
    if tokenizer_src.exists():
        (OUTPUT_DIR / "tokenizer.json").write_bytes(tokenizer_src.read_bytes())

    print("Done.")


if __name__ == "__main__":
    build()
