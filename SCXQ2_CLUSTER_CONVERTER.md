# SCXQ2 Cluster Converter

Use `hf_to_scxq2_cluster.py` to export a HuggingFace CausalLM model into
cluster-ready SCXQ2 shards.

## Features

- INT4 quantization with GPU acceleration when CUDA/MPS is available.
- Entropy compression layer exposed as ANS API (`scxq2_cluster/ans.py`).
- Automatic shard splitting with configurable max shard size.
- MoE expert isolation into `expert_XXX.scxq2` files.
- `model.transformers.json` compatibility metadata for Transformers.js loaders.

## Usage

```bash
python hf_to_scxq2_cluster.py Qwen/Qwen2-7B output/
```

Optional flags:

- `--max-shard-size-mb 256`
- `--no-ans`
- `--revision <tag_or_commit>`
- `--local-files-only`

## Output

- `model.scxq2.index`
- `model.transformers.json`
- `shard_000.scxq2`, `shard_001.scxq2`, ...
- `expert_000.scxq2`, ... (for MoE models)
