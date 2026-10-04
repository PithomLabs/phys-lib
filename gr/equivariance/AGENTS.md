# AGENTS.md — Equivariance Reference Fixture (phys-lib/gr)

This package is the canonical Phase-1 equivariance-theorem reference fixture,
owned by `phys-lib/gr`. It follows the GR protocol precedent
(`phys-gr/solution/`: embed manifest, strict typed decode, ordered executable
derivation, firewalls, mutation negatives) in the canonical nested layout:

```text
phys-lib/gr/equivariance/
    AGENTS.md
    manifest.json
    protocol/
        derivation_test.go
```

Flat (`phys-gr/solution/`) vs nested (here) is a packaging relocation only.
Schema, loading, and execution semantics are identical; this note records the
equivalence. The original `phys-gr/solution/` is preserved unchanged as the
historical reference and is never a premise of this fixture.

## Fixture contract (exact, no substitutes)

```text
f(v)       = 2 * v        (1 parameter, display "v")
rho_X(g,v) = g * v        (2 parameters, displays "g", "v")
rho_Y(g,v) = g * v        (2 parameters, displays "g", "v")
```

Six transformations, in order: APPLY, APPLY, CANON, APPLY, APPLY, CANON.
Staged `LHS_1`/`RHS_1` Calls are pre-constructed fixture data with byte-linkage
checks, because EXPR-APPLY-001 is root-Call-only and Phase 1 has no generic
AST subtree-replacement rule.

## Firewalls (the test fails if violated)

```text
manifest.json is metadata only — never imported by fixture.go, never a premise
no distributivity, factorization, or x+x -> 2x anywhere in the derivation
no Map-for-f, no durable FunctionApplication staged sides
no GroupAction sort, no trig content, no "mu" symbols
every durable free symbol carries its declaring Function's ParameterID
rho_X/rho_Y canonical coincidence is intentional and documented in fixture.go
```

## Semantic contract (claim-to-conclusion trace)

The claimed proposition is `f(rho_X(g,x)) = rho_Y(g,f(x))`, formally LHS0/RHS0.
The persisted theorem must preserve a traceable correspondence from the
conclusion sides to those proposition sides. Normal-form equality
(`Step2 == Step5`) is necessary but never sufficient on its own.

The verifier (`Verify` + `VerifyClaim` + `protocol/derivation_test.go`)
establishes:

```text
LHS0 → innerRho premise → Step0 → staged LHS1 → Step1 → Step2 → conclusion.LHS
RHS0 → innerFx premise  → Step3 → staged RHS1 → Step4 → Step5 → conclusion.RHS
```

via structural decomposition, byte-identity linkage, replayed rule
execution, and the conclusion gate. No theorem-side artifact is
semantically dead: corrupting LHS0/RHS0 or any linkage while holding the
finals fixed fails verification.

## Mutation scope

Skipped steps, swapped stages, wrong declaring Functions, altered display
names, and foreign conclusion bytes must each fail replay or the conclusion
gate. A mutation that passes is a fixture bug, not tolerance.
