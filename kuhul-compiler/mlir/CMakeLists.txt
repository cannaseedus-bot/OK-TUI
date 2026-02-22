cmake_minimum_required(VERSION 3.20)
project(kuhul-mlir LANGUAGES CXX)

find_package(MLIR REQUIRED CONFIG)

message(STATUS "Using MLIRConfig.cmake in: ${MLIR_DIR}")

list(APPEND CMAKE_MODULE_PATH "${MLIR_CMAKE_DIR}")
include(TableGen)
include(AddMLIR)
include(AddLLVM)

include_directories(${PROJECT_SOURCE_DIR}/include)
include_directories(${MLIR_INCLUDE_DIRS})
add_definitions(${MLIR_DEFINITIONS})

set(LLVM_TARGET_DEFINITIONS include/Kuhul/KuhulOps.td)
mlir_tablegen(include/Kuhul/KuhulOps.h.inc -gen-op-decls)
mlir_tablegen(include/Kuhul/KuhulOps.cpp.inc -gen-op-defs)
add_public_tablegen_target(KuhulOpsIncGen)

set(LLVM_TARGET_DEFINITIONS include/Kuhul/KuhulDialect.td)
mlir_tablegen(include/Kuhul/KuhulDialect.h.inc -gen-dialect-decls)
mlir_tablegen(include/Kuhul/KuhulDialect.cpp.inc -gen-dialect-defs)
add_public_tablegen_target(KuhulDialectIncGen)

add_mlir_dialect_library(MLIRKuhul
  lib/KuhulDialect.cpp
  lib/KuhulOps.cpp
  lib/KuhulLowering.cpp
  lib/KuhulVerifier.cpp

  DEPENDS
  KuhulOpsIncGen
  KuhulDialectIncGen

  LINK_LIBS PUBLIC
  MLIRIR
  MLIRLLVMCommonConversion
  MLIRLLVMDialect
)
