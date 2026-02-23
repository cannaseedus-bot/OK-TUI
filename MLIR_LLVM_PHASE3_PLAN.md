# Phase 3: MLIR/LLVM Compiler Layer - Implementation Guide

**Status**: Design & Planning Phase
**Estimated Duration**: 2-3 weeks implementation
**Target**: JIT compilation, native code generation, GPU acceleration

---

## Overview

Phase 3 introduces a compiler layer that transforms K'UHUL code into native machine code through MLIR (Multi-Level Intermediate Representation) and LLVM (Low-Level Virtual Machine). This enables:

- **JIT Compilation**: Compile K'UHUL code at runtime for immediate execution
- **Native Code Generation**: Produce platform-specific optimized binaries
- **GPU Acceleration**: Target NVIDIA (CUDA), AMD (HIP), Intel (XPU) via MLIR dialects
- **Self-Hosting**: K'UHUL compiler can compile itself (written in K'UHUL)
- **Optimization Passes**: Dead code elimination, constant propagation, vectorization

---

## Architecture Overview

```
K'UHUL Source Code
    ↓
K'UHUL Lexer/Parser
    ↓
K'UHUL AST (Abstract Syntax Tree)
    ↓
┌─────────────────────────────────────┐
│  MLIR Compiler (Phase 3)            │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ AST → K'UHUL MLIR Dialect     │ │
│  └────┬─────────────────────────┘ │
│       ↓                            │
│  ┌────────────────────────────────┐ │
│  │ K'UHUL Ops Lowering           │ │
│  │ kuhul.pop → std.func           │ │
│  │ kuhul.wo → memref.alloca       │ │
│  │ kuhul.sek → cf.br/cf.cond_br   │ │
│  └────┬─────────────────────────┘ │
│       ↓                            │
│  ┌────────────────────────────────┐ │
│  │ MLIR Standard Dialect          │ │
│  └────┬─────────────────────────┘ │
│       ↓                            │
│  ┌────────────────────────────────┐ │
│  │ Optimization Passes            │ │
│  │ - Dead code elimination         │ │
│  │ - Constant propagation          │ │
│  │ - Loop unrolling                │ │
│  │ - Vectorization                 │ │
│  └────┬─────────────────────────┘ │
│       ↓                            │
│  ┌────────────────────────────────┐ │
│  │ LLVM IR Lowering               │ │
│  │ std → llvm dialect              │ │
│  └────┬─────────────────────────┘ │
└─────┬──────────────────────────────┘
      ↓
  LLVM IR (.ll)
      ↓
  LLVM Backend
      ↓
┌──────────┬──────────┬──────────┐
│ CPU Code │GPU Code  │ WASM     │
│ (x86-64) │ (CUDA)   │ (Web)    │
└──────────┴──────────┴──────────┘
```

---

## Phase 3.1: MLIR K'UHUL Dialect

### 3.1.1 Directory Structure

```
/mlir/
  ├── include/kuhul/
  │   ├── KuhulDialect.h         (C++ dialect definition)
  │   ├── KuhulOps.h             (C++ operation definitions)
  │   ├── KuhulOps.td            (TableGen operation specs)
  │   └── KuhulTypes.h           (Custom type definitions)
  │
  └── lib/
      ├── KuhulDialect.cpp       (Dialect implementation)
      ├── KuhulOps.cpp           (Operation implementation)
      └── KuhulTypes.cpp         (Type system implementation)

/kuhul/compiler/
  ├── mlir_compiler.go           (Go bindings to MLIR)
  ├── mlir_compiler_test.go      (MLIR compiler tests)
  └── codegen.go                 (Code generation utilities)
```

### 3.1.2 K'UHUL MLIR Operations

Map each K'UHUL language construct to an MLIR operation:

