#!/usr/bin/env python3
"""One-command HuggingFace -> SCXQ2 converter.

Example:
    python hf_to_scxq2.py Qwen/Qwen2-7B output/qwen2-7b.scxq2 --int4
"""

from __future__ import annotations

import argparse
import struct
from pathlib import Path

import torch
from huggingface_hub import snapshot_download
from tqdm import tqdm
from transformers import AutoModelForCausalLM

MAGIC = b"SCXQ2"
VERSION = 1


def _load_hf_state_dict(repo: str, revision: str | None, local_files_only: bool) -> dict[str, torch.Tensor]:
    local_path = snapshot_download(
        repo_id=repo,
        revision=revision,
        local_files_only=local_files_only,
    )

    model = AutoModelForCausalLM.from_pretrained(
        local_path,
        torch_dtype=torch.float16,
        low_cpu_mem_usage=True,
        local_files_only=local_files_only,
    )
    return model.state_dict()


def _quantize_int4(t: torch.Tensor) -> tuple[torch.Tensor, float]:
    t = t.detach().to(torch.float32)
    max_val = torch.amax(torch.abs(t))
    scale = 1.0 if max_val.item() == 0 else float(max_val.item() / 7.0)

    q = torch.round(t / scale)
    q = torch.clamp(q, -8, 7).to(torch.int8)
    return q, scale


def _quantize_int8(t: torch.Tensor) -> tuple[torch.Tensor, float]:
    t = t.detach().to(torch.float32)
    max_val = torch.amax(torch.abs(t))
    scale = 1.0 if max_val.item() == 0 else float(max_val.item() / 127.0)

    q = torch.round(t / scale)
    q = torch.clamp(q, -128, 127).to(torch.int8)
    return q, scale


def _pack_int4(values: torch.Tensor) -> bytes:
    flat = values.reshape(-1).to(torch.int8).cpu().tolist()
    packed = bytearray()

    for i in range(0, len(flat), 2):
        lo = flat[i] & 0x0F
        hi = flat[i + 1] & 0x0F if i + 1 < len(flat) else 0
        packed.append((hi << 4) | lo)

    return bytes(packed)


def _pack_tensor(name: str, tensor: torch.Tensor, mode: str) -> bytes:
    name_bytes = name.encode("utf-8")

    if mode == "int4":
        q, scale = _quantize_int4(tensor)
        data = _pack_int4(q)
        dtype = 1
    elif mode == "int8":
        q, scale = _quantize_int8(tensor)
        data = q.reshape(-1).to(torch.int8).cpu().numpy().tobytes()
        dtype = 2
    else:
        scale = 1.0
        data = tensor.detach().to(torch.float16).reshape(-1).cpu().numpy().tobytes()
        dtype = 3

    header = struct.pack("<I I I f", len(name_bytes), len(data), dtype, scale)
    return header + name_bytes + data


def _hash_node(data: bytes) -> bytes:
    import hashlib

    return hashlib.sha256(data).digest()


def _merkle_root(nodes: list[bytes]) -> bytes:
    if not nodes:
        return _hash_node(b"")

    hashes = [_hash_node(node) for node in nodes]
    while len(hashes) > 1:
        new_hashes: list[bytes] = []
        for i in range(0, len(hashes), 2):
            left = hashes[i]
            right = hashes[i + 1] if i + 1 < len(hashes) else left
            new_hashes.append(_hash_node(left + right))
        hashes = new_hashes

    return hashes[0]


def convert(repo: str, output: Path, mode: str, revision: str | None, local_files_only: bool) -> None:
    print(f"downloading/loading model: {repo}")
    state = _load_hf_state_dict(repo, revision=revision, local_files_only=local_files_only)

    print(f"packing tensors in {mode}...")
    lanes: list[bytes] = []
    for name in tqdm(sorted(state.keys()), desc="tensors"):
        tensor = state[name]
        if not isinstance(tensor, torch.Tensor) or not tensor.is_floating_point() or tensor.numel() == 0:
            continue
        lanes.append(_pack_tensor(name, tensor.cpu(), mode))

    root = _merkle_root(lanes)

    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open("wb") as f:
        f.write(MAGIC)
        f.write(struct.pack("<I", VERSION))
        f.write(root)
        for lane in lanes:
            f.write(lane)

    print(f"done: {output} ({len(lanes)} tensors)")


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Download a HuggingFace model and convert it to SCXQ2")
    parser.add_argument("repo", help="HuggingFace repo id (e.g. Qwen/Qwen2-7B)")
    parser.add_argument("output", type=Path, help="Output .scxq2 path")
    parser.add_argument("--revision", default=None, help="Optional HF branch/tag/commit")
    parser.add_argument("--local-files-only", action="store_true", help="Only use local HF cache")

    mode_group = parser.add_mutually_exclusive_group()
    mode_group.add_argument("--int4", action="store_true", help="Use INT4 quantization")
    mode_group.add_argument("--int8", action="store_true", help="Use INT8 quantization")
    mode_group.add_argument("--fp16", action="store_true", help="Use FP16 weights")
    return parser.parse_args()


def main() -> None:
    args = _parse_args()

    if args.int4:
        mode = "int4"
    elif args.int8:
        mode = "int8"
    else:
        mode = "fp16"

    convert(
        repo=args.repo,
        output=args.output,
        mode=mode,
        revision=args.revision,
        local_files_only=args.local_files_only,
    )


if __name__ == "__main__":
    main()
