package minisuperspace

// Manifest/source-mapping consistency tests.
//
// These are STRUCTURAL checks only. They verify that the intended
// paper-equation-to-fixture correspondence recorded in source_mapping.json
// is internally consistent: every fixture_ref names a symbol the fixture
// actually exposes, every source_eq lies in the fixture's declared scope,
// and no required field is empty. They do NOT, and cannot, verify that any
// mapping is scientifically correct against the paper's text — that is
// analysis-layer work outside this fixture's executed scope (see the
// source_mapping_structural_only limitation in manifest.json).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// mappingEntry is one source_mapping record. Every field is required and
// must be non-empty; dependencies must name at least one fixture symbol.
type mappingEntry struct {
	SourceEq         string   `json:"source_eq"`
	SourceQuantity   string   `json:"source_quantity"`
	Meaning          string   `json:"meaning"`
	DefiningRelation string   `json:"defining_relation"`
	FixtureRef       string   `json:"fixture_ref"`
	Dependencies     []string `json:"dependencies"`
}

// mappingDoc wraps the sidecar array so an unknown top-level key is caught.
type mappingDoc struct {
	SourceMapping []mappingEntry `json:"source_mapping"`
}

// limitEntry is one manifest limitation record.
type limitEntry struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
}

// manifestLimits is the subset of manifest.json these tests read.
type manifestLimits struct {
	Limitations []limitEntry `json:"limitations"`
}

func readJSONStrict(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("strict decode %s: %v", path, err)
	}
}

func sidecarPath(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unavailable")
	}
	return filepath.Join(filepath.Dir(file), name)
}

// loadMapping reads the relocated source_mapping.json sidecar.
func loadMapping(t *testing.T) mappingDoc {
	t.Helper()
	var doc mappingDoc
	readJSONStrict(t, sidecarPath(t, "source_mapping.json"), &doc)
	return doc
}

