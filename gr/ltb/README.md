# ZEH/LTB research-workflow reference

Working title: *Spatial Backreaction and Light-Cone Observables in LTB
Cosmology: An Identifiability Analysis.*

This package formalizes the investigation; it does not settle it.
ZEH-Q-03 (non-identifiability) is an open conjecture with no theorem
artifact. A green suite proves the executed fragment replays and the
open status is preserved — never that the LTB theorem is proved.

## Question

Given an admissible LTB class and a fixed past-light-cone distance
observation, what does it determine about present-day Buchert
backreaction `Q_D(t0)` on a specified comoving domain? Backreaction is
identifiable iff `Q[D_A*] = {Q_D^M(t0)}` is a singleton under the
stated class, foliation, domain, and normalization.

## What is executed

- `dM/dr` for `M(r) = r^3/2` via DIFFERENTIATE-001, replayed, bound to
  the hand golden `3r^2/2` through a stored EQ relation and theorem
  conclusion gate.
- EXPR-CANON-001 on the EdS `theta = H + 2H` construction, replayed.
- Go rational corroboration (pointwise, not symbolic proof) of the EdS
  cone identity and the series coefficients below.

## What is represented only

Spatial averaging and the volume measure (no integration primitive was
added — growth gate), sigma vanishing beyond its factored structure,
the off-cone extension, the `Q_2` functional, and all claim statuses.

## Hand derivations (independently re-derived, not copied)

Conventions: `M(r) = r^3/2` exactly (never `r^{3/2}`); area radius
`R = D_A` on the central cone; `R'(t0,0) = 1`, `t0 = 2/3` as declared
constitution.

1. **Target series.** `(1+z)^-1/2 = 1 - z/2 + 3z^2/8 - 5z^3/16 + ...`
   (binomial). Numerator `2[1 - ...] = z - 3z^2/4 + 5z^3/8 - ...`.
   Dividing by `(1+z)` via `(1 - z + z^2 - ...)` gives
   `D_A*(z) = z - 7z^2/4 + 19z^3/8 + ...`. The `19/8` coefficient
   corrects the earlier `23/8` slip caught by the composition check.
2. **EdS vertex profile.** `Z0(r) = (1-r/2)^-2 = 1 + r + 3r^2/4 +
   r^3/2 + ...`, so `z(r) = r + 3r^2/4 + ...`: `z_2 = 3/4` (data-fixed
   under the central normalization). The `r^3` coefficient `1/2`
   verifies the `beta = 0` case of the candidate recursion
   `z_3 = 1/2 + 2*beta/3`; the `2*beta/3` slope itself stays a
   candidate and is not used as a golden.
3. **Cone identity.** With `u = 1-r/2`, `V0 = r/u`, `Q0 = 1/u`,
   `Rc0 = r*u^2`, `P0 = u^2`: LHS `2*V0*Q0 = 2r/u^2`; RHS
   `3r^2/Rc0 - r^3*P0/Rc0^2 = 3r/u^2 - r/u^2 = 2r/u^2`. Holds
   identically for `u != 0`.
4. **Gauge derivative.** `d/dr[r^3/2] = 3r^2/2` (monomial power rule).

## Claim register

- ZEH-Q-01: literature claim requiring verification (MHE background).
- ZEH-Q-02: candidate construction, incomplete (O1, O2, O3 open).
- ZEH-Q-03: open conjecture (O1, O3, O4, O5, O6, O7 open). No
  function-counting inference: metric freedom does not imply `Q_D`
  variance until `Q_2[h] != 0` is established under the exact
  constraint. `CheckClaimPromotion` only rejects a `proved` status
  with missing or out-of-inventory evidence; inventory membership
  does not establish that a derivation proves the promoted claim.
- ZEH-Q-04: deferred branch, out of this slice.

## Stored EdS vertex coefficients

`VertexCoeffs` maps series order to rational artifacts
(`1->1`, `2->3/4`, `3->1/2`) for `z(r) = Z0(r)-1`. These are
finite-order baseline values with independent mechanical
recomputation — not the general candidate recursion
`z_3 = 1/2 + 2*beta/3`, analytic existence, or any ZEH-Q claim.

## Variable-identity note

The production `H` is an ephemeral symbol: its encoded identity is
its display string, with no `ParameterID`. Tests assert both
operands match the same intended stored representation
(byte-identical encodings) — syntactic sameness, not
bound-parameter identity and not physical meaning. A synthetic
bound-symbol control shows same-display/different-`ParameterID`
symbols are distinguished; it does not confer bound-parameter
identity on the production artifact.

## Layout

`fixture.go` (`Build` + `Verify`), `claim.go` (metadata consistency
check), `manifest.json` (strict machine metadata), `manifest_test.go`
(schema + status pins), `protocol/derivation_test.go` (execution,
goldens, anti-goldens, provenance, coverage, determinism).
