# Classification records (plan9.3 §6, D19)

Every material `phys-gr` abstraction re-expressed in canonical modules gets
exactly one classification:

```text
PHYS-MATH | PHYS-LIB/CORE | PHYS-LIB/GR | SPLIT | REPLACE | DELETE | DEFER
```

## Record format

One file per refactoring-map row:

```text
phys-lib/classifications/<slug>.md
```

`<slug>` is derived from the source component (e.g. `symbolic-expr.md`).
Committed with the PR that moves the abstraction. Front-matter first:

```text
---
id: <stable-slug>
source: <phys-gr path, e.g. symbolic/normalize.go>
class: <one of the seven classes>
target: <canonical destination, e.g. phys-math rules / phys-lib/gr>
status: adopted
---

Why it is theory-agnostic:
Known/current consumers:
Potential future consumers:
Theory-specific assumptions:
Stable interface:
Dependencies:
Why it does not belong in phys-math:
Why it does not belong in phys-lib/gr:
```

Rules:

- `PHYS-LIB/CORE` requires every evidence field non-empty (CI-enforced).
- `SPLIT` requires the exact ownership/interface boundary named in the body.
- All other classes require classification + destination rationale.
- Ambiguity defaults to `phys-lib/gr`.

These records are documentation/governance artifacts. They are NOT exported
API and are NOT part of `package core`.

## Verification

`check_test.go` (package `classifications_test`, test-only) enforces:

1. every `*.md` (except this README) parses with complete front-matter;
2. every `PHYS-LIB/CORE` record has all evidence fields non-empty;
3. `package core` exports zero identifiers (empty-core assertion);
4. no production import path reaches `internal/conformance` (reserved).
