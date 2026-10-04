---
id: symbolic-expr
source: phys-gr/symbolic/expr.go
class: SPLIT
target: phys-math ast+codec AND phys-lib/gr local algebra
status: adopted
---

Why it is theory-agnostic:
N/A — this row splits. The tree mechanics (closed node kinds, recursive
canonical encoding, byte-equality) are mathematics; the 9-kind set itself is
workload-shaped.

Known/current consumers:
phys-math Expr/DecodeExpr/Canonicalize (mechanics); gr GRExpr/ToPhysMath (local kinds).

Potential future consumers:
Other phys-* frameworks consume phys-math mechanics; none consume GR kinds.

Theory-specific assumptions:
MaxDerivativeOrder=2 (Ricci needs 2nd derivatives), UnknownFunction shaped for
A(r)/B(r), ValidSymbolName bound, Sin/Cos present for spherical terms.

Stable interface:
phys-math: constructors/Encode/DecodeExpr/Canonicalize/Equal. gr: GRExpr
constructors + ToPhysMath() subset projection + ErrUnrepresentableProjection.

Dependencies:
phys-math side: phys-artifact only. gr side: phys-math (projection target).

Why it does not belong in phys-math:
Sin/Cos/UnknownFunction/order caps are GR-workload bounds, not mathematics.

Why it does not belong in phys-lib/gr:
Recursive canonical tree mechanics and byte-identity are the mathematical
common denominator reused by every framework.
