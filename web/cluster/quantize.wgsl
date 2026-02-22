@group(0) @binding(0)
var<storage> input : array<f32>;

@group(0) @binding(1)
var<storage, read_write> output : array<i32>;

@group(0) @binding(2)
var<uniform> scale : f32;

@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) id : vec3<u32>) {
    let i = id.x;
    let q = clamp(round(input[i] / scale), -8.0, 7.0);
    output[i] = i32(q);
}
