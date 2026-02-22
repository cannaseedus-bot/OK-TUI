@group(0) @binding(0)
var<storage, read> Q : array<f32>;

@group(0) @binding(1)
var<storage, read> K : array<f32>;

@group(0) @binding(2)
var<storage, read> V : array<f32>;

@group(0) @binding(3)
var<storage, read_write> O : array<f32>;

@compute @workgroup_size(64)
fn main(@builtin(global_invocation_id) id : vec3<u32>) {
  let i = id.x;
  let score : f32 = Q[i] * K[i];
  O[i] = score * V[i];
}
