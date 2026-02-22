#!/bin/bash

echo "Building KUHUL compiler..."

cargo build --release

mkdir -p output

echo "Done"
