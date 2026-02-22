# KUHUL Self-Hosting Compiler Architecture (Bootstrap Plan)

This document captures a practical self-hosting compiler pipeline for KUHUL that mirrors real-world language bootstrap patterns (e.g., Rust, Go, Zig, Swift).

> Note: This architecture describes the **compiler implementation stack** (KUHUL compiler language tooling), not the frozen KUHUL π execution core.

## 0) What “self-hosting KUHUL compiler” means

```text
KUHUL source
   ↓
KUHUL compiler (written in KUHUL)
   ↓
AST
   ↓
MLIR
   ↓
LLVM / WebGPU / SCXQ2
```

The compiler itself is a KUHUL program (instead of being permanently written in Rust/C++).

## 1) Bootstrapping strategy (critical)

A three-stage bootstrap keeps the transition safe and verifiable:

```text
Stage 0 — Host compiler (Rust version)
Stage 1 — KUHUL compiler compiled by Rust compiler
Stage 2 — KUHUL compiler compiled by KUHUL compiler
```

After Stage 2, Rust is optional for ongoing KUHUL compiler evolution.

## 2) Proposed repository structure

```text
kuhulc/
│
├─ src/
│   ├─ main.khl
│   ├─ lexer.khl
│   ├─ parser.khl
│   ├─ ast.khl
│   ├─ lowering.khl
│   ├─ mlir_emit.khl
│   ├─ llvm_emit.khl
│   ├─ scxq2_emit.khl
│   └─ gpu_emit.khl
│
├─ std/
│   ├─ string.khl
│   ├─ vector.khl
│   ├─ map.khl
│   └─ io.khl
│
├─ bootstrap/
│   ├─ kuhulc_stage0
│   └─ kuhulc_stage1
```

## 3) Compiler entry point (`main.khl`)

```kuhul
⟁Pop compiler.main

    Wo args = io.args()

    Wo input = args[0]
    Wo output = args[1]

    Wo source = io.read_file(input)

    Wo tokens = lexer.tokenize(source)

    Wo ast = parser.parse(tokens)

    Wo mlir = lowering.to_mlir(ast)

    Wo binary = llvm_emit.compile(mlir)

    io.write_file(output, binary)

⟁Ch'en
```

## 4) Lexer sketch (`lexer.khl`)

```kuhul
⟁Pop lexer.tokenize

    Wo src = @input
    Wo tokens = []

    Wo i = 0

    Sek while i < len(src)

        Wo c = src[i]

        Sek if is_letter(c)

            Wo ident = read_identifier(src, i)

            tokens.push({
                type: "IDENT",
                value: ident
            })

            i += len(ident)

        Sek else

            i += 1

    Ch'en tokens
```

## 5) Parser sketch (`parser.khl`)

```kuhul
⟁Pop parser.parse

    Wo tokens = @input

    Wo ast = {
        type: "Module",
        body: []
    }

    Sek for token in tokens

        Sek if token.type == "IDENT"

            ast.body.push({
                type: "Symbol",
                name: token.value
            })

    Ch'en ast
```

## 6) KUHUL → MLIR lowering (`lowering.khl`)

```kuhul
⟁Pop lowering.to_mlir

    Wo ast = @input

    Wo module = mlir.create_module()

    Sek for node in ast.body

        Sek if node.type == "Symbol"

            module.add_op({
                dialect: "kuhul",
                op: "symbol",
                name: node.name
            })

    Ch'en module
```

## 7) MLIR → LLVM backend (`llvm_emit.khl`)

```kuhul
⟁Pop llvm_emit.compile

    Wo module = @input

    Wo llvm_module = llvm.lower(module)

    Wo binary = llvm.codegen(llvm_module)

    Ch'en binary
```

## 8) GPU backend (`gpu_emit.khl`)

```kuhul
⟁Pop gpu_emit.compile

    Wo module = @input

    Wo shader = gpu.lower(module)

    Wo binary = gpu.compile(shader)

    Ch'en binary
```

## 9) SCXQ2 backend (`scxq2_emit.khl`)

```kuhul
⟁Pop scxq2_emit.compile

    Wo module = @input

    Wo fieldmap = scxq2.lower(module)

    Wo packed = scxq2.pack(fieldmap)

    Ch'en packed
```

## 10) AST model sketch (`ast.khl`)

```kuhul
Wo AST = {

    Module: {
        body: []
    },

    Symbol: {
        name: ""
    },

    FluxTick: {
        tick: 0
    }

}
```

## 11) Standard library example (`string.khl`)

```kuhul
⟁Pop string.len

    Wo s = @input

    Wo count = 0

    Sek for c in s

        count += 1

    Ch'en count
```

## 12) Bootstrap process

1. Compile KUHUL compiler using Rust bootstrap:

```text
kuhulc_stage0 main.khl → kuhulc_stage1
```

2. Compile KUHUL compiler using itself:

```text
kuhulc_stage1 main.khl → kuhulc_stage2
```

3. Verify deterministic equivalence:

```text
diff kuhulc_stage1 kuhulc_stage2
```

If equivalent, self-hosting is complete.

## 13) Flux-governed compiler execution

```kuhul
@flux.phase { enter: "compile" }

lexer.tokenize()
parser.parse()
lowering.to_mlir()
llvm_emit.compile()

@flux.phase { enter: "idle" }
```

## 14) Deterministic replay log

Each compile can emit a replay/provenance artifact:

```json
{
  "@type": "compile_log",
  "@source_hash": "...",
  "@ast_hash": "...",
  "@mlir_hash": "...",
  "@binary_hash": "..."
}
```

## 15) End-state

At maturity, KUHUL can compile:

- itself
- native binaries
- GPU code
- SCXQ2 packed executables

## 16) Final architecture

```text
KUHUL source
   ↓
KUHUL compiler (KUHUL)
   ↓
MLIR
   ↓
LLVM / GPU / SCXQ2
   ↓
native executable
```

A fully self-hosted pipeline with deterministic replay and verifiable outputs.
