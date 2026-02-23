# iGPU WebGPU Support Plan - Phase 3 Extension

**Status**: Planned for Phase 3 (Q2 2026)
**Platform**: Windows 10/11 with iGPU (Intel UHD, AMD Radeon Graphics)
**API**: WebGPU (cross-platform GPU abstraction)
**Performance Target**: 2-5x speedup on iGPU inference

---

## Overview

Add integrated GPU (iGPU) support via WebGPU to Ollama-K Windows port, enabling GPU acceleration on systems without discrete GPUs. This extends Phase 3 MLIR/LLVM compiler with GPU codegen capabilities.

---

## 1. Architecture

### Current State (v1.0.0)
```
K'UHUL Source
    ↓
Compiler (7 phases) → JavaScript
    ↓
Runtime (5-tiers)
    ↓
CPU Execution Only
```

### With iGPU/WebGPU Support (v1.1+)
```
K'UHUL Source
    ↓
Compiler (9 phases)
├─ Phase 1-7: Original pipeline
├─ Phase 8: MLIR → GPU IR (WebGPU WGSL)
└─ Phase 9: WebGPU binding generation
    ↓
Dual-Path Runtime
├─ CPU Path: JavaScript (fallback)
└─ GPU Path: WebGPU kernels + JavaScript bridge
    ↓
Execution Strategy
├─ iGPU available → GPU path
├─ GPU unavailable → CPU fallback
└─ Hybrid: Distribute work based on cost model
```

---

## 2. WebGPU Integration Points

### 2.1 Compiler Enhancement

**New Phase 8: MLIR → GPU IR**
```go
// kuhul/compiler/gpu_codegen.go (NEW)
type GPUCodeGenerator struct {
    MLIRContext  *mlir.Context
    Module       mlir.Module
    TargetGPU    GPUTarget  // IGPU, DGPU, CPU
}

// Convert MLIR to WebGPU WGSL
func (gcg *GPUCodeGenerator) LowerToWGSL() (string, error) {
    // MLIR ops → WebGPU WGSL
    // Linear algebra → compute shaders
    // Memory layout → GPU buffers
}

// GPU kernel generation
func (gcg *GPUCodeGenerator) GenerateKernels(mlirOps []mlir.Operation) []Kernel {
    // One kernel per vectorizable operation
    // Thread group sizing for iGPU capabilities
}
```

**New Phase 9: WebGPU Binding**
```go
// kuhul/compiler/webgpu_bindings.go (NEW)
type WebGPUBinding struct {
    Kernels      []WGSLKernel
    Buffers      []BufferLayout
    Uniforms     []UniformVariable
    PipelineInfo PipelineMetadata
}

func (wb *WebGPUBinding) GenerateJSBridge() string {
    // Generate JavaScript code that:
    // 1. Detects WebGPU support
    // 2. Creates GPU device/queue
    // 3. Compiles WGSL shaders
    // 4. Manages buffer allocation
    // 5. Dispatches compute work
}
```

### 2.2 Runtime GPU Tier

**New @gpu Tier**
```go
// kuhul/runtime/gpu_tier.go (NEW)
type GPUTier struct {
    Adapter          *wgpu.Adapter
    Device           *wgpu.Device
    Queue            *wgpu.Queue
    ComputePipeline  *wgpu.ComputePipeline
    Buffers          map[string]*wgpu.Buffer
    Textures         map[string]*wgpu.Texture
}

// Operations
func (gt *GPUTier) AllocateBuffer(size uint64, usage wgpu.BufferUsage) (*wgpu.Buffer, error)
func (gt *GPUTier) CopyToGPU(buffer *wgpu.Buffer, data []byte) error
func (gt *GPUTier) DispatchCompute(x, y, z uint32) error
func (gt *GPUTier) CopyFromGPU(buffer *wgpu.Buffer) ([]byte, error)
func (gt *GPUTier) CreatePipeline(wgsl string) (*wgpu.ComputePipeline, error)
```

### 2.3 Hybrid Execution Layer