```
K'UHUL Syntax         →  MLIR Op Name      →  LLVM Lowering
─────────────────────────────────────────────────────────
pop identifier        →  kuhul.pop         →  llvm.func
  (body)                (function)

wo binding            →  kuhul.wo          →  memref.alloca
  = value                (variable)

sek condition         →  kuhul.sek         →  cf.br / cf.cond_br
  (true)              (control flow)
  (false)

xul (block)           →  kuhul.xul         →  llvm.bb (basic block)
  statements          (block)

chen value            →  kuhul.chen        →  llvm.return
(return)              (return)

plen value            →  kuhul.plen        →  arith.constant
(constant)            (constant)

qa operand            →  kuhul.qa          →  std.call
call                  (call)

nep [values]          →  kuhul.nep         →  memref.alloc
array literal         (array)

sulu { ... }          →  kuhul.sulu        →  llvm.struct
object literal        (object)

qesh condition        →  kuhul.qesh        →  scf.if / scf.while
  (body)              (conditional)

xar array             →  kuhul.xar         →  scf.for
  (body)              (iteration)
```

### 3.1.3 Custom Type System

Define types that match K'UHUL's type system:

```cpp
// In KuhulTypes.h and KuhulOps.td

class K'UHULType : public mlir::Type {
    kuhul_i32       // 32-bit integer
    kuhul_f64       // 64-bit float
    kuhul_string    // String type
    kuhul_array     // Array with element type
    kuhul_object    // Object/struct type
    kuhul_pack      // Pack reference (external module)
    kuhul_function  // Function type
    kuhul_any       // Dynamic type (fallback)
};

// Type examples:
// array<i32>       → Array of 32-bit integers
// object<name, i32, value, f64> → Object with fields
// function<(i32, f64) -> (i32)> → Function signature
```

### 3.1.4 Tablegen Operation Definitions

Create `KuhulOps.td` with MLIR TableGen format:

```tablegen
// kuhul.pop - Function definition
def KuhulPop : Kuhul_Op<"pop", [IsIsolatedFromAbove]> {
  let arguments = (ins SymbolNameAttr:$sym_name,
                       TypeAttrOf<FunctionType>:$type);
  let regions = (region AnyRegion:$body);
  let results = (outs);
  let printer = [{ /* custom printer */ }];
  let parser = [{ /* custom parser */ }];
}

// kuhul.wo - Variable binding
def KuhulWo : Kuhul_Op<"wo"> {
  let arguments = (ins SymbolNameAttr:$name,
                       AnyType:$value);
  let results = (outs);
  let hasFolder = 1;
}

// kuhul.sek - Control flow
def KuhulSek : Kuhul_Op<"sek", [Terminator]> {
  let arguments = (ins I1:$condition);
  let successors = (successor AnySuccessor:$trueBB,
                                AnySuccessor:$falseBB);
}

// ... more operations
```

---

## Phase 3.2: K'UHUL to MLIR Compiler

### 3.2.1 Go Bindings via CGo

Create `mlir_compiler.go`:

