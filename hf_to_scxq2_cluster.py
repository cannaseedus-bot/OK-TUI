#!/usr/bin/env python3
"""HuggingFace -> SCXQ2 clustered converter.

This converter extends the basic SCXQ2 packer with:
- automatic shard splitting
- MoE expert sharding
- entropy compression layer (ANS API)
- GPU-accelerated quantization path (torch backend)
- Transformers.js compatibility metadata
"""

from __future__ import annotations

import argparse
import json
import math
import re
import struct
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Iterable

import torch
from huggingface_hub import snapshot_download
from transformers import AutoConfig, AutoModelForCausalLM

from scxq2_cluster.ans import ans_decode, ans_encode

MAGIC_SHARD = b"SCXQ2S"
MAGIC_INDEX = "SCXQ2_INDEX"
VERSION = 1


@dataclass
class PackedTensor:
    name: str
    shape: tuple[int, ...]
    dtype: str
    scale: float
    data: bytes
    compressed: bool

    @property
    def payload_size(self) -> int:
        return len(self.data)


def _load_hf_state_dict(repo: str, revision: str | None, local_files_only: bool) -> tuple[dict[str, torch.Tensor], AutoConfig]:
    local_path = snapshot_download(
        repo_id=repo,
        revision=revision,
        local_files_only=local_files_only,
    )

    config = AutoConfig.from_pretrained(local_path, local_files_only=local_files_only)
    model = AutoModelForCausalLM.from_pretrained(
        local_path,
        torch_dtype=torch.float16,
        low_cpu_mem_usage=True,
        local_files_only=local_files_only,
    )
    return model.state_dict(), config


def _best_device() -> torch.device:
    if torch.cuda.is_available():
        return torch.device("cuda")
    if getattr(torch.backends, "mps", None) and torch.backends.mps.is_available():
        return torch.device("mps")
    return torch.device("cpu")


def _quantize_int4(t: torch.Tensor, device: torch.device) -> tuple[torch.Tensor, float]:
    tt = t.detach().to(device=device, dtype=torch.float32)
    max_val = torch.amax(torch.abs(tt))
    scale = 1.0 if max_val.item() == 0 else float(max_val.item() / 7.0)
    q = torch.round(tt / scale)
    q = torch.clamp(q, -8, 7).to(torch.int8)
    return q.to("cpu"), scale


def _pack_int4(values: torch.Tensor) -> bytes:
    flat = values.reshape(-1).tolist()
    packed = bytearray(math.ceil(len(flat) / 2))
    j = 0
    for i in range(0, len(flat), 2):
        lo = int(flat[i]) & 0x0F
        hi = (int(flat[i + 1]) & 0x0F) if i + 1 < len(flat) else 0
        packed[j] = (hi << 4) | lo
        j += 1
    return bytes(packed)


def _pack_tensor(name: str, tensor: torch.Tensor, use_ans: bool, device: torch.device) -> PackedTensor:
    q, scale = _quantize_int4(tensor, device=device)
    raw = _pack_int4(q)

    if use_ans:
        compressed, table = ans_encode(raw)
        # Keep raw if compression was ineffective.
        if len(compressed) + len(table) + 12 < len(raw):
            payload = struct.pack("<I", len(table)) + table + compressed
            return PackedTensor(name=name, shape=tuple(tensor.shape), dtype="int4", scale=scale, data=payload, compressed=True)

    return PackedTensor(name=name, shape=tuple(tensor.shape), dtype="int4", scale=scale, data=raw, compressed=False)


def _serialize_tensor(t: PackedTensor) -> bytes:
    name_b = t.name.encode("utf-8")
    shape_b = struct.pack("<I", len(t.shape)) + b"".join(struct.pack("<I", d) for d in t.shape)
    flags = 1 if t.compressed else 0
    header = struct.pack("<I I f I", len(name_b), flags, t.scale, len(t.data))
    return header + name_b + shape_b + t.data


def _parse_expert_id(name: str) -> int | None:
    patterns = [
        r"expert[s]?\.(\d+)",
        r"experts?_(\d+)",
        r"\.e(\d+)\.",
    ]
    for pat in patterns:
        m = re.search(pat, name)
        if m:
            return int(m.group(1))
    return None


def _split_into_shards(tensors: list[PackedTensor], max_shard_size: int) -> list[list[PackedTensor]]:
    shards: list[list[PackedTensor]] = []
    current: list[PackedTensor] = []
    size = 0

    for tensor in tensors:
        encoded_len = len(_serialize_tensor(tensor))
        if current and size + encoded_len > max_shard_size:
            shards.append(current)
            current = []
            size = 0
        current.append(tensor)
        size += encoded_len

    if current:
        shards.append(current)

    return shards


