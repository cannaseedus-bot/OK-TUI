LLVM=build-llvm

build-mlir:
	cmake -G Ninja mlir \
		-DMLIR_DIR=$(LLVM)/lib/cmake/mlir

build-kuhul:
	cargo build --release --manifest-path bootstrap/Cargo.toml

compile:
	build/kuhulc_stage1 examples/hello.khl build/hello