```go
package compiler

/*
#cgo CXXFLAGS: -fPIC -std=c++17
#cgo LDFLAGS: -lMLIR -lMLIRAnalysis -lMLIRDialect

#include "kuhul/mlir_compiler.h"

typedef struct {
    void* module;
    char* error_msg;
} CompileResult;
*/
import "C"

type MLIRModule struct {
    ptr *C.void
}

type MLIRCompiler struct {
    context *C.void
    builder *C.void
}

// NewMLIRCompiler creates a new MLIR compiler instance
func NewMLIRCompiler() (*MLIRCompiler, error) {
    ctx := C.CreateMLIRContext()
    if ctx == nil {
        return nil, fmt.Errorf("failed to create MLIR context")
    }

    return &MLIRCompiler{context: ctx}, nil
}

// CompileAST compiles a K'UHUL AST to MLIR module
func (c *MLIRCompiler) CompileAST(ast *kuhul.Program) (*MLIRModule, error) {
    // Convert K'UHUL AST to MLIR operations
    // This would call C++ functions

    result := C.CompileKuhulAST(
        c.context,
        c.builder,
        (*C.KuhulASTNode)(unsafe.Pointer(ast.Ptr())),
    )

    if result.error_msg != nil {
        msg := C.GoString(result.error_msg)
        C.free(unsafe.Pointer(result.error_msg))
        return nil, fmt.Errorf("compilation failed: %s", msg)
    }

    return &MLIRModule{ptr: result.module}, nil
}

// EmitLLVMIR emits LLVM IR from MLIR module
func (m *MLIRModule) EmitLLVMIR() (string, error) {
    cStr := C.EmitLLVMIR(m.ptr)
    if cStr == nil {
        return "", fmt.Errorf("IR emission failed")
    }
    defer C.free(unsafe.Pointer(cStr))
    return C.GoString(cStr), nil
}

// Optimize applies optimization passes
func (m *MLIRModule) Optimize(level int) error {
    success := C.ApplyOptimizations(m.ptr, C.int(level))
    if !success {
        return fmt.Errorf("optimization failed at level %d", level)
    }
    return nil
}

// Compile to native code
func (m *MLIRModule) CompileToNative(target string) ([]byte, error) {
    cTarget := C.CString(target)
    defer C.free(unsafe.Pointer(cTarget))

    result := C.CompileToNative(m.ptr, cTarget)
    if result.error_msg != nil {
        msg := C.GoString(result.error_msg)
        C.free(unsafe.Pointer(result.error_msg))
        return nil, fmt.Errorf("native compilation failed: %s", msg)
    }

    // Convert C byte array to Go slice
    // ...
    return nativeCode, nil
}
```

### 3.2.2 C++ MLIR Compiler Implementation

Create `mlir_lib/KuhulCompiler.cpp`:

```cpp
#include "kuhul/KuhulCompiler.h"
#include "mlir/IR/MLIRContext.h"
#include "mlir/IR/Module.h"
#include "mlir/Parser/Parser.h"
#include "mlir/Pass/PassManager.h"
#include "mlir/Conversion/ReconcileUnrealizedCasts/ReconcileUnrealizedCasts.h"
#include "llvm/Support/SourceMgr.h"

using namespace mlir;
using namespace kuhul;

class KuhulCompilerImpl {
public:
    MLIRContext context;
    OpBuilder builder{&context};
    ModuleOp module;

    KuhulCompilerImpl() {
        // Register K'UHUL dialect
        context.loadDialect<KuhulDialect>();
        // Register standard dialect
        context.loadDialect<arith::ArithmeticDialect>();
        // ... other dialects
    }

    OwningModuleRef compileAST(const KuhulAST& ast) {
        Location loc = UnknownLoc::get(&context);

        // Create module
        module = builder.create<ModuleOp>(loc);

        // Compile top-level constructs
        for (const auto& stmt : ast.statements) {
            compileStatement(stmt);
        }

        return std::move(module);
    }

    void compileStatement(const Statement& stmt) {
        if (auto* funcDef = std::get_if<FunctionDef>(&stmt)) {
            compileFunctionDef(*funcDef);
        } else if (auto* varDef = std::get_if<VarDef>(&stmt)) {
            compileVarDef(*varDef);
        }
        // ... more statement types
    }

    void compileFunctionDef(const FunctionDef& funcDef) {
        Location loc = UnknownLoc::get(&context);

        // Create function type
        auto funcType = builder.getFunctionType(
            getMLIRTypes(funcDef.paramTypes),
            getMLIRTypes(funcDef.returnTypes)
        );

        // Create function operation
        auto func = builder.create<FuncOp>(
            loc,
            funcDef.name,
            funcType,
            ArrayRef<NamedAttribute>{}
        );

        // Create function body
        auto* entryBlock = func.addEntryBlock();
        builder.setInsertionPointToStart(entryBlock);

        // Compile function body
        compileBlock(funcDef.body);

        // Add return if needed
        if (funcDef.returnTypes.empty()) {
            builder.create<ReturnOp>(loc);
        }
    }
};

extern "C" {
    CompileResult CompileKuhulAST(
        void* context_ptr,
        void* builder_ptr,
        const KuhulAST* ast
    ) {
        try {
            auto impl = reinterpret_cast<KuhulCompilerImpl*>(context_ptr);
            auto module = impl->compileAST(*ast);

            return {
                .module = new ModuleOp(module),
                .error_msg = nullptr
            };
        } catch (const std::exception& e) {
            return {
                .module = nullptr,
                .error_msg = strdup(e.what())
            };
        }
    }

    const char* EmitLLVMIR(void* module_ptr) {
        try {
            auto module = reinterpret_cast<ModuleOp*>(module_ptr);

            std::string ir;
            llvm::raw_string_ostream os(ir);
            module->print(os);

            return strdup(ir.c_str());
        } catch (...) {
            return nullptr;
        }
    }

    bool ApplyOptimizations(void* module_ptr, int level) {
        try {
            auto module = reinterpret_cast<ModuleOp*>(module_ptr);
            PassManager pm(module->getContext());

            // Add passes based on level
            if (level >= 1) {
                pm.addPass(createDeadCodeEliminationPass());
                pm.addPass(createConstantPropagationPass());
            }
            if (level >= 2) {
                pm.addPass(createLoopUnrollingPass());
            }
            if (level >= 3) {
                pm.addPass(createVectorizationPass());
            }

            return succeeded(pm.run(*module));
        } catch (...) {
            return false;
        }
    }
}
```

