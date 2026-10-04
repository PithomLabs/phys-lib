---
id: symbolic-normalize
source: phys-gr/symbolic/normalize.go
class: SPLIT
target: phys-math rules (CANON/SIMPLIFY/RAT) AND phys-lib/gr local algebra
status: adopted
---

Why it is theory-agnostic:
N/A — split. Commutative flatten/sort/fold and exact rational folding are
generic rewriting; trigReduce, regime ZeroTest, and SplitLinear are
GR-regime helpers with generic names.

Known/current consumers:
phys-math Canonicalize + EXPR-CANON/SIMPLIFY/RAT-NORMALIZE-001.

Potential future consumers:
Any framework needing canonical commutative normalization reuses the
phys-math rules; none reuse spherical trig reduction.

Theory-specific assumptions:
Sin²+Cos² pairing (spherical terms); NONZERO-for-symbols (r,A,B,sin ≠ 0
regularity smuggled into zero-testing); SplitLinear M-1 weak-field path.

Stable interface:
Rule contracts §17.3–17.5 with fail-closed conditions; gr keeps honestly
renamed weak-field/spherical helpers (WeakFieldTruncate-style), never as
generic Normalize behavior.

Dependencies:
phys-math side: phys-artifact + identity. gr side: framework assumptions.

Why it does not belong in phys-math:
Regime non-vanishing and spherical/weak-field reductions encode GR physics.

Why it does not belong in phys-lib/gr:
Flatten/sort/fold/idempotent rational rules are the mathematical common
denominator; the regime ZeroTest itself is DELETE-replaced by the D2
explicit-condition bridge (no implicit nonzero rule survives).
