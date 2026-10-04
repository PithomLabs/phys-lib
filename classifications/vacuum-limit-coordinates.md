---
id: vacuum-limit-coordinates
source: phys-gr/vacuum/ + phys-gr/limit/ + phys-gr/coordinates/
class: PHYS-LIB/GR
target: phys-lib/gr framework data and checks
status: adopted
---

Why it is theory-agnostic:
It is not. R_uv=0 assertions (M-2 schema excluding expr), the Phi=-GM/r
causal extraction, and chart ID spherical-static [t r theta phi] are GR
regime content.

Known/current consumers:
GR solution/verification path (preserved in phys-gr).

Potential future consumers:
None demonstrated.

Theory-specific assumptions:
Vacuum Einstein equations with Lambda=0; weak-field regime; static
spherical chart; mu=GM/c² identification.

Stable interface:
vacuum assertions re-expressed as framework checks where PhysMath-expressible;
chart as framework data bound to physmath/collection/1 SpaceRef; limit as
WeakFieldTruncate local engine.

Dependencies:
gr framework + phys-math.

Why it does not belong in phys-math:
Field equations, weak-field identification, and chart choice are physics.

Why it does not belong in phys-lib/gr:
N/A — it belongs here.