### 3.2.3 AST to MLIR Lowering

Create `mlir_lib/ASTtoMLIR.cpp` with visitor pattern:

```cpp
class ASTtoMLIRVisitor {
    MLIRContext& context;
    OpBuilder& builder;
    SymbolTable symbolTable;

public:
    Value visit(const Expression& expr) {
        if (auto* bin = std::get_if<BinaryOp>(&expr)) {
            return visitBinaryOp(*bin);
        } else if (auto* lit = std::get_if<Literal>(&expr)) {
            return visitLiteral(*lit);
        } else if (auto* var = std::get_if<Variable>(&expr)) {
            return visitVariable(*var);
        }
        // ... more expression types
    }

    Value visitBinaryOp(const BinaryOp& binOp) {
        auto lhs = visit(binOp.left);
        auto rhs = visit(binOp.right);
        Location loc = UnknownLoc::get(&context);

        switch (binOp.op) {
            case BinOp::Add:
                return builder.create<arith::AddIOp>(loc, lhs, rhs);
            case BinOp::Sub:
                return builder.create<arith::SubIOp>(loc, lhs, rhs);
            case BinOp::Mul:
                return builder.create<arith::MulIOp>(loc, lhs, rhs);
            // ... more operators
        }
    }

    Value visitLiteral(const Literal& lit) {
        Location loc = UnknownLoc::get(&context);

        if (auto* intLit = std::get_if<int64_t>(&lit.value)) {
            return builder.create<arith::ConstantOp>(
                loc,
                builder.getIntegerAttr(
                    builder.getIntegerType(64),
                    *intLit
                )
            );
        }
        // ... more literal types
    }
};
```

---

## Phase 3.3: LLVM IR Lowering

### 3.3.1 Standard Dialect to LLVM Lowering

Create `mlir_lib/ASTtoLLVMIR.cpp`:

