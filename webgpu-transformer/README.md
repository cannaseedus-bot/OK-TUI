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
