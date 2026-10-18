# PROTOCOL GAPs — phys-lib manifest schema deviations

Authority: `PHYSMATH_PHYS_PHYSSTAR_REPRESENTATION_SPEC_v1.2.md` §1.4.

> The package-level `manifest.json` MUST use the exact schema and loading
> convention of the existing General Relativity example. … If the existing GR
> protocol cannot express the Phase-1 fixture without inventing a new semantic
> package-protocol field, the coding agent MUST report a `PROTOCOL GAP` rather
> than silently extending the protocol.

The normative GR schema is the top-level key set of
`phys-gr/solution/manifest.json` at frozen revision `ea5ec2de` (16 keys). The
`gr/equivariance` fixture — the Phase-1 target — conforms exactly and is the
only CONFORMANT manifest. Every other phys-lib fixture deviates; each
deviation is recorded below and machine-enforced by
`internal/manifestschema/manifest_conformance_test.go`.

**A registered gap is not conformance.** A fixture with any open gap below is
reported `NONCONFORMANT`, never "conformant with exceptions." The gap record
makes the deviation visible and accountable; it does not resolve it. Each gap
closes only through the permitted protocol (extending the GR manifest schema
via an authorized decision) or an explicitly authorized architectural change.

Corrections applied before registering these gaps were limited to values the
fixtures' own content establishes (`anomalies: []`, `coordinates` from a real
chart). No scientific content was invented to fill a GR-shaped field: an empty
`anomalies` list means *no anomalies observed in the fixture's declared,
executed scope* — it does not assert that the broader scientific model is
problem-free. Missing source verification, unsupported derivation steps, and
metadata-integrity limits live in each manifest's `limitations` and in these
gap records, not in `anomalies`.

## GAP-EXPECTED-RESULT

- **Fixtures:** `gr/ltb`, `bohm/minisuperspace`, `gr/maxwellbox`.
- **Deviation:** GR baseline key `expected_result` is absent.
- **Why the GR protocol cannot express it:** GR's `expected_result` is an
  executable canonical expression compared against a replayed derivation — it
  asserts an *executed, verified* closed-form result. LTB's `target` is an
  explicitly unexecuted research claim (obligations O1–O7 block it); bohm's
  results are per-quantity evidence pieces, not one closed form; maxwellbox's
  K1 endpoint fragments are replay checks. Encoding any of these as
  `expected_result` would falsify execution status.
- **Resolution status:** open. Closing requires either an executed closed form
  per fixture or an authorized decision that the GR schema admits a
  represented-only result field.

## GAP-FALSIFICATION

- **Fixtures:** `gr/ltb`, `bohm/minisuperspace`, `gr/maxwellbox`.
- **Deviation:** GR baseline key `falsification_conditions` is absent.
- **Why the GR protocol cannot express it:** each fixture already enforces its
  rejection criteria as *executable* tests (LTB anti-goldens and claim-status
  checks; bohm `Verify` corruption tests and historical wrong-path goldens;
  maxwellbox unsupported-step assertion). A manifest prose array would be a
  weaker, parallel copy of a stronger machine check. Migrating them into the
  GR field is a protocol redesign, not a metadata edit.
- **Resolution status:** open. Favored resolution is extending the GR protocol
  to reference executable falsifiers rather than duplicating them as prose.

## GAP-PREMISES

- **Fixture:** `gr/ltb`.
- **Deviation:** GR baseline key `premises` is absent.
- **Why the GR protocol cannot express it:** LTB's `assumptions` are already
  bound to fixture artifacts (each `ltb_*` id resolves). A prose `premises`
  array would duplicate content the artifact graph already owns, splitting the
  source of truth.
- **Resolution status:** open. Closing requires an authorized decision on
  whether premises must be artifact-bound prose or may be omitted where
  assumptions already bind.

## GAP-COORD-MAXWB

- **Fixture:** `gr/maxwellbox`.
- **Deviation:** GR baseline key `coordinates` is absent.
- **Why the GR protocol cannot express it:** the box fixture mints no
  `coordinate_system/1` artifact. `gr/ltb` (`["r"]`) and `bohm`
  (`["x","y","z"]`) were corrected to carry `coordinates` because their charts
  exist; maxwellbox has no chart, and any value would be invented.