```cpp
class MLIRtoLLVMLowering : public ConversionPattern {
public:
    MLIRtoLLVMLowering(MLIRContext& ctx)
        : ConversionPattern(/*...*/) {}

    LogicalResult matchAndRewrite(
        Operation* op,
        ArrayRef<Value> operands,
        ConversionPatternRewriter& rewriter
    ) const override {
        if (auto funcOp = dyn_cast<FuncOp>(op)) {
            return lowerFunction(funcOp, operands, rewriter);
        } else if (auto retOp = dyn_cast<ReturnOp>(op)) {
            return lowerReturn(retOp, operands, rewriter);
        }
        // ... more operations
    }

private:
    LogicalResult lowerFunction(
        FuncOp funcOp,
        ArrayRef<Value> operands,
        ConversionPatternRewriter& rewriter
    ) const {
        // Convert function signature
        auto llvmFuncType = lowering::convertFunctionType(funcOp.getType());

        // Create LLVM function
        auto llvmFunc = rewriter.create<LLVM::LLVMFuncOp>(
            funcOp.getLoc(),
            funcOp.getName(),
            llvmFuncType
        );

        // Move function body
        rewriter.inlineRegionBefore(funcOp.getBody(), llvmFunc.getBody());
        rewriter.eraseOp(funcOp);

        return success();
    }
};
```

---

## Phase 3.4: JIT Compilation Engine

### 3.4.1 JIT Execution Engine

Create `mlir_lib/JIT.cpp`:

```cpp
class KuhulJIT {
private:
    ExecutionEngine engine;
    std::map<std::string, JITModule> modules;

public:
    KuhulJIT(LLVMContext& llvmCtx, TargetMachine& targetMachine)
        : engine(llvmCtx, targetMachine) {}

    // Compile and execute a function immediately
    void* jitCompile(const std::string& funcName, ModuleOp module) {
        // Lower to LLVM
        auto llvmModule = lower(module);

        // Add to execution engine
        engine.addModule(std::move(llvmModule));

        // Lookup function address
        return engine.lookup(funcName);
    }

    // Execute a K'UHUL function with arguments
    template<typename ReturnType, typename... Args>
    ReturnType execute(const std::string& funcName, Args... args) {
        auto funcPtr = reinterpret_cast<ReturnType(*)(Args...)>(
            engine.lookup(funcName)
        );

        if (!funcPtr) {
            throw std::runtime_error("Function not found: " + funcName);
        }

        return funcPtr(args...);
    }
};
```

### 3.4.2 Go JIT Interface

Add to `mlir_compiler.go`:

```go
type JITExecutor struct {
    ptr *C.void  // KuhulJIT pointer
}

// NewJITExecutor creates a JIT execution engine
func NewJITExecutor() (*JITExecutor, error) {
    ptr := C.CreateJITExecutor()
    if ptr == nil {
        return nil, fmt.Errorf("failed to create JIT executor")
    }
    return &JITExecutor{ptr: ptr}, nil
}

// Execute runs a compiled function
func (j *JITExecutor) Execute(funcName string, args ...interface{}) (interface{}, error) {
    cfuncName := C.CString(funcName)
    defer C.free(unsafe.Pointer(cfuncName))

    // Convert Go args to C args
    // Call C++ execute function
    // Convert result back to Go

    result := C.ExecuteJITFunction(j.ptr, cfuncName, /* converted args */)
    return convertResult(result), nil
}
```

---

## Phase 3.5: GPU Acceleration

### 3.5.1 NVIDIA CUDA Target

Create MLIR lowering to NVIDIA dialects:

```cpp
// In mlir_lib/GPULowering.cpp
class MLIRtoNVIDIALowering : public ConversionPattern {
public:
    LogicalResult matchAndRewrite(
        Operation* op,
        ArrayRef<Value> operands,
        ConversionPatternRewriter& rewriter
    ) const override {
        if (auto loop = dyn_cast<scf::ForOp>(op)) {
            return lowerLoopToGPU(loop, rewriter);
        }
        // ... parallelize appropriate operations to GPU
    }

private:
    LogicalResult lowerLoopToGPU(
        scf::ForOp loop,
        ConversionPatternRewriter& rewriter
    ) const {
        // Mark loop for GPU execution
        rewriter.create<gpu::LaunchOp>(
            loop.getLoc(),
            /* grid size */,
            /* block size */
        );
        return success();
    }
};
```

---

## Phase 3.6: Optimization Passes

