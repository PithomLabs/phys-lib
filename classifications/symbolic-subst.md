---
id: symbolic-subst
source: phys-gr/symbolic/subst.go
class: SPLIT
target: phys-math rules (SUBSTITUTE/APPLY) AND phys-lib/gr ansatz pinning
status: adopted
---

Why it is theory-agnostic:
N/A — split. Structural simultaneous substitution matched by ParameterID is
generic; zero-order-UF-only targets and no-UF-arg-rewrite pin the
static-spherical ansatz invariant.

Known/current consumers:
phys-math EXPR-SUBSTITUTE-001 / EXPR-APPLY-001 with declaring-Function
inputs, sorted entries, distinct-ParameterID enforcement.

Potential future consumers:
Any framework substituting into declared function bodies reuses the rules;
none reuse the A-is-always-a-function-of-r pinning.

Theory-specific assumptions:
Substitution targets limited to Symbol/zero-order UFN; identity pinned to
declared symbols (forbids α-rename/coordinate change); single-arg f(r) shape.

Stable interface:
Rule contracts §§17.7–17.8; gr-local callers supply declaring Functions and
pre-isolated Call subtrees (no NodePath selectors).

Dependencies:
phys-math side: phys-artifact + identity + codec.

Why it does not belong in phys-math:
Ansatz-shape restrictions are GR-specific application policy.

Why it does not belong in phys-lib/gr:
Simultaneous ParameterID substitution with explicit declaring context is the
generic mechanism behind both rules.
