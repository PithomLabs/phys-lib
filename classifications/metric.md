---
id: metric
source: phys-gr/metric/
class: PHYS-LIB/GR
target: phys-lib/gr framework artifacts (PhysicalDefinition/PhysicalEquation)
status: adopted
---

Why it is theory-agnostic:
It is not. Build() constructs one ansatz — diag(-A(r),B(r),r²,r²sin²θ) —
encoding signature, chart, topology, asymptotic-flatness-ready form,
Lambda=0, and regularity in a single constructor.

Known/current consumers:
GR Schwarzschild pipeline (preserved in phys-gr).

Potential future consumers:
None; other frameworks define their own metric/geometry objects.

Theory-specific assumptions:
Static spherical symmetry; -+++ signature; spherical-static chart; Lambda=0;
A,B,r,sin nonzero regularity; diagonal no-rotation/no-charge form.

Stable interface:
PhysicalDefinition + PhysicalEquation over PhysMath-expressible fragments,
with convention/regularity assumptions as explicit artifacts (D15); the
computation pipeline itself is F1-deferred.

Dependencies:
gr framework + phys-math (projection target only).

Why it does not belong in phys-math:
Every line beyond the diagonal inverse check is GR semantics.

Why it does not belong in phys-lib/gr:
N/A — it belongs here.
