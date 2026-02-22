#include "Kuhul/KuhulDialect.h"

#include "Kuhul/KuhulDialect.cpp.inc"

using namespace mlir;
using namespace kuhul;

void KuhulDialect::initialize() {
  addOperations<
#define GET_OP_LIST
#include "Kuhul/KuhulOps.cpp.inc"
      >();
}
