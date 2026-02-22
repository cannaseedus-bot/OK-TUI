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
