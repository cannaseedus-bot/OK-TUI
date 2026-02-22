@group(0) @binding(0)
var<storage, read> X : array<f32>;

@group(0) @binding(1)
var<storage, read_write> Y : array<f32>;

@compute @workgroup_size(64)
fn main(@builtin(global_invocation_id) id : vec3<u32>) {
  let i = id.x;
  Y[i] = exp(X[i]);
}
