---
id: curvature
source: phys-gr/curvature/
class: PHYS-LIB/GR
target: phys-lib/gr framework semantics
status: adopted
---

Why it is theory-agnostic:
It is not. MTW sign, R^rho_murhonu contraction, Scalar/Esstein without Λ
term present convention choices as THE Riemann/Ricci/Einstein.

Known/current consumers:
GR vacuum pipeline (preserved in phys-gr).

Potential future consumers:
None; other frameworks define their own curvature semantics.

Theory-specific assumptions:
MTW sign convention; first+third contraction; Lambda=0 (no +Λg term);
4D loops over the GR connection.

Stable interface:
Convention assumptions as explicit convention-Category artifacts (D15);
computation pipeline F1-deferred.

Dependencies:
gr framework + phys-math.

Why it does not belong in phys-math:
Sign/contraction/Λ choices are physical conventions, not mathematics.

Why it does not belong in phys-lib/gr:
N/A — it belongs here.
