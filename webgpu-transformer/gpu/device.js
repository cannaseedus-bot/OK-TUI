export async function initGPU() {
  if (!navigator.gpu) {
    throw new Error("WebGPU is not supported in this browser.");
  }

  const adapter = await navigator.gpu.requestAdapter();
  if (!adapter) {
    throw new Error("Failed to acquire WebGPU adapter.");
  }

  return adapter.requestDevice();
}
