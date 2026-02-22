#!/bin/bash
set -euo pipefail

mkdir -p build-llvm
cd build-llvm

cmake -G Ninja ../llvm-project/llvm \
  -DLLVM_ENABLE_PROJECTS="mlir" \
  -DLLVM_BUILD_EXAMPLES=OFF \
  -DLLVM_BUILD_TESTS=OFF \
  -DCMAKE_BUILD_TYPE=Release

ninja
cd ..