**Cost-Based Kernel Selection**
```go
// kuhul/runtime/hybrid_executor.go (NEW)
type HybridExecutor struct {
    CPUPath  func() interface{}    // Fallback
    GPUPath  func() interface{}    // iGPU accelerated
    CostModel CostEstimator
}

func (he *HybridExecutor) Execute(ctx context.Context, op Operation) interface{} {
    cost_cpu := he.CostModel.EstimateCPU(op)      // Estimated CPU time
    cost_gpu := he.CostModel.EstimateGPU(op)      // Estimated GPU time

    if cost_gpu < cost_cpu {
        return he.GPUPath()  // Use GPU
    } else {
        return he.CPUPath()  // Use CPU fallback
    }
}
```

---

## 3. iGPU Detection & Capabilities

### 3.1 Platform Detection

**Windows iGPU Detection**
```go
// kuhul/runtime/gpu_detection.go (NEW)
type GPUCapabilities struct {
    HasWebGPU          bool
    GPUType            string          // "igpu", "dgpu", "none"
    VendorID           uint32          // Intel, AMD, NVIDIA
    Compute Capability string          // "gen12", "rdna2", etc.
    MaxComputeWorkgroups Limits
    MaxBufferSize      uint64
    MaxTextureSize     uint32
}

func DetectGPU(ctx context.Context) (*GPUCapabilities, error) {
    // Windows: Check WDDM 3.0+ support
    // Detect iGPU via DXGI or Direct3D 12
    // Get WebGPU adapter info
}

func (gc *GPUCapabilities) IsIGPU() bool {
    return gc.GPUType == "igpu"
}

func (gc *GPUCapabilities) GetOptimalWorkgroupSize() (x, y, z uint32) {
    // Return optimal work group size for detected iGPU
    // Intel iGPU: 8x8 typical
    // AMD iGPU: 8x8 or 16x4
}
```

### 3.2 iGPU-Specific Tuning

**Intel UHD Graphics (12th Gen+)**
- Max compute work groups: 2048
- Wave size: 8
- Shared memory per workgroup: 96KB
- Best work group size: 8x8 or 16x4

**AMD Radeon Graphics (RDNA)**
- Max compute work groups: 1024
- Wave size: 64
- Shared memory per workgroup: 96KB
- Best work group size: 64x1 or 32x2

---

## 4. WebGPU WGSL Kernel Examples

### 4.1 Matrix Multiplication (iGPU-optimized)

```wgsl
// matrix_multiply.wgsl
@group(0) @binding(0) var<storage, read> A: array<f32>;
@group(0) @binding(1) var<storage, read> B: array<f32>;
@group(0) @binding(2) var<storage, read_write> C: array<f32>;

@compute @workgroup_size(8, 8, 1)
fn matmul_tile(
    @builtin(global_invocation_id) global_id: vec3<u32>,
    @builtin(local_invocation_id) local_id: vec3<u32>,
    @builtin(workgroup_id) workgroup_id: vec3<u32>,
) {
    let i = global_id.x;
    let j = global_id.y;
    let cols = 256u; // M dimension

    var sum = 0.0;
    for (var k = 0u; k < cols; k = k + 1u) {
        sum += A[i * cols + k] * B[k * cols + j];
    }

    C[i * cols + j] = sum;
}
```

### 4.2 Tensor Convolution

```wgsl
// tensor_conv.wgsl
@group(0) @binding(0) var<storage, read> input_tensor: array<f32>;
@group(0) @binding(1) var<storage, read> kernel: array<f32>;
@group(0) @binding(2) var<storage, read_write> output: array<f32>;

@compute @workgroup_size(8, 8, 1)
fn conv2d(
    @builtin(global_invocation_id) global_id: vec3<u32>,
) {
    let out_x = global_id.x;
    let out_y = global_id.y;
    let out_idx = out_y * 256u + out_x;

    var acc = 0.0;
    for (var ky = 0u; ky < 3u; ky = ky + 1u) {
        for (var kx = 0u; kx < 3u; kx = kx + 1u) {
            let in_x = out_x + kx;
            let in_y = out_y + ky;
            let in_idx = in_y * 258u + in_x;
            let k_idx = ky * 3u + kx;
            acc += input_tensor[in_idx] * kernel[k_idx];
        }
    }

    output[out_idx] = acc;
}
```

### 4.3 Layer Normalization

```wgsl
// layer_norm.wgsl
@group(0) @binding(0) var<storage, read> input: array<f32>;
@group(0) @binding(1) var<storage, read_write> output: array<f32>;

var<workgroup> shared_sum: f32;
var<workgroup> shared_variance: f32;

@compute @workgroup_size(256, 1, 1)
fn layer_norm(
    @builtin(global_invocation_id) global_id: vec3<u32>,
    @builtin(local_invocation_id) local_id: vec3<u32>,
) {
    let idx = global_id.x;
    let local_idx = local_id.x;

    // Parallel reduction for mean
    let value = input[idx];
    shared_sum = value; // Simplified - full impl uses atomic ops
    workgroupBarrier();

    let mean = shared_sum / 256.0;
    let diff = value - mean;

    // Parallel reduction for variance
    shared_variance = diff * diff;
    workgroupBarrier();

    let variance = shared_variance / 256.0;
    let normalized = diff / sqrt(variance + 1e-6);

    output[idx] = normalized;
}
```

---

## 5. Implementation Phases

### Phase 3.1: WebGPU Foundation (Weeks 1-2)
- [ ] Add WebGPU/WGPU Go bindings
- [ ] Implement GPU detection for Windows iGPU
- [ ] Create basic GPU tier with buffer management
- [ ] Add simple compute shader compilation

**Files**:
- `kuhul/runtime/gpu_tier.go` (300 lines)
- `kuhul/runtime/gpu_detection.go` (200 lines)
- `kuhul/gpu/webgpu_bindings.go` (400 lines)

### Phase 3.2: Compiler GPU Codegen (Weeks 3-4)
- [ ] Implement MLIR → WGSL lowering
- [ ] Add kernel generation from MLIR ops
- [ ] Create WebGPU binding generator
- [ ] Generate JavaScript GPU bridge

**Files**:
- `kuhul/compiler/gpu_codegen.go` (500 lines)
- `kuhul/compiler/wgsl_generator.go` (400 lines)
- `kuhul/compiler/gpu_bindings.go` (300 lines)

### Phase 3.3: Hybrid Execution (Weeks 5-6)
- [ ] Implement cost model for kernel selection
- [ ] Add hybrid executor
- [ ] CPU↔GPU data transfer optimization
- [ ] Fallback mechanisms

**Files**:
- `kuhul/runtime/hybrid_executor.go` (400 lines)
- `kuhul/runtime/cost_model.go` (300 lines)

### Phase 3.4: Optimization (Weeks 7-8)
- [ ] GPU memory pooling
- [ ] Kernel fusion
- [ ] Workgroup size auto-tuning
- [ ] Shared memory utilization

**Files**:
- `kuhul/gpu/memory_pool.go` (200 lines)
- `kuhul/gpu/kernel_fusion.go` (300 lines)

### Phase 3.5: Testing & Benchmarking (Weeks 9-10)
- [ ] GPU unit tests
- [ ] Performance benchmarks (iGPU vs CPU)
- [ ] Stress testing
- [ ] Documentation

**Files**:
- `kuhul/gpu/gpu_test.go` (500 lines)
- `kuhul/gpu/benchmarks_gpu_test.go` (400 lines)

---

## 6. Performance Expectations

### iGPU Speedups (vs CPU)

| Operation | Size | CPU Time | GPU Time | Speedup |
|-----------|------|----------|----------|---------|
| Matrix Mult | 256x256 | 5.2ms | 1.8ms | 2.9x |
| Matrix Mult | 1024x1024 | 85ms | 22ms | 3.9x |
| Conv2D | 3x3 kernel | 12ms | 3.5ms | 3.4x |
| LayerNorm | 4096 elements | 2.1ms | 0.8ms | 2.6x |
| Attention | 64x64 | 8.5ms | 2.2ms | 3.9x |

### Real-World Inference

```
Llama2-7B on iGPU (Intel 12th Gen):
├─ CPU-only:      ~500ms per token
├─ iGPU-only:     ~180ms per token (2.8x faster)
└─ Hybrid (adaptive): ~160ms per token (3.1x faster)
```

---

## 7. Fallback Strategy

### Automatic Fallback
```go
// GPU kernel fails → CPU fallback
if err := gpuPath(); err != nil {
    logger.Warn("GPU path failed, falling back to CPU", "error", err)
    return cpuPath()
}
```

### Manual Fallback Triggers
```powershell
# Environment variables to control GPU usage
$env:OLLAMA_GPU = "auto"    # Automatic selection (default)
$env:OLLAMA_GPU = "igpu"    # Force iGPU
$env:OLLAMA_GPU = "cpu"     # Force CPU only
$env:OLLAMA_GPU = "hybrid"  # Cost-based hybrid
```

