include "Kuhul/KuhulDialect.td"
include "mlir/IR/OpBase.td"

def KuhulDialectOp<string mnemonic, list<Trait> traits = []> :
    Op<KuhulDialect, mnemonic, traits>;

def FluxTickOp : KuhulDialectOp<"flux.tick"> {
  let summary = "Flux tick operation";
  let assemblyFormat = "attr-dict";
}

def FluxPhaseOp : KuhulDialectOp<"flux.phase"> {
  let arguments = (ins StrAttr:$phase);
  let assemblyFormat = "$phase attr-dict";
}
