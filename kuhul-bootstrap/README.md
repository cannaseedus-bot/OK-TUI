# KUHUL Bootstrap Compiler

Deterministic Atomic Block Compiler.

Supports:

- Native C backend
- WebGPU backend
- SCXQ2 compressed backend

## Build

Linux / Mac:

```bash
./build/build.sh
```

Windows:

```powershell
.\build\build.ps1
```

## Compile example

```bash
./target/release/kuhulc samples/hello.agl --target c --output output/hello.c
```

## Run native example

```bash
gcc output/hello.c runtime/agl_runtime.c -o hello
./hello
```

## SCXQ2 output

```bash
./target/release/kuhulc samples/hello.agl --target scxq2 --output output/hello.scx
```

## GPU output

```bash
./target/release/kuhulc samples/gpu_demo.agl --target gpu --output output/gpu.ts
```
