---
id: symbolic-bridge-certify
source: phys-gr/symbolic/bridge.go + phys-gr/symbolic/certify.go
class: SPLIT
target: D9 fallible projection pattern AND phys-lib/gr local checks
status: adopted
---

Why it is theory-agnostic:
N/A — split. The bridge discipline (explicit deterministic mapping with
documented preserved/rejected information, no implicit fallback) is the
generic theory-local pattern; denominator-clearing certification and
phys-core object conversion serve the GR vacuum pipeline.

Known/current consumers:
gr ToPhysMath() projection + ErrUnrepresentableProjection; D2 bridge.

Potential future consumers:
Any framework-local algebra reuses the projection discipline; none reuse
GR denominator-clearing.

Theory-specific assumptions:
CombinedNumerator/CertifyZero denominator clearing tuned to GR rational
functions; TestOracle numeric falsifier bound to the GR regime.

Stable interface:
One-way projection of the representable subset; explicit failure otherwise;
mapping documented per §15 (source type, target, conditions, preserved and
rejected information).

Dependencies:
gr side: phys-math (projection target).

Why it does not belong in phys-math:
Certification tactics and core-object conversion encode GR pipeline needs.

Why it does not belong in phys-lib/gr:
N/A for the discipline half (it is the §15 pattern itself, owned by the
architecture); the GR tactic half belongs here.
