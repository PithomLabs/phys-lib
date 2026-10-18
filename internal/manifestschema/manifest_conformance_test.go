package manifestschema

// Manifest conformance test.
//
// v1.2 §1.4 requires every formalization package's manifest.json to use the
// exact schema of the General Relativity example, and to report a PROTOCOL
// GAP rather than silently extend the protocol. This test enforces that
// contract mechanically:
//
//  1. Any key in a manifest that is not in the GR baseline and not declared
//     as a PROTOCOL GAP fails, naming the fixture and the offending key.
//  2. Any GR baseline key missing from a manifest and not declared as a gap
//     fails, naming the fixture and the missing key.
//  3. Every declared gap must map to a real deviation, and every deviation
//     must have a gap record — checked bidirectionally against
//     PROTOCOL_GAPS.md.
//  4. A fixture with any open gap is reported NONCONFORMANT, never
//     "conformant with exceptions." Only a fixture with zero deviations is
//     CONFORMANT.
//
// Scope: this test enforces TOP-LEVEL key sets only. Nested value-form
// conformance (the shape of coordinates, premises, expected_result, …) is
// enforced by each package's own manifest_test.go against the GR precedent;
// this test does not claim to verify nested structure. Registered gaps are
// not conformance: they make a known deviation visible and accountable, and
// the affected fixture remains nonconformant until the gap is resolved
// through the permitted protocol or an explicitly authorized architectural
// decision.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// grBaselineKeys is the exact top-level key set of the normative GR
// manifest, phys-gr/solution/manifest.json at frozen revision ea5ec2de.
// phys-lib must not import frozen phys-gr (dependency boundary), so the key
// set is pinned as a literal with its provenance recorded here.
var grBaselineKeys = []string{
	"anomalies",
	"artifact_type",
	"assumptions",
	"conventions",
	"coordinates",
	"derivation_id",
	"entry_points",
	"expected_result",
	"falsification_conditions",
	"framework_id",
	"limitations",
	"name",
	"premises",
	"regime",
	"schema_version",
	"verification",
}

// gap records one declared deviation. Fields:
//
//	Key     — the manifest key that deviates (extra, or missing from baseline).
//	GapID   — the PROTOCOL_GAPS.md section that documents it.
//	Reason  — short human-readable justification (mirrors the doc).
//	Sidecar — for relocated metadata: the sidecar file that now carries the
//	          content, relative to the phys-lib root. Empty if none.
type gap struct {
	Key     string
	GapID   string
	Reason  string
	Sidecar string
}

// fixtureConformance declares one manifest's expected deviations.
type fixtureConformance struct {
	// Manifest is the manifest path relative to the phys-lib root.
	Manifest string
	// Extra lists keys present in the manifest but absent from the GR
	// baseline.
	Extra []gap
	// Missing lists GR baseline keys deliberately absent from this manifest.
	Missing []gap
	// Relocations lists research metadata moved out of the manifest into a
	// sidecar because the GR schema has no field for it. Each is a registered
	// gap; the manifest key itself is gone, but the gap stays open.
	Relocations []gap
}

