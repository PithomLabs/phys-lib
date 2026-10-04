---
id: tensor
source: phys-gr/tensor/
class: SPLIT
target: phys-math Tensor/Index/Collection AND phys-lib/gr chart/signature validation
status: adopted
---

Why it is theory-agnostic:
N/A — split. Dense row-major slots, offsets, SET-vs-SEQUENCE discipline, and
shape validation are mathematics; Dim=4, Domain=spacetime, ChartID-equality
gating, symmetric-metric moves, and re-Normalize coupling are GR bindings.

Known/current consumers:
phys-math Tensor/Index/Collection construction, validation, serialization.

Potential future consumers:
Any framework needing dense component arrays reuses the parametric core;
none reuse spherical-static chart binding.

Theory-specific assumptions:
Four-dimensional spacetime; single-chart fail-closed ops (no covariance);
symmetric Levi-Civita metric in index moves; regime zero-test in symmetry
verification.

Stable interface:
Parametric core types (rank/dimension/variance/dense checks) with chart,
signature, symmetry, and connection-distinction enforced in the gr
validation layer (plan9.3 D11).

Dependencies:
phys-math side: phys-artifact + codec. gr side: framework assumptions.

Why it does not belong in phys-math:
Chart equality, signature, and connection semantics are physical meaning.

Why it does not belong in phys-lib/gr:
Rank/slots/row-major/SET-vs-SEQUENCE mechanics are the reusable core.
