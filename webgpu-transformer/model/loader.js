import { loadSCXQ2Model, loadWeightsGPU } from "../scxq2/gpu_loader.js";

export async function loadModel(device, url) {
  const isShardIndex = url.endsWith(".index");
  const shardModel = isShardIndex
    ? await loadSCXQ2Model(device, {
        index: url,
        gpu_limit_mb: 512,
        shard_strategy: "split_attn_mlp",
        hybrid_execution: true,
      })
    : null;
  const weights = isShardIndex
    ? shardModel.gpuWeights
    : await loadWeightsGPU(device, url);

  return {
    weights,
    cpuWeights: shardModel?.cpuWeights ?? {},
    profile: shardModel?.profile ?? "default",
    hidden: 4096,
    layers: 16,
  };
}