// registry declares every known, reviewed deviation. Adding a manifest key
// that is not here fails TestManifestConformance; adding a gap here requires
// a matching PROTOCOL_GAPS.md section (enforced bidirectionally).
var registry = map[string]fixtureConformance{
	"gr/equivariance/manifest.json": {
		Manifest: "gr/equivariance/manifest.json",
		// Phase-1 target: conforms exactly to the GR baseline.
	},
	"gr/ltb/manifest.json": {
		Manifest: "gr/ltb/manifest.json",
		Missing: []gap{
			{Key: "expected_result", GapID: "GAP-EXPECTED-RESULT",
				Reason: "GR expected_result is an executed, verified closed-form result; LTB's target is an explicitly unexecuted research claim, so encoding it there would falsify status"},
			{Key: "falsification_conditions", GapID: "GAP-FALSIFICATION",
				Reason: "rejection criteria are enforced as executable anti-golden tests, not a manifest array; migrating them is protocol redesign"},
			{Key: "premises", GapID: "GAP-PREMISES",
				Reason: "LTB assumptions are already bound to fixture artifacts; a prose premises array would duplicate them"},
		},
		Relocations: []gap{
			{Key: "claims", GapID: "GAP-CLAIM-STATUS",
				Reason:  "epistemic claim-status tracking has no GR field; moved to research_metadata.json",
				Sidecar: "gr/ltb/research_metadata.json"},
			{Key: "obligations", GapID: "GAP-CLAIM-STATUS",
				Reason:  "derivation obligations have no GR field; moved to research_metadata.json",
				Sidecar: "gr/ltb/research_metadata.json"},
			{Key: "literature", GapID: "GAP-PROVENANCE",
				Reason:  "literature citations with verification status have no GR field; moved to research_metadata.json",
				Sidecar: "gr/ltb/research_metadata.json"},
			{Key: "provenance_note", GapID: "GAP-PROVENANCE",
				Reason:  "provenance narrative has no GR field; moved to research_metadata.json",
				Sidecar: "gr/ltb/research_metadata.json"},
			{Key: "target", GapID: "GAP-TARGET",
				Reason:  "unexecuted research target is not an executed expected_result; moved to research_metadata.json",
				Sidecar: "gr/ltb/research_metadata.json"},
			{Key: "domain", GapID: "GAP-TARGET",
				Reason:  "physical domain description has no GR field; moved to research_metadata.json",
				Sidecar: "gr/ltb/research_metadata.json"},
		},
	},
	"bohm/minisuperspace/manifest.json": {
		Manifest: "bohm/minisuperspace/manifest.json",
		Extra: []gap{
			{Key: "source", GapID: "GAP-SOURCE-IDENTITY",
				Reason: "GR schema has no paper-identity field; frozen phys-trace MCP reads source.* at this path, so it cannot be relocated"},
		},
		Missing: []gap{
			{Key: "expected_result", GapID: "GAP-EXPECTED-RESULT",
				Reason: "GR expected_result is an executed closed-form result; bohm's results are per-quantity evidence, not one closed form"},
			{Key: "falsification_conditions", GapID: "GAP-FALSIFICATION",
				Reason: "rejection criteria are enforced as executable anti-golden/Verify tests, not a manifest array"},
		},
		Relocations: []gap{
			{Key: "source_mapping", GapID: "GAP-SOURCE-MAPPING",
				Reason:  "equation-to-fixture correspondence has no GR field; moved to source_mapping.json and validated by TestSourceMappingStructuralConsistency",
				Sidecar: "bohm/minisuperspace/source_mapping.json"},
		},
	},
	"gr/maxwellbox/manifest.json": {
		Manifest: "gr/maxwellbox/manifest.json",
		Missing: []gap{
			{Key: "coordinates", GapID: "GAP-COORD-MAXWB",
				Reason: "no coordinate_system artifact exists for the box fixture; any value would be invented"},
			{Key: "expected_result", GapID: "GAP-EXPECTED-RESULT",
				Reason: "GR expected_result is an executed closed-form result; the K1 endpoint fragments are replay checks, not one closed form"},
			{Key: "falsification_conditions", GapID: "GAP-FALSIFICATION",
				Reason: "rejection criteria are enforced as executable tests, not a manifest array"},
		},
		Relocations: []gap{
			{Key: "source", GapID: "GAP-SOURCE-IDENTITY",
				Reason:  "GR schema has no paper-identity field; moved to source.json (no consumer reads it in place)",
				Sidecar: "gr/maxwellbox/source.json"},
		},
	},
}

// relocatedSidecars declares metadata moved out of a manifest into a sidecar
// because the GR schema has no field for it. Each relocation is a registered
// gap; the conformance test asserts the sidecar exists and is non-empty.
var relocatedSidecars = map[string][]string{
	"gr/ltb/research_metadata.json": {
		"gr/ltb/manifest.json",
	},
	"bohm/minisuperspace/source_mapping.json": {
		"bohm/minisuperspace/manifest.json",
	},
	"gr/maxwellbox/source.json": {
		"gr/maxwellbox/manifest.json",
	},
}

