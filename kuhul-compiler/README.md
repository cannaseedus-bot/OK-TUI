# KUHUL Compiler

KUHUL is a self-hosting compiler pipeline scaffold wired for MLIR + LLVM integration.

## Pipeline

```text
KUHUL source
  -> KUHUL compiler (KUHUL)
  -> MLIR (KuhulDialect)
  -> LLVM IR
  -> native machine code
```

## Quick start

```bash
./bootstrap.sh
./build.sh
make -C . build-kuhul
```
