---
id: replay
source: phys-gr/replay/
class: SPLIT
target: phys-math derivation replay AND phys-lib/gr protocol adapter
status: adopted
---

Why it is theory-agnostic:
N/A — split. Hash-chained ordered steps, VerifyChain linkage checks, and
deterministic ReExecute dispatch shape are generic replay mechanics; the
gr:* operation namespace and curvature/vacuum_check-as-operations are
GR-domain inferences sharing one flat level.

Known/current consumers:
phys-math Replay (same-run availability, rule-family validation, exact
output-ref equality, condition-context match, carry-through results).

Potential future consumers:
Any executable derivation protocol reuses the replay shape; none reuse
gr:* operation identities.

Theory-specific assumptions:
Closed gr:* allowlist mixing math operations with physical inferences;
kernel:differentiate/compare provenance split; byte-identical implication
assertions (vacuum_check) presented as machine-checked steps.

Stable interface:
Transformation (durable, standalone) vs DerivationStep (inline, ordered)
lifecycle boundary; protocol adapters in gr map legacy steps to the seven
closed rules or declare them F1-deferred pipeline content.

Dependencies:
phys-math side: rules + codec + store. gr side: framework families.

Why it does not belong in phys-math:
GR operation identities and recorded physical implications are domain content.

Why it does not belong in phys-lib/gr:
Ordered hash-chained replay with exact-output equality is the generic
executable-derivation mechanism.