### 3.6.1 Custom Optimization Passes

Create pass framework:

```cpp
class DeadCodeEliminationPass : public Pass {
public:
    StringRef getArgument() const override { return "dce"; }
    StringRef getDescription() const override {
        return "K'UHUL dead code elimination";
    }

    void runOnOperation() override {
        auto func = getOperation();

        // Collect used values
        SetVector<Value> usedValues;
        collectUsedValues(func, usedValues);

        // Remove unused operations
        func.walk([&](Operation* op) {
            if (!isDeadOperation(op, usedValues)) return;
            rewriter.eraseOp(op);
        });
    }
};

class ConstantPropagationPass : public Pass {
public:
    StringRef getArgument() const override { return "const-prop"; }

    void runOnOperation() override {
        auto func = getOperation();

        func.walk([&](arith::ConstantOp op) {
            // Propagate constants through the IR
            propagateConstant(op);
        });
    }
};
```

---

## Implementation Sequence

### Week 1: Foundation
- [ ] Set up MLIR/LLVM build system
- [ ] Create K'UHUL MLIR dialect definition
- [ ] Define K'UHUL operations (kuhul.pop, kuhul.wo, etc.)
- [ ] Implement basic type system
- [ ] Create Go CGo bindings

### Week 2: Compiler Core
- [ ] Implement AST to MLIR lowering
- [ ] Create MLIR to LLVM IR lowering
- [ ] Build optimization pass framework
- [ ] Implement dead code elimination
- [ ] Add constant propagation pass

### Week 3: Execution & Targets
- [ ] Build JIT execution engine
- [ ] Implement CPU code generation
- [ ] Add CUDA/GPU support skeleton
- [ ] Create end-to-end tests
- [ ] Performance benchmarks

---

## Build System Integration

### CMake Configuration

```cmake
# mlir/CMakeLists.txt

project(KuhulMLIR CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(MLIR REQUIRED CONFIG)
find_package(LLVM REQUIRED CONFIG)

add_library(kuhul_mlir SHARED
    lib/KuhulDialect.cpp
    lib/KuhulOps.cpp
    lib/KuhulCompiler.cpp
    lib/ASTtoMLIR.cpp
    lib/ASTtoLLVMIR.cpp
    lib/JIT.cpp
    lib/GPULowering.cpp
)

target_include_directories(kuhul_mlir
    PUBLIC ${MLIR_INCLUDE_DIRS}
    PUBLIC ${LLVM_INCLUDE_DIRS}
    PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/include
)

target_link_libraries(kuhul_mlir
    ${MLIR_LIBRARIES}
    ${LLVM_LIBRARIES}
)

# Enable cgo linking in Go
set_target_properties(kuhul_mlir PROPERTIES
    VERSION 3.0.0
    SOVERSION 3
)

install(TARGETS kuhul_mlir LIBRARY DESTINATION lib)
install(DIRECTORY include/kuhul DESTINATION include)
```

### Go Build Tags

```bash
# Build with MLIR support (when available)
CGO_ENABLED=1 CGO_CXXFLAGS="-I/usr/include/mlir" \
CGO_LDFLAGS="-L/usr/lib -lMLIR" \
go build -tags "mlir" ./cmd/ollama

# Build without MLIR (fallback to interpreter)
go build ./cmd/ollama
```

---

## Testing Strategy

### Unit Tests
- Individual MLIR operations
- Type system correctness
- Lowering correctness
- Pass behavior

### Integration Tests
- Full K'UHUL → MLIR → LLVM → native code flow
- Cross-platform code generation
- JIT compilation accuracy

### Performance Tests
- Compilation time benchmarks
- Generated code performance vs interpreted
- Memory usage profiling
- GPU offload efficiency

### Example Test

