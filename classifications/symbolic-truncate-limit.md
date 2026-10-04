---
id: symbolic-truncate-limit
source: phys-gr/symbolic/truncate.go + phys-gr/limit/
class: PHYS-LIB/GR
target: phys-lib/gr (renamed WeakFieldTruncate + weak-field extraction)
status: adopted
---

Why it is theory-agnostic:
It is not. Single rule Pow(1+u,-1)→1-u at eps-degree 1, eps-degree counting
over Symbols, M-1 causal path SplitLinear — all shaped by the Schwarzschild
closed form, static spherical symmetry, Lambda=0, Phi=-GM/r extraction.

Known/current consumers:
GR weak-field extraction only.

Potential future consumers:
None demonstrated; a future framework needing series truncation must bring
its own rule with its own evidence (no generic series engine is implied).

Theory-specific assumptions:
Schwarzschild closed form; static spherical background; Lambda=0;
first-order weak-field regime; eps=GM/c²r identification.

Stable interface:
Framework-local function with explicit mapping documentation (§15
theory-local mechanism); deterministic fallible projection where representable.

Dependencies:
gr framework + phys-math (projection target only).

Why it does not belong in phys-math:
It is a single-workload reduction engine, not general mathematics. The name
Truncate must not survive as a generic claim — rename, don't generalize.

Why it does not belong in phys-lib/gr:
N/A — it belongs here.
