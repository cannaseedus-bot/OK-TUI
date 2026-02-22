# KUHUL System Core — Recommended Next Execution Plan

Given the current architecture (stdlib + MLIR dialect sketch + WebGPU runtime skeleton + deterministic replay skeleton), the most leverage comes from **locking correctness first**.

## Recommendation: Start with Option 1 (MLIR verifier passes)

Why first:
- It enforces phase and determinism invariants before backend codegen.
- It prevents invalid IR from reaching LLVM/GPU/replay targets.
- It gives a contract every later subsystem can rely on.

Without verifier guarantees, GPU/runtime and replay work may diverge subtly and become expensive to debug.

## Milestone order

1. **MLIR verifier passes (now)**
2. GPU memory layout optimizer
3. Replay compression (SCXQ2 bit-packing + ANS)
4. WebGPU tensor engine
5. LLVM native AOT backend
6. KUHUL v1.0 language spec export

## Milestone 1 scope (concrete)

### New verifier checks

- `phase-transition-check`
  - Allowed transitions only (`compute -> render -> train` or policy-defined graph).
  - No implicit phase entry.

- `barrier-satisfaction-check`
  - `kuhul.flux.barrier` must reference known channels.
  - All required channels must be provably satisfied before dependent ops.

- `determinism-check`
  - Disallow non-deterministic side effects in deterministic regions.
  - Enforce stable operation ordering for replayable segments.

- `glyph-resolution-check`
  - Every `kuhul.glyph.exec` must resolve to a registered glyph symbol.

### Expected diagnostics

- Error messages include:
  - op location
  - violated rule id
  - actionable fix hint

Example style:

`error: [KUHUL-PHASE-001] illegal phase transition compute -> train; insert render phase or update phase policy.`

## Definition of Done (M1)

- Verifier rejects all invalid phase/barrier/determinism patterns covered by tests.
- Verifier accepts all valid golden programs.
- CI includes positive and negative test fixtures for each rule.
- `--target mlir` emits diagnostics with stable rule IDs.

## Deliverables

- `mlir/verify/PhaseTransitionVerifier.cpp`
- `mlir/verify/BarrierVerifier.cpp`
- `mlir/verify/DeterminismVerifier.cpp`
- `mlir/verify/GlyphResolutionVerifier.cpp`
- `mlir/test/verify/*.mlir`
- verifier pass registration and pipeline wiring

## Minimal test matrix

- Valid phase chain (pass)
- Invalid transition (fail)
- Missing barrier dependency (fail)
- Unknown barrier channel (fail)
- Unregistered glyph (fail)
- Deterministic region with non-deterministic op (fail)
- Canonical valid mixed program (pass)

## Suggested immediate command backlog

- Add verifier pass interfaces + rule IDs.
- Wire passes into `Parse AGL -> KuhulDialect -> Verify -> Lower` pipeline.
- Add fixture-based tests and diagnostic snapshots.
- Stabilize diagnostic wording before adding more backends.

---

If we execute this order, every subsequent backend task becomes mechanical rather than speculative.
