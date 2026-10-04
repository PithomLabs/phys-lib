---
id: symbolic-differentiate
source: phys-gr/symbolic/differentiate.go
class: SPLIT
target: phys-math rule (DIFFERENTIATE) AND phys-lib/gr caller-side bounds
status: adopted
---

Why it is theory-agnostic:
N/A — split. Linearity, Leibniz, Pow/Sin/Cos structural rules are generic
calculus; the order-2 cap, Symbol-name-only variable, and no-implicit-chain
restrictions are workload bounds.

Known/current consumers:
phys-math DIFFERENTIATE-001 (closed scalar rules, single target,
nonnegative-integer Pow exponent, Order chaining).

Potential future consumers:
Any framework differentiating scalar expressions reuses the rule; none reuse
the order-2 cap.

Theory-specific assumptions:
MaxDerivativeOrder=2 (Ricci workload); differentiation variable restricted to
a declared Symbol name; no chain through UnknownFunction args.

Stable interface:
BoundParameterRef + Order parameters; BPR-matches-declaring-Function gate;
Call/Relation/BranchSet/Sqrt out of scope (fail closed).

Dependencies:
phys-math side: phys-artifact + identity.

Why it does not belong in phys-math:
Caps and variable restrictions encode what GR needs, not what differentiation is.

Why it does not belong in phys-lib/gr:
Closed ordinary partial differentiation is mathematical infrastructure.