def _write_shard(path: Path, tensors: Iterable[PackedTensor]) -> list[str]:
    names: list[str] = []
    with path.open("wb") as f:
        f.write(MAGIC_SHARD)
        f.write(struct.pack("<I", VERSION))
        for tensor in tensors:
            names.append(tensor.name)
            f.write(_serialize_tensor(tensor))
    return names


def _write_metadata(output_dir: Path, config: AutoConfig, shard_files: list[str]) -> None:
    metadata = {
        "format": "scxq2",
        "hidden_size": getattr(config, "hidden_size", None),
        "num_attention_heads": getattr(config, "num_attention_heads", None),
        "num_hidden_layers": getattr(config, "num_hidden_layers", None),
        "shards": shard_files,
    }
    (output_dir / "model.transformers.json").write_text(json.dumps(metadata, indent=2), encoding="utf-8")


def convert_cluster(
    repo: str,
    output_dir: Path,
    revision: str | None,
    local_files_only: bool,
    max_shard_size_mb: int,
    use_ans: bool,
    profile: dict[str, Any] | None = None,
) -> None:
    state, config = _load_hf_state_dict(repo=repo, revision=revision, local_files_only=local_files_only)
    output_dir.mkdir(parents=True, exist_ok=True)

    device = _best_device()
    tensors: list[PackedTensor] = []
    experts: dict[int, list[PackedTensor]] = {}

    for name in sorted(state):
        t = state[name]
        if not isinstance(t, torch.Tensor) or not t.is_floating_point() or t.numel() == 0:
            continue

        packed = _pack_tensor(name=name, tensor=t.cpu(), use_ans=use_ans, device=device)
        expert_id = _parse_expert_id(name)
        if expert_id is None:
            tensors.append(packed)
        else:
            experts.setdefault(expert_id, []).append(packed)

    max_shard_size = max_shard_size_mb * 1024 * 1024
    regular_shards = _split_into_shards(tensors, max_shard_size=max_shard_size)

    shard_index: list[dict[str, object]] = []
    shard_files: list[str] = []

    for i, shard in enumerate(regular_shards):
        file_name = f"shard_{i:03d}.scxq2"
        tensor_names = _write_shard(output_dir / file_name, shard)
        shard_index.append({"file": file_name, "tensors": tensor_names})
        shard_files.append(file_name)

    for expert_id in sorted(experts):
        file_name = f"expert_{expert_id:03d}.scxq2"
        tensor_names = _write_shard(output_dir / file_name, experts[expert_id])
        shard_index.append({"file": file_name, "tensors": tensor_names, "expert_id": expert_id})
        shard_files.append(file_name)

    index = {
        "magic": MAGIC_INDEX,
        "version": VERSION,
        "hidden": getattr(config, "hidden_size", None),
        "layers": getattr(config, "num_hidden_layers", None),
        "experts": len(experts),
        "quantization": "int4",
        "compression": "ans" if use_ans else "none",
        "device_used": str(device),
        "shards": shard_index,
    }
    if profile:
        index["profile"] = profile
    (output_dir / "model.scxq2.index").write_text(json.dumps(index, indent=2), encoding="utf-8")
    _write_metadata(output_dir, config, shard_files)


def _load_profile(path: Path | None) -> dict[str, Any]:
    if path is None:
        return {}

    raw = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(raw, dict):
        raise ValueError("profile JSON must be an object")
    return raw


def _parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Cluster-aware HuggingFace to SCXQ2 converter")
    p.add_argument("repo", help="HuggingFace repo id, e.g. Qwen/Qwen2-7B")
    p.add_argument("output", type=Path, help="Output directory")
    p.add_argument("--revision", default=None)
    p.add_argument("--local-files-only", action="store_true")
    p.add_argument("--max-shard-size-mb", type=int, default=256)
    p.add_argument("--no-ans", action="store_true", help="Disable ANS entropy compression")
    p.add_argument("--profile", type=Path, default=None, help="Optional profile JSON for target-specific settings")
    return p.parse_args()


def main() -> None:
    args = _parse_args()
    profile = _load_profile(args.profile)

    max_shard_size_mb = int(profile.get("max_shard_size_mb", args.max_shard_size_mb)) if profile else args.max_shard_size_mb
    ans_enabled = not args.no_ans
    if profile:
        ans_enabled = bool(profile.get("entropy", {}).get("ans", ans_enabled))

    convert_cluster(
        repo=args.repo,
        output_dir=args.output,
        revision=args.revision,
        local_files_only=args.local_files_only,
        max_shard_size_mb=max_shard_size_mb,
        use_ans=ans_enabled,
        profile=profile or None,
    )


if __name__ == "__main__":
    main()