// physLibRoot returns the phys-lib repository root via runtime.Caller.
func physLibRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unavailable")
	}
	// .../phys-lib/internal/manifestschema/<this file> -> up three levels.
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// readManifestKeys returns the top-level key set of the manifest at path.
func readManifestKeys(t *testing.T, path string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest %s: %v", path, err)
	}
	keys, err := topLevelKeys(raw)
	if err != nil {
		t.Fatalf("parse manifest %s: %v", path, err)
	}
	set := map[string]bool{}
	for _, k := range keys {
		set[k] = true
	}
	return set
}

// checkConformance compares an actual key set against the GR baseline plus a
// declared fixture, returning violations that name the offending key. It is
// a pure function so the corruption drills can exercise it directly.
func checkConformance(actual map[string]bool, fc fixtureConformance) []string {
	var violations []string

	declaredExtra := map[string]bool{}
	for _, g := range fc.Extra {
		declaredExtra[g.Key] = true
	}
	declaredMissing := map[string]bool{}
	for _, g := range fc.Missing {
		declaredMissing[g.Key] = true
	}

	baseline := map[string]bool{}
	for _, k := range grBaselineKeys {
		baseline[k] = true
	}

	// Every actual key must be in the baseline or declared as an extra gap.
	var actualList []string
	for k := range actual {
		actualList = append(actualList, k)
	}
	sort.Strings(actualList)
	for _, k := range actualList {
		if !baseline[k] && !declaredExtra[k] {
			violations = append(violations, fmt.Sprintf(
				"undeclared extra key %q (not in GR baseline, not a registered PROTOCOL GAP)", k))
		}
		if baseline[k] && declaredExtra[k] {
			violations = append(violations, fmt.Sprintf(
				"declared-extra key %q is in the GR baseline; remove it from the gap registry", k))
		}
	}

	// Every GR baseline key must be present or declared missing.
	var baselineList []string
	for k := range baseline {
		baselineList = append(baselineList, k)
	}
	sort.Strings(baselineList)
	for _, k := range baselineList {
		if !actual[k] && !declaredMissing[k] {
			violations = append(violations, fmt.Sprintf(
				"undeclared missing key %q (in GR baseline, not a registered PROTOCOL GAP)", k))
		}
		if actual[k] && declaredMissing[k] {
			violations = append(violations, fmt.Sprintf(
				"declared-missing key %q is present; remove it from the gap registry", k))
		}
	}

	// Every declared gap must map to a real deviation.
	for _, g := range fc.Extra {
		if !actual[g.Key] {
			violations = append(violations, fmt.Sprintf(
				"gap %s declares extra key %q but it is not present", g.GapID, g.Key))
		}
	}
	for _, g := range fc.Missing {
		if actual[g.Key] {
			violations = append(violations, fmt.Sprintf(
				"gap %s declares missing key %q but it is present", g.GapID, g.Key))
		}
	}

	return violations
}