```go
// mlir_compiler_test.go
func TestMLIRCompilation(t *testing.T) {
    compiler, _ := NewMLIRCompiler()

    // Parse K'UHUL code
    code := `
        pop add(a, b)
            chen a + b
    `

    ast, _ := kuhul.Parse(code)

    // Compile to MLIR
    module, _ := compiler.CompileAST(ast)

    // Generate LLVM IR
    llvmIR, _ := module.EmitLLVMIR()

    // Verify IR structure
    assert.Contains(t, llvmIR, "@add")
    assert.Contains(t, llvmIR, "define")
}

func BenchmarkCompilation(b *testing.B) {
    compiler, _ := NewMLIRCompiler()
    ast, _ := kuhul.Parse(largeProgram)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        compiler.CompileAST(ast)
    }
}
```

---

## Performance Expectations

### Compilation
- Small functions: < 10ms
- Medium programs (1-10 funcs): 50-200ms
- Large programs: 500ms - 2s

### Execution
- Interpreted: 1x baseline
- JIT compiled: 5-20x faster
- GPU accelerated (suitable workloads): 50-1000x faster

### Memory
- MLIR context: ~5-10 MB
- JIT compiled module: variable (100 KB - 50 MB)
- GPU memory (CUDA): 100 MB - 8 GB

---

## Dependencies

### Required
- LLVM 14+ (with development headers)
- MLIR 14+ (built with LLVM)
- CMake 3.13+
- C++17 compiler

### Optional
- CUDA Toolkit 11+ (for GPU support)
- HIP (for AMD GPU support)
- Intel oneAPI (for Intel GPU support)

### Installation

```bash
# Ubuntu/Debian
sudo apt-get install llvm-14-dev mlir-14-dev cmake

# macOS
brew install llvm cmake

# Build MLIR from source (if needed)
git clone https://github.com/llvm/llvm-project.git
cd llvm-project
mkdir build
cd build
cmake -G Ninja ../llvm \
    -DLLVM_ENABLE_PROJECTS="mlir" \
    -DCMAKE_BUILD_TYPE=Release
ninja
sudo ninja install
```

---

## Milestones

### ✅ Phase 1-2: Platform Abstraction (COMPLETE)
- Windows/Unix path handling
- Process execution
- Service discovery
- HTTP proxy bridge

### 🔄 Phase 3.1: MLIR Dialect (IN PROGRESS)
- K'UHUL MLIR operations
- Custom type system
- TableGen definitions

### ⏳ Phase 3.2: Compiler Core (PLANNED)
- AST to MLIR lowering
- MLIR to LLVM lowering
- Optimization passes

### ⏳ Phase 3.3: Execution (PLANNED)
- JIT compilation
- CPU code generation
- GPU targets

### ⏳ Phase 3.4: Optimization (PLANNED)
- Performance tuning
- Benchmark suite
- Production hardening

---

## Success Criteria

✅ Phase 3 Complete when:
1. K'UHUL code compiles to LLVM IR
2. JIT compilation produces executable native code
3. Compiled code runs 5-20x faster than interpreted
4. GPU acceleration available (proof of concept)
5. All tests pass on Windows and Linux
6. Documentation complete with examples

---

## Resources & References

### MLIR Documentation
- https://mlir.llvm.org/
- MLIR Toy Tutorial
- MLIR Dialect Creation Guide

### LLVM Resources
- https://llvm.org/docs/
- LLVM Language Reference Manual
- LLVM JIT Compilation

### Related Projects
- Swift compiler (uses MLIR)
- CIRCT (hardware design in MLIR)
- Enzyme (automatic differentiation in MLIR)

---

## Next Steps After Phase 3

- **Production Hardening**: Error recovery, edge case handling
- **Distribution**: Vendored LLVM in binary, optional system LLVM
- **Self-Hosting**: Rewrite K'UHUL compiler in K'UHUL
- **Advanced Features**: GPU kernels, vectorization tuning, profiling

---

**Status**: Ready for Phase 3.1 Implementation
**Estimated Total Duration**: 2-3 weeks
**Team**: 1-2 developers (C++, MLIR knowledge required)
