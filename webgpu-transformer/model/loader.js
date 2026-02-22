import { loadWeightsGPU } from "../scxq2/gpu_loader.js";

export async function loadModel(device, url) {
  const weights = await loadWeightsGPU(device, url);

  return {
    weights,
    hidden: 4096,
    layers: 16,
  };
}
