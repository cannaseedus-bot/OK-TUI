#!/usr/bin/env bash
set -euo pipefail

mkdir -p build

cargo build --release --manifest-path bootstrap/Cargo.toml

bootstrap/target/release/kuhulc_stage0 \
  kuhulc/main.khl \
  -o build/kuhulc_stage1

build/kuhulc_stage1 \
  kuhulc/main.khl \
  -o build/kuhulc_stage2

cmp build/kuhulc_stage1 build/kuhulc_stage2

echo "Self-hosting verified"