// TestManifestConformance enforces the GR exact-schema contract across every
// phys-lib manifest, fails on any undeclared deviation, and reports each
// fixture's true conformance status.
func TestManifestConformance(t *testing.T) {
	root := physLibRoot(t)
	var report []string
	conformantCount := 0

	// Iterate in sorted order for a stable report.
	var names []string
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fc := registry[name]
		actual := readManifestKeys(t, filepath.Join(root, name))
		violations := checkConformance(actual, fc)
		for _, v := range violations {
			t.Errorf("%s: %v", name, v)
		}

		gapCount := len(fc.Extra) + len(fc.Missing) + len(fc.Relocations)
		status := "CONFORMANT"
		if gapCount > 0 {
			status = fmt.Sprintf("NONCONFORMANT (%d declared gap(s): %s)",
				gapCount, gapSummary(fc))
		} else {
			conformantCount++
		}
		report = append(report, fmt.Sprintf("  %-42s %s", name, status))

		// A relocation means the key left the manifest and now lives in a
		// sidecar. Verify both halves: absent here, present there.
		for _, r := range fc.Relocations {
			if actual[r.Key] {
				t.Errorf("%s: relocated key %q is still present in the manifest (gap %s)", name, r.Key, r.GapID)
			}
			sidecarKeys := readManifestKeys(t, filepath.Join(root, r.Sidecar))
			if !sidecarKeys[r.Key] {
				t.Errorf("%s: relocated key %q missing from sidecar %s (gap %s)", name, r.Key, r.Sidecar, r.GapID)
			}
		}
	}

	t.Logf("manifest conformance status (GR exact-schema contract, v1.2 §1.4):")
	for _, line := range report {
		t.Log(line)
	}

	// Safeguard: a fixture with open gaps must never be reported conformant.
	// Only the Phase-1 equivariance target conforms exactly; assert that.
	if conformantCount != 1 || len(registry["gr/equivariance/manifest.json"].Extra) != 0 ||
		len(registry["gr/equivariance/manifest.json"].Missing) != 0 {
		t.Errorf("want exactly one CONFORMANT fixture (gr/equivariance), got %d", conformantCount)
	}
}

// gapSummary lists a fixture's gap IDs for the status line.
func gapSummary(fc fixtureConformance) string {
	seen := map[string]bool{}
	var ids []string
	for _, g := range append(append(append([]gap{}, fc.Extra...), fc.Missing...), fc.Relocations...) {
		if !seen[g.GapID] {
			seen[g.GapID] = true
			ids = append(ids, g.GapID)
		}
	}
	sort.Strings(ids)
	return strings.Join(ids, ", ")
}

// TestSidecarsExist asserts every declared relocation sidecar exists and is
// non-empty, so metadata moved out of a manifest was not silently dropped.
func TestSidecarsExist(t *testing.T) {
	root := physLibRoot(t)
	for sidecar := range relocatedSidecars {
		info, err := os.Stat(filepath.Join(root, sidecar))
		if err != nil {
			t.Errorf("relocation sidecar %s missing: %v", sidecar, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("relocation sidecar %s is empty", sidecar)
		}
	}
}

// TestGapRegistryMatchesDoc ties the machine registry to PROTOCOL_GAPS.md
// bidirectionally: every gap ID referenced by the registry must have a doc
// section, and every doc section must be referenced by the registry. This is
// what makes a registered gap more than a note — the prose and the enforced
// registry cannot drift apart.
func TestGapRegistryMatchesDoc(t *testing.T) {
	root := physLibRoot(t)
	docRaw, err := os.ReadFile(filepath.Join(root, "PROTOCOL_GAPS.md"))
	if err != nil {
		t.Fatalf("read PROTOCOL_GAPS.md: %v", err)
	}
	docIDs := docGapIDs(string(docRaw))

	regIDs := map[string]bool{}
	for _, fc := range registry {
		for _, g := range append(append(append([]gap{}, fc.Extra...), fc.Missing...), fc.Relocations...) {
			regIDs[g.GapID] = true
		}
	}

	for id := range regIDs {
		if !docIDs[id] {
			t.Errorf("gap %s is registered in the conformance test but has no section in PROTOCOL_GAPS.md", id)
		}
	}
	for id := range docIDs {
		if !regIDs[id] {
			t.Errorf("PROTOCOL_GAPS.md documents %s but the conformance registry does not reference it", id)
		}
	}
	if len(regIDs) == 0 {
		t.Fatal("gap registry is empty: expected declared deviations")
	}
}

// docGapIDs extracts the gap section identifiers (## GAP-…) from the report.
func docGapIDs(doc string) map[string]bool {
	ids := map[string]bool{}
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			fields := strings.Fields(strings.TrimPrefix(line, "## "))
			if len(fields) > 0 && strings.HasPrefix(fields[0], "GAP-") {
				ids[fields[0]] = true
			}
		}
	}
	return ids
}
