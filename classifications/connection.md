---
id: connection
source: phys-gr/connection/
class: PHYS-LIB/GR
target: phys-lib/gr framework semantics (connection-vs-tensor distinction kept)
status: adopted
---

Why it is theory-agnostic:
It is not. Compute implements the Levi-Civita formula specifically
(torsion-free + metric-compatible, unnamed), with GR Diff and chart binding,
and a Schwarzschild-workload NonZeroClasses count nearby in solution/.

Known/current consumers:
GR curvature pipeline (preserved in phys-gr).

Potential future consumers:
None demonstrated; generic affine-connection capability is not claimed.

Theory-specific assumptions:
Torsion zero; metric compatibility; symmetric lower pair; single chart.

Stable interface:
Connection ≠ Tensor by type (the existing clean separation is preserved);
re-expression against phys-math fragments only where PhysMath-expressible;
pipeline F1-deferred.

Dependencies:
gr framework + phys-math.

Why it does not belong in phys-math:
Christoffel semantics with unnamed geometric assumptions are GR meaning.

Why it does not belong in phys-lib/gr:
N/A — it belongs here.
