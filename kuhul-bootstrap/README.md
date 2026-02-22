# KUHUL Self-Hosting Compiler

This repository bootstraps the KUHUL compiler and enables full self-hosting.

Pipeline:

KUHUL → AST → MLIR → LLVM / WebGPU / SCXQ2

Bootstrap stages:

Stage0: Rust bootstrap compiler  
Stage1: KUHUL compiler compiled by Stage0  
Stage2: KUHUL compiler compiled by Stage1

Stage2 must match Stage1 binary hash.

This proves self-hosting correctness.