// readJSONLenient decodes a subset view of a larger document (e.g. reading
// only `limitations` out of manifest.json); unknown fields are expected.
func readJSONLenient(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

// loadManifestLimits reads the limitations block of manifest.json.
func loadManifestLimits(t *testing.T) manifestLimits {
	t.Helper()
	var m manifestLimits
	readJSONLenient(t, sidecarPath(t, "manifest.json"), &m)
	return m
}

// eqScopeBase is the Section-6 equation range declared by the manifest's
// source.scope_note ("Section 6 eqs 67-78").
var eqScopeBase = [2]int{67, 78}

// eqQualifiers is the registered set of source_eq suffix qualifiers
// ("-cov", "-conn", "-derived" plus the empty suffix). A new qualifier is a
// deliberate scope decision and must be added here, not silently accepted.
var eqQualifiers = []string{"", "-derived", "-cov", "-conn"}

var eqQualifierSet = func() map[string]bool {
	m := map[string]bool{}
	for _, q := range eqQualifiers {
		m[q] = true
	}
	return m
}()

// checkSourceEq validates one source_eq against the declared scope.
func checkSourceEq(eq string) error {
	if eq == "" {
		return fmt.Errorf("source_eq is empty")
	}
	baseStr := eq
	qual := ""
	for _, q := range eqQualifiers {
		if q != "" && strings.HasSuffix(eq, q) {
			baseStr = strings.TrimSuffix(eq, q)
			qual = q
			break
		}
	}
	if !eqQualifierSet[qual] {
		return fmt.Errorf("source_eq %q has unregistered qualifier (want one of %v)", eq, eqQualifiers)
	}
	if strings.Contains(baseStr, "-") {
		// A hyphen that no registered qualifier consumed is an
		// unregistered qualifier, not a malformed integer.
		return fmt.Errorf("source_eq %q has unregistered qualifier (want one of %v)", eq, eqQualifiers)
	}
	base, err := strconv.Atoi(baseStr)
	if err != nil {
		return fmt.Errorf("source_eq %q: base %q is not an integer", eq, baseStr)
	}
	if base < eqScopeBase[0] || base > eqScopeBase[1] {
		return fmt.Errorf("source_eq %q: base %d outside declared scope %d-%d", eq, base, eqScopeBase[0], eqScopeBase[1])
	}
	return nil
}

// fixtureUniverse returns every name a source_mapping fixture_ref or
// dependencies entry may legitimately cite: exported ArtifactRef struct
// fields, plus the Conds and Symbols map keys the fixture populates.
// Reflection keeps this in sync with the fixture automatically.
func fixtureUniverse(t *testing.T) map[string]string {
	t.Helper()
	f := Build()
	universe := map[string]string{}
	ft := reflect.TypeOf(f).Elem()
	for i := 0; i < ft.NumField(); i++ {
		if ft.Field(i).Type == reflect.TypeOf(f.RRef) {
			universe[ft.Field(i).Name] = "field"
		}
	}
	for k := range f.Conds {
		universe[k] = "condition"
	}
	for k := range f.Symbols {
		universe[k] = "symbol"
	}
	return universe
}

// splitRefs splits a comma-separated fixture_ref list and trims spaces.
func splitRefs(ref string) []string {
	parts := strings.Split(ref, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

var validName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// checkMapping performs the full structural consistency check and returns a
// list of violations, each naming the offending record index and field.
func checkMapping(entries []mappingEntry, universe map[string]string) []string {
	var violations []string
	seenEq := map[string]int{}
	for i, e := range entries {
		idx := i // report the sidecar index for traceability
		if err := checkSourceEq(e.SourceEq); err != nil {
			violations = append(violations, fmt.Sprintf("entry %d (source_eq %q): %v", idx, e.SourceEq, err))
		}
		if prev, dup := seenEq[e.SourceEq]; dup {
			violations = append(violations, fmt.Sprintf("entry %d: duplicate source_eq %q (also entry %d)", idx, e.SourceEq, prev))
		} else {
			seenEq[e.SourceEq] = idx
		}
		for _, field := range []struct {
			name string
			val  string
		}{
			{"source_quantity", e.SourceQuantity},
			{"meaning", e.Meaning},
			{"defining_relation", e.DefiningRelation},
			{"fixture_ref", e.FixtureRef},
		} {
			if strings.TrimSpace(field.val) == "" {
				violations = append(violations, fmt.Sprintf("entry %d (source_eq %q): required field %s is empty", idx, e.SourceEq, field.name))
			}
		}
		for _, ref := range splitRefs(e.FixtureRef) {
			if _, ok := universe[ref]; !ok {
				violations = append(violations, fmt.Sprintf("entry %d (source_eq %q): fixture_ref %q is not a known fixture symbol", idx, e.SourceEq, ref))
			}
		}
		if len(e.Dependencies) == 0 {
			violations = append(violations, fmt.Sprintf("entry %d (source_eq %q): dependencies is empty", idx, e.SourceEq))
		}
		for _, d := range e.Dependencies {
			if !validName.MatchString(d) {
				violations = append(violations, fmt.Sprintf("entry %d (source_eq %q): dependency %q is not a bare identifier", idx, e.SourceEq, d))
				continue
			}
			if _, ok := universe[d]; !ok {
				violations = append(violations, fmt.Sprintf("entry %d (source_eq %q): dependency %q is not a known fixture symbol", idx, e.SourceEq, d))
			}
		}
	}
	return violations
}

// TestSourceMappingStructuralConsistency asserts the relocated mapping is
// internally consistent: refs resolve, equations are in scope, required
// fields are present. It makes no scientific-correctness claim.
func TestSourceMappingStructuralConsistency(t *testing.T) {
	doc := loadMapping(t)
	if len(doc.SourceMapping) == 0 {
		t.Fatal("source_mapping is empty: the fixture declares Section 6 mappings")
	}
	universe := fixtureUniverse(t)
	if len(universe) == 0 {
		t.Fatal("fixture universe is empty")
	}
	for _, v := range checkMapping(doc.SourceMapping, universe) {
		t.Error(v)
	}
}

// TestSourceMappingRejectsCorruptions pins that each deliberate corruption
// is caught with a message naming the offending record and field. Each case
// mutates a copy; the pristine mapping must still pass (checked first).
func TestSourceMappingRejectsCorruptions(t *testing.T) {
	pristine := loadMapping(t).Mapping()
	universe := fixtureUniverse(t)
	if got := checkMapping(pristine, universe); len(got) != 0 {
		t.Fatalf("pristine mapping rejected: %v", got)
	}

	cases := []struct {
		name    string
		mutate  func([]mappingEntry)
		wantSub string
	}{
		{
			name: "unknown fixture_ref",
			mutate: func(es []mappingEntry) {
				es[0].FixtureRef = "NonexistentRef"
			},
			wantSub: `fixture_ref "NonexistentRef" is not a known fixture symbol`,
		},
		{
			name: "unknown dependency",
			mutate: func(es []mappingEntry) {
				es[0].Dependencies = append([]string{}, es[0].Dependencies...)
				es[0].Dependencies[0] = "GhostDep"
			},
			wantSub: `dependency "GhostDep" is not a known fixture symbol`,
		},
		{
			name: "out-of-scope equation identifier",
			mutate: func(es []mappingEntry) {
				es[0].SourceEq = "99"
			},
			wantSub: "outside declared scope 67-78",
		},
		{
			name: "non-integer equation base",
			mutate: func(es []mappingEntry) {
				es[0].SourceEq = "abc"
			},
			wantSub: "is not an integer",
		},
		{
			name: "unregistered qualifier",
			mutate: func(es []mappingEntry) {
				es[0].SourceEq = "76-extra"
			},
			wantSub: "unregistered qualifier",
		},
		{
			name: "missing required field (meaning)",
			mutate: func(es []mappingEntry) {
				es[0].Meaning = ""
			},
			wantSub: "required field meaning is empty",
		},
		{
			name: "missing dependencies",
			mutate: func(es []mappingEntry) {
				es[0].Dependencies = nil
			},
			wantSub: "dependencies is empty",
		},
		{
			name: "duplicate source_eq",
			mutate: func(es []mappingEntry) {
				es[1].SourceEq = es[0].SourceEq
			},
			wantSub: "duplicate source_eq",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			es := make([]mappingEntry, len(pristine))
			copy(es, pristine)
			tc.mutate(es)
			violations := checkMapping(es, universe)
			if len(violations) == 0 {
				t.Fatalf("corruption %q was accepted; want a violation containing %q", tc.name, tc.wantSub)
			}
			found := false
			for _, v := range violations {
				if strings.Contains(v, tc.wantSub) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("violations %q do not name the offending field/record %q", violations, tc.wantSub)
			}
		})
	}
}

// TestSourceMappingStatusDeclared pins the manifest's honesty limitation:
// the mapping is structural metadata, not source-verified correspondence.
func TestSourceMappingStatusDeclared(t *testing.T) {
	m := loadManifestLimits(t)
	for _, l := range m.Limitations {
		if l.ID == "source_mapping_structural_only" {
			if !strings.Contains(l.Statement, "not source-verified") {
				t.Fatalf("limitation %q does not state that mechanical consistency is not source-verified correspondence: %q", l.ID, l.Statement)
			}
			return
		}
	}
	t.Fatal("limitation source_mapping_structural_only missing from manifest.json limitations")
}

// Mapping exposes the sidecar entries for the corruption tests above.
func (d mappingDoc) Mapping() []mappingEntry {
	out := make([]mappingEntry, len(d.SourceMapping))
	copy(out, d.SourceMapping)
	return out
}
