import { KVCache } from "./kv_cache.js";

export async function infer(device, model, tokens) {
  let state = tokens;
  const kv = new KVCache(device, 1024 * 1024);

  for (let layer = 0; layer < model.layers; layer += 1) {
    state = await transformerLayer(device, model, state, kv, layer);
  }

  return state;
}

async function transformerLayer(device, model, state, kv, layer) {
  const q = matmul(device, state, model.weights[`q_${layer}`]);
  const k = matmul(device, state, model.weights[`k_${layer}`]);
  const v = matmul(device, state, model.weights[`v_${layer}`]);

  kv.last = { q, k, v, layer };
  return attention(device, q, k, v);
}

function matmul(_device, state, _weight) {
  return state;
}

function attention(_device, q, _k, _v) {
  return q;
}
