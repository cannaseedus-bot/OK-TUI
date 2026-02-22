import { initGPU } from "./gpu/device.js";
import { loadModel } from "./model/loader.js";
import { infer } from "./model/inference.js";
import { tokenize, detokenize } from "./tokenizer/bpe.js";

let gpu;
let model;

async function init() {
  gpu = await initGPU();
  model = await loadModel(gpu, "/model/model.scxq2");
}

window.run = async function run() {
  if (!gpu || !model) {
    document.getElementById("output").textContent = "Model is still loading...";
    return;
  }

  const text = document.getElementById("input").value;
  const tokens = tokenize(text);
  const output = await infer(gpu, model, tokens);

  document.getElementById("output").textContent = detokenize(output);
};

init().catch((err) => {
  document.getElementById("output").textContent = `Initialization failed: ${err.message}`;
});
