# WebGPU Transformer Runtime (SCXQ2)

This directory contains a browser-native prototype runtime for SCXQ2 model loading and WebGPU execution.

## Structure

- `index.html`, `main.js`: browser entry point.
- `scxq2/`: streaming decoder and GPU weight loader.
- `gpu/`: WebGPU initialization and WGSL kernels.
- `model/`: model loading, KV cache, and inference loop scaffolding.
- `tokenizer/`: minimal tokenizer and detokenizer.

## Run locally

Serve the repository root and open `/webgpu-transformer/index.html` in a WebGPU-enabled browser.

## Intel iGPU shard profile (SCXQ2)

The loader now supports SCXQ2 shard index manifests (`model.scxq2.index`) tuned for Intel HD/UHD iGPUs:

```js
const model = await loadModel(device, "/model/model.scxq2.index");
```

When loading an index, the runtime applies the `split_attn_mlp` strategy by default:

- attention shards are loaded on GPU (WebGPU)
- MLP shards are kept CPU-side for hybrid execution
- GPU residency is capped to a 512 MB safety budget

This aligns with Intel shared-memory behavior and helps avoid GPU memory fragmentation during streaming inference.
