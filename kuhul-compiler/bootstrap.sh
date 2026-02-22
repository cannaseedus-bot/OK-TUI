#!/bin/bash
set -euo pipefail

git clone https://github.com/llvm/llvm-project.git
cd llvm-project
git checkout llvmorg-18.1.0
cd ..