---

## 8. Dependencies

### Go Libraries
- `wgpu-go`: WebGPU bindings (~500KB)
- `mlir-go`: MLIR bindings (~1MB, optional)
- `spirv-go`: SPIR-V tools (~200KB)

### System Requirements
- Windows 10 21H2+ or Windows 11
- iGPU driver with WDDM 2.1+
- WebGPU runtime (bundled or via Edge)

---

## 9. Optimization Opportunities

### 1. Kernel Fusion
```go
// Fuse multiple operations into single kernel
// [MatMul] → [Add] → [Relu] → [LayerNorm]
// Becomes: single_fused_kernel()
```

### 2. Memory Coalescing
```go
// Rearrange buffer access patterns for iGPU cache efficiency
// Row-major → Tiled layout
```

### 3. Workgroup Tuning
```go
// Auto-tune workgroup sizes per iGPU model
// Intel: 8x8, AMD: 16x4
```

### 4. Register Pressure
```go
// Balance register usage vs shared memory
// Typical iGPU: 128 registers per thread
```

---

## 10. Testing Strategy

### Unit Tests (GPU Tier)
```go
TestGPUBufferAllocation
TestGPUBufferCopy
TestGPUKernelCompile
TestGPUComputeDispatch
TestGPUFallback
```

### Integration Tests
```go
TestE2EMatrixMultiplicationGPU
TestE2EConvolutionGPU
TestE2ELLMInferenceGPU
TestGPUCPUConsistency
```

### Performance Benchmarks
```bash
go test ./kuhul/gpu -bench=. -benchmem -benchtime=10s
```

---

## 11. Rollout Plan

### v1.1 (April 2026) - iGPU Foundation
- WebGPU support for Intel/AMD iGPU
- Basic matrix operations on GPU
- 2-3x speedup on supported operations

### v1.2 (May 2026) - Full GPU Stack
- MLIR → GPU codegen
- Kernel fusion and optimization
- 3-5x speedup on complex operations

### v1.3 (June 2026) - Production Hardening
- Extended testing on diverse iGPU models
- Memory optimization
- Production-ready stability

---

## 12. Documentation

### For Users
- **iGPU Setup Guide**: Enable/disable GPU acceleration
- **Performance Tuning**: Workgroup sizing, memory management
- **Troubleshooting**: GPU detection issues, fallback handling

### For Developers
- **GPU Kernel Development**: Writing WGSL kernels
- **GPU Compiler Extension**: Adding GPU codegen to compiler
- **Hybrid Execution**: Cost models and kernel selection

---

## 13. Success Criteria

✅ **Functional**
- [ ] WebGPU kernels execute correctly on Intel iGPU
- [ ] WebGPU kernels execute correctly on AMD iGPU
- [ ] CPU fallback works seamlessly
- [ ] Hybrid execution selects optimal path

✅ **Performance**
- [ ] 2.5x+ speedup on matrix operations
- [ ] 3x+ speedup on convolutions
- [ ] < 1% performance overhead for CPU fallback

✅ **Reliability**
- [ ] 100% test pass rate
- [ ] Zero memory leaks (GPU and CPU)
- [ ] Handles edge cases (OOM, GPU disconnect)

✅ **Usability**
- [ ] Automatic GPU detection
- [ ] No user configuration required
- [ ] Clear logging and diagnostics

---

## 14. Future Extensions

### CUDA Support (v2.0)
- NVIDIA discrete GPUs
- 5-20x speedup vs iGPU
- CUDA kernel generation from MLIR

### Vulkan Support (v2.0)
- Alternative to WebGPU
- Better Linux/macOS support

### Multi-GPU (v2.1)
- Workload distribution across multiple GPUs
- 2-8x speedup with 2-8 GPUs

---

## Summary

iGPU WebGPU support brings:
- ✅ 2-5x inference speedup on integrated GPUs
- ✅ No discrete GPU required
- ✅ Seamless CPU fallback
- ✅ Cross-platform (Windows → Linux/macOS)
- ✅ Production-ready for Q2 2026

**Estimated Effort**: 10 weeks
**Expected Release**: May-June 2026
**Performance Gain**: 2.8-3.9x on typical iGPU

---

**Status**: Planned for Phase 3.2
**Priority**: High (enables GPU on most Windows systems)
**Target Release**: v1.1 (April 2026)
