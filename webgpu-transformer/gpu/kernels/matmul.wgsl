@group(0) @binding(0)
var<storage, read> A : array<f32>;

@group(0) @binding(1)
var<storage, read> B : array<f32>;

@group(0) @binding(2)
var<storage, read_write> C : array<f32>;

@compute @workgroup_size(16, 16)
fn main(@builtin(global_invocation_id) id : vec3<u32>) {
  let row = id.x;
  let col = id.y;

  var sum : f32 = 0.0;
  for (var k = 0u; k < 4096u; k++) {
    sum = sum + A[row * 4096u + k] * B[k * 4096u + col];
  }

  C[row * 4096u + col] = sum;
}
