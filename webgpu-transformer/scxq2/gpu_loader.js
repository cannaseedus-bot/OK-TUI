import { decodeSCXQ2Stream } from "./stream.js";

export async function loadWeightsGPU(device, url) {
  const weights = {};

  for await (const node of decodeSCXQ2Stream(url)) {
    if (node.type !== "tensor") {
      continue;
    }

    const buffer = device.createBuffer({
      size: node.data.byteLength,
      usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_DST,
    });

    device.queue.writeBuffer(buffer, 0, node.data);
    weights[node.name] = buffer;
  }

  return weights;
}

async function loadIndex(indexUrl) {
  const res = await fetch(indexUrl);
  if (!res.ok) {
    throw new Error(`Unable to fetch SCXQ2 index: ${res.status}`);
  }

  return res.json();
}

function resolveShardURL(indexUrl, shardFile) {
  return new URL(shardFile, indexUrl).toString();
}

export async function loadSCXQ2Model(
  device,
  {
    index,
    gpu_limit_mb: gpuLimitMB = 512,
    shard_strategy: shardStrategy = "split_attn_mlp",
    hybrid_execution: hybridExecution = true,
  },
) {
  const manifest = await loadIndex(index);
  const gpuWeights = {};
  const cpuWeights = {};
  let gpuBytes = 0;

  const profile = manifest.profile ?? "default";
  const maxResidentBytes = Math.min(
    (manifest.execution?.max_gpu_resident_mb ?? gpuLimitMB) * 1024 * 1024,
    gpuLimitMB * 1024 * 1024,
  );

  for (const shard of manifest.shards ?? []) {
    const shardUrl = resolveShardURL(index, shard.file);

    // In split-attn-mlp mode we prefer attention tensors on GPU and keep MLP tensors CPU-side.
    const shouldLoadOnGPU =
      !hybridExecution || shardStrategy !== "split_attn_mlp" || shard.type !== "mlp";

    if (!shouldLoadOnGPU) {
      const res = await fetch(shardUrl);
      if (!res.ok) {
        throw new Error(`Unable to fetch CPU shard ${shard.file}: ${res.status}`);
      }

      cpuWeights[shard.file] = new Uint8Array(await res.arrayBuffer());
      continue;
    }

    const shardWeights = await loadWeightsGPU(device, shardUrl);
    const shardBytes = Object.values(shardWeights).reduce(
      (total, buffer) => total + buffer.size,
      0,
    );

    if (gpuBytes + shardBytes > maxResidentBytes) {
      throw new Error(
        `GPU shard budget exceeded while loading ${shard.file}. ` +
          `Used ${Math.round(gpuBytes / (1024 * 1024))} MB of ${Math.round(maxResidentBytes / (1024 * 1024))} MB.`,
      );
    }

    gpuBytes += shardBytes;
    gpuWeights[shard.file] = shardWeights;
  }

  return {
    profile,
    manifest,
    gpuWeights,
    cpuWeights,
    gpuResidentMB: Math.round(gpuBytes / (1024 * 1024)),
  };
}