- **Resolution status:** open. Closing requires minting a real chart artifact
  for the Rindler/box geometry, or an authorized decision that schematic
  geometry needs no coordinates field.

## GAP-SOURCE-IDENTITY

- **Fixtures:** `bohm/minisuperspace` (key `source` present — declared extra);
  `gr/maxwellbox` (key `source` relocated to `gr/maxwellbox/source.json`).
- **Deviation:** GR schema has no paper-identity field (arXiv id, PDF hash,
  representation lineage, scope note).
- **Why the GR protocol cannot express it:** the GR precedent records no
  source identity anywhere; there is no slot to reuse.
- **Relocation asymmetry:** bohm's `source` **stays in the manifest** because
  frozen `phys-trace` (`cmd/phystrace-mcp/tools.go`) parses
  `source.arxiv_id`/`source.scope_note` from that exact path — relocating it
  would break an out-of-scope consumer. maxwellbox has no such consumer, so
  its `source` was moved to a sidecar, removing an avoidable deviation.
- **Resolution status:** open. Closing requires an authorized GR-schema
  extension for paper identity; bohm's in-manifest `source` remains a
  necessary compatibility exception until then.

## GAP-SOURCE-MAPPING

- **Fixture:** `bohm/minisuperspace` (key `source_mapping` relocated to
  `bohm/minisuperspace/source_mapping.json`).
- **Deviation:** GR schema has no field for equation-level
  source-to-fixture correspondence.
- **Why the GR protocol cannot express it:** the GR precedent has no
  equation-mapping structure; `premises` carry statements, not a machine-checkable
  equation ↔ quantity ↔ fixture-ref mapping.
- **Enforcement:** the sidecar is validated by
  `TestSourceMappingStructuralConsistency` (refs resolve, equations in scope,
  required fields present). The mapping is explicitly *structural metadata*,
  not source-verified scientific correspondence — see the
  `source_mapping_structural_only` limitation in bohm's manifest.
- **Resolution status:** open. Relocated (zero Go consumers read it in place;
  frozen phys-trace reads only `name`, `source.*`, `regime`, `premises`,
  `limitations`). Closing requires an authorized GR-schema extension, or
  Project A's correspondence machinery (v1.5 §21.3).

## GAP-CLAIM-STATUS

- **Fixture:** `gr/ltb` (keys `claims`, `obligations` relocated to
  `gr/ltb/research_metadata.json`).
- **Deviation:** GR schema has no epistemic-status or obligation-tracking field.
- **Why the GR protocol cannot express it:** GR's `anomalies`/
  `falsification_conditions` describe a derivation's behavior; they cannot
  carry "this claim is an open conjecture and these obligations block it."
  This is the load-bearing epistemic-status machinery the whole LTB honesty
  story rests on (`StatusOpenConjecture`, `CheckClaimPromotion`).
- **Resolution status:** open. Relocated (no external consumer). Closing
  requires an authorized GR-schema extension for claim status.

## GAP-PROVENANCE

- **Fixture:** `gr/ltb` (keys `literature`, `provenance_note` relocated to
  `gr/ltb/research_metadata.json`).
- **Deviation:** GR schema has no literature-citation or provenance-narrative field.
- **Why the GR protocol cannot express it:** the GR precedent records no
  citations or provenance; `premises` are statements, not
  `cite`/`establishes`/`status` triples.
- **Resolution status:** open. Relocated. Closing requires an authorized
  GR-schema extension for literature/provenance.

## GAP-TARGET

- **Fixture:** `gr/ltb` (keys `target`, `domain` relocated to
  `gr/ltb/research_metadata.json`).
- **Deviation:** GR schema has no unexecuted-target or physical-domain field.
- **Why the GR protocol cannot express it:** GR's `expected_result` asserts an
  executed result (see GAP-EXPECTED-RESULT); LTB's `target` is an unexecuted
  claim and `domain` is a physical-domain description. Neither can be placed
  in `expected_result` without falsifying status, and no other GR field fits.
- **Resolution status:** open. Relocated. Closing requires an authorized
  GR-schema extension, coupled to how GAP-EXPECTED-RESULT closes.
