# SCXQ2 Packer

Minimal SCXQ2 packer pipeline for exporting PyTorch-compatible checkpoints to a stream-friendly `.scxq2` container.

## Features

- Quantization modes: `int4`, `int8`, `fp16`
- Deterministic tensor ordering (sorted by tensor name)
- Lane-oriented tensor blocks with shape metadata
- File-level Merkle root over all tensor lanes
- JSON metadata footer for runtime inspection

## Supported Inputs

- `.safetensors`
- `.bin`, `.pt`, `.pth` (PyTorch checkpoints)
- `.npy` (single tensor)

> `.gguf` is not parsed by this script yet.

## Usage

```bash
python scxq2-packer/pack.py /path/to/model.safetensors scxq2-packer/output/model.scxq2 --quantization int4
```

## Container Layout

```text
magic      5 bytes  "SCXQ2"
version    u32
merkle     32 bytes SHA256 Merkle root over lane blobs
tensor_n   u32

repeat tensor_n:
  lane_len u32
  lane bytes

meta_len   u32
meta_json  bytes
```

Each lane is:

```text
name_len   u32
tensor_len u32
scale      f32
format_id  u32  (1=int4, 2=int8, 3=fp16)
rank       u32
name       bytes
shape      rank * u32
payload    bytes
```

## HuggingFace one-command conversion

You can convert directly from a HuggingFace repo without manually downloading checkpoint files:

```bash
python hf_to_scxq2.py Qwen/Qwen2-7B output/qwen2-7b.scxq2 --int4
python hf_to_scxq2.py mistralai/Mistral-7B-v0.1 output/mistral.scxq2 --int8
python hf_to_scxq2.py microsoft/phi-2 output/phi2.scxq2 --fp16
```

Install dependencies:

```bash
pip install torch transformers huggingface_hub tqdm numpy
```
