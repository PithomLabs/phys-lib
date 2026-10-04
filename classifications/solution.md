---
id: solution
source: phys-gr/solution/
class: DEFER
target: preserved reference only; future separately-scoped re-expression (F1)
status: adopted
---

Why it is theory-agnostic:
It is not classifiable for promotion: the 14-step Schwarzschild derivation,
M1–M8 harness, SessionCorrespondence, and frozen-Phys coupling form one
GR-workload executable, not a separable abstraction.

Known/current consumers:
Legacy regression baseline; semantic anchor for the GR port.

Potential future consumers:
A future Schwarzschild re-expression phase (separately scoped, with its own
frozen-Phys decision).

Theory-specific assumptions:
All of GR: ansatz, MTW, -+++, Lambda=0, asymptotic flatness, mu=GM/c²,
HYPOTHESIS correspondence k2↔-2mu.

Stable interface:
None in Phase 1 — no code is extracted from solution/ in Phase 1. The
protocol PATTERN (embed/strict-decode/ordered-expectations/firewalls/
mutation harness) is reused from it, never its content.

Dependencies:
None (read-only reference).

Why it does not belong in phys-math:
Pure GR derivation content.

Why it does not belong in phys-lib/gr:
Phase-1 gr scope (D7/F1) excludes the pipeline re-expression; importing it
now would force frozen-Phys integration and scope explosion.
