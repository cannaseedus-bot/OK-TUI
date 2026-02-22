#include "Kuhul/KuhulDialect.h"

namespace kuhul {

static mlir::LogicalResult verifyFluxPhase(const mlir::StringAttr &phase) {
  if (!phase || phase.getValue().empty()) {
    return mlir::failure();
  }
  return mlir::success();
}

} // namespace kuhul
