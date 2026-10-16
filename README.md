# PhysLib — physical-framework semantics

`github.com/PithomLabs/phys-lib` adds physical meaning on top of
mathematical representation. Where
[PhysMath](https://github.com/PithomLabs/phys-math) represents an
expression without knowing what it denotes, PhysLib records what a
framework takes it to mean: assumptions, definitions, equations,
theorems, and interpretations. Mathematical representation and
physical meaning stay in different layers by design.

## How it is organized

- **`core/`** — intentionally minimal (currently only `doc.go`).
  Nothing is promoted here for convenience: an abstraction belongs
  in the core only with evidence that it is framework-independent,
  stable across theories, and reusable. When in doubt, meaning
  stays in the specific framework package.
- **`gr/`** — General Relativity, the first conforming framework:
  framework identity, assumptions (convention, regularity, model),
  definitions, equations, theorems, and interpretations, plus
  reference fixtures (`maxwellbox`, `equivariance`, `ltb`).
- **`bohm/`** — Bohmian-mechanics framework semantics, including
  the `minisuperspace` reference fixture described below.
- **`classifications/`** — supporting classification notes (see its
  own `README.md`).
- **`internal/`** — shared test conformance helpers, not physical
  content.

A framework package may keep local representations where PhysMath
cannot express the mathematics, provided the projection boundary
is explicit and fallible. Framework-local meaning is never pushed
into PhysMath or the core to make one fixture convenient.

## What a physical derivation looks like

`bohm/minisuperspace` is the established reference: structured
mathematical artifacts minted through PhysMath/PhysArtifact APIs,
explicit derivation steps with replayed evidence, independently
constructed goldens with anti-goldens, assumption and condition
linkage, a `Verify()` entry point covering the claimed evidence,
a strict `manifest.json`, and documented limitations. It is a
symmetry-reduced toy model with executed fragments — not a proof
of the full theory, and its suite passing does not make its physics
true.

`gr/ltb` is a second example: a research-workflow fixture whose
executed fragments replay while its main theorem (backreaction
non-identifiability) remains explicitly open, with unresolved proof
obligations recorded rather than hidden.

## How it fits into Phys

```text
phys-lib/<framework>   this repository: physical meaning
        ↓ references pure mathematics
phys-math              representation + bounded execution
        ↓ encodes through
phys-artifact          canonical bytes + identity
```

[PhysArtifact](https://github.com/PithomLabs/phys-artifact) supplies
identity and encoding;
[PhysMath](https://github.com/PithomLabs/phys-math) supplies
theory-neutral representation and execution;
[PhysTrace](https://github.com/PithomLabs/phys-trace) coordinates
paper-derivation workflows. Related repositories:
[phys-artifact](https://github.com/PithomLabs/phys-artifact),
[phys-math](https://github.com/PithomLabs/phys-math).

## How do I start?

```bash
go build ./...
go test ./gr/ltb/...
```

Read `gr/doc.go` and `bohm/doc.go` for framework responsibilities,
then `bohm/minisuperspace/fixture.go` with its `manifest.json` and
`protocol/derivation_test.go` for the derivation pattern.
Contributor guidance: this repository has no root `AGENTS.md`;
package-local instructions apply per package (for example,
`gr/equivariance/AGENTS.md`), and cross-repository workflow rules
live in the collective
[phys-system manual](https://github.com/PithomLabs/phys-system/blob/main/AGENTS.md).

## How should new derivations be formalized?

Under the normative
[Scientific Derivation Specification Guidelines](https://github.com/PithomLabs/phys-trace/blob/main/DERIVATION_SPEC_GUIDELINES.md):
pin the source, map each quantity to its defining relation, keep
assumptions explicit and propagated, derive test expectations
independently of the implementation, bind provenance to exact
artifact references, and state what the evidence does and does not
establish. A package must preserve the assumptions and semantics
of its scientific source. Canonical encoding, hashes, replay,
structural validation, and passing tests establish
representational facts — never physical truth or empirical
confirmation.


## License 

MIT