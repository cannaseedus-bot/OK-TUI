#include "mlir/Conversion/LLVMCommon/ConversionTarget.h"
#include "mlir/Dialect/LLVMIR/LLVMDialect.h"
#include "mlir/IR/BuiltinOps.h"
#include "mlir/IR/Builders.h"

using namespace mlir;

namespace kuhul {

void lowerFluxTick(ModuleOp module, OpBuilder &builder, Operation *op) {
  auto func = module.lookupSymbol<LLVM::LLVMFuncOp>("agl_flux_tick");
  if (!func) {
    auto voidTy = LLVM::LLVMVoidType::get(builder.getContext());
    func = builder.create<LLVM::LLVMFuncOp>(op->getLoc(), "agl_flux_tick", LLVM::LLVMFunctionType::get(voidTy, {}));
  }

  builder.create<LLVM::CallOp>(op->getLoc(), TypeRange{}, builder.getSymbolRefAttr(func), ValueRange{});
}

} // namespace kuhul
