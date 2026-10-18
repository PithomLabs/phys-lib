package ltb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/PithomLabs/phys-math"
)

// radialGaugeCanonical is the unambiguous mass-profile reading. The
// "exactly, not r^(3/2)" clause is load-bearing: r^3/2 means (1/2)·r^3
// (exponent 3, coefficient 1/2); r^(3/2) is a different, wrong function.
// TestManifestRadialGauge pins this string to the value the fixture builds.
const radialGaugeCanonical = "M(r)=r^3/2 exactly, not r^(3/2)"

type manifestClaim struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	EvidenceRefs []string `json:"evidence_refs"`
}

type manifestObligation struct {
	ID               string   `json:"id"`
	Statement        string   `json:"statement"`
	Blocks           []string `json:"blocks"`
	EvidenceRequired string   `json:"evidence_required"`
}

type manifest struct {
	SchemaVersion string `json:"schema_version"`
	ArtifactType  string `json:"artifact_type"`
	DerivationID  string `json:"derivation_id"`
	FrameworkID   string `json:"framework_id"`
	Name          string `json:"name"`
	EntryPoints   struct {
		Primary   string `json:"primary"`
		Secondary string `json:"secondary"`
	} `json:"entry_points"`
	Regime       string            `json:"regime"`
	Coordinates  []string          `json:"coordinates"`
	Conventions  map[string]string `json:"conventions"`
	Assumptions  []string          `json:"assumptions"`
	Anomalies    []string          `json:"anomalies"`
	Verification struct {
		Executed        string `json:"executed"`
		RepresentedOnly string `json:"represented_only"`
	} `json:"verification"`
	Limitations []struct {
		ID        string `json:"id"`
		Statement string `json:"statement"`
	} `json:"limitations"`
}

// researchMetadata is the relocated sidecar holding the research-workflow
// fields the GR manifest schema has no room for (see PROTOCOL_GAPS.md).
type researchMetadata struct {
	ProvenanceNote string               `json:"provenance_note"`
	Target         string               `json:"target"`
	Domain         string               `json:"domain"`
	Claims         []manifestClaim      `json:"claims"`
	Obligations    []manifestObligation `json:"obligations"`
	Literature     []struct {
		Cite        string `json:"cite"`
		Establishes string `json:"establishes"`
		Status      string `json:"status"`
	} `json:"literature"`
}

func callerDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unavailable")
	}
	return filepath.Dir(file)
}

func loadManifest(t *testing.T) manifest {
	t.Helper()
	path := filepath.Join(callerDir(t), "manifest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		t.Fatalf("manifest strict parse: %v", err)
	}
	return m
}

// loadResearchMetadata reads the relocated research_metadata.json sidecar.
func loadResearchMetadata(t *testing.T) researchMetadata {
	t.Helper()
	path := filepath.Join(callerDir(t), "research_metadata.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rm researchMetadata
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rm); err != nil {
		t.Fatalf("research_metadata strict parse: %v", err)
	}
	return rm
}

// checkRadialGaugeConvention rejects any reading other than the pinned
// unambiguous string. The offending value is named in the error.
func checkRadialGaugeConvention(v string) error {
	if v != radialGaugeCanonical {
		return fmt.Errorf("conventions.radial_gauge = %q, want pinned canonical %q ((1/2)·r^3, exponent 3; NOT r^(3/2))", v, radialGaugeCanonical)
	}
	return nil
}

// checkRadialGaugeArtifact structurally pins the stored M body to (1/2)·r^3
// and rejects the r^(3/2) reading (a Pow whose exponent is 3/2). This binds
// the manifest prose to the artifact the fixture actually builds.
func checkRadialGaugeArtifact(body physmath.Expr) error {
	if body.Kind != physmath.TagMul {
		return fmt.Errorf("M body kind = %d, want TagMul", body.Kind)
	}
	half, rpow := false, false
	for _, c := range body.Children {
		want := physmath.MustRational(1, 2)
		if c.Equal(want) {
			half = true
			continue
		}
		if c.Kind == physmath.TagPow && c.Exp != nil {
			// Reject the wrong r^(3/2) reading outright.
			if c.Exp.Equal(physmath.MustRational(3, 2)) {
				return fmt.Errorf("M body carries exponent 3/2 (the r^(3/2) reading); want exponent 3 with coefficient 1/2")
			}
			if c.Exp.Equal(physmath.MustRational(3, 1)) {
				rpow = true
			}
		}
	}
	if !half || !rpow {
		return fmt.Errorf("M body is not (1/2)·r^3: half=%v r^3=%v (children=%d)", half, rpow, len(body.Children))
	}
	return nil
}

func TestManifestContract(t *testing.T) {
	m := loadManifest(t)
	if m.SchemaVersion != "1" || m.ArtifactType != "derivation" {
		t.Fatalf("manifest header = %+v", m)
	}
	if m.DerivationID != "ltb_zeh_reference" {
		t.Fatalf("derivation_id = %q", m.DerivationID)
	}
	if m.EntryPoints.Primary != "ltb.Build" || m.EntryPoints.Secondary != "ltb.Verify" {
		t.Fatalf("entry points = %+v", m.EntryPoints)
	}
	// The GR manifest schema has coordinates and anomalies; the fixture's
	// executed fragment is the radial mass profile over the single bound
	// coordinate r, and no anomalies were observed in that scope.
	if len(m.Coordinates) != 1 || m.Coordinates[0] != "r" {
		t.Fatalf("coordinates = %v, want [r]", m.Coordinates)
	}
	if m.Anomalies == nil || len(m.Anomalies) != 0 {
		t.Fatalf("anomalies = %v, want empty (none observed in executed scope)", m.Anomalies)
	}
}

// TestManifestRadialGauge pins the convention string to its canonical value
// and structurally ties it to the M body the fixture builds, rejecting the
// r^(3/2) interpretation.
func TestManifestRadialGauge(t *testing.T) {
	m := loadManifest(t)
	if err := checkRadialGaugeConvention(m.Conventions["radial_gauge"]); err != nil {
		t.Fatal(err)
	}

	f := Build()
	body := mustResolve(f.Store, f.MRef)
	if err := checkRadialGaugeArtifact(body); err != nil {
		t.Fatalf("stored M(r) artifact disagrees with the pinned radial_gauge convention: %v", err)
	}
}

// TestManifestRadialGaugeRejectsRationalExponent feeds the ambiguous/wrong
// r^(3/2) readings (and drift) to the convention checker and requires each
// to be rejected by name.
func TestManifestRadialGaugeRejectsRationalExponent(t *testing.T) {
	bad := []string{
		"M(r)=r^(3/2) exactly",
		"M(r)=r^{3/2}",
		"M(r)=r^3/2",                   // drops the disambiguating clause
		"M(r)=r^3/2, r^(3/2) accepted", // endorses the wrong reading
		"",
	}
	for _, v := range bad {
		if err := checkRadialGaugeConvention(v); err == nil {
			t.Fatalf("convention %q was accepted; want rejection naming conventions.radial_gauge", v)
		} else if !strings.Contains(err.Error(), "conventions.radial_gauge") {
			t.Fatalf("rejection of %q does not name conventions.radial_gauge: %v", v, err)
		}
	}
}

// TestManifestRadialGaugeArtifactRejectsRationalExponent pins the artifact
// side: a stored body carrying exponent 3/2 (the r^(3/2) reading) is rejected.
func TestManifestRadialGaugeArtifactRejectsRationalExponent(t *testing.T) {
	r := physmath.NewBoundSymbol("r", physmath.DeriveParameterID(0))
	wrong := physmath.Mul(physmath.MustRational(1, 1), physmath.Pow(r, physmath.MustRational(3, 2)))
	if err := checkRadialGaugeArtifact(wrong); err == nil {
		t.Fatal("artifact with exponent 3/2 was accepted; want rejection")
	} else if !strings.Contains(err.Error(), "3/2") {
		t.Fatalf("rejection does not mention the 3/2 exponent: %v", err)
	}
}

func TestManifestResearchMetadataContract(t *testing.T) {
	rm := loadResearchMetadata(t)
	if rm.ProvenanceNote == "" {
		t.Fatal("provenance note missing: plan2_z origin must be recorded")
	}
	if rm.Target == "" || rm.Domain == "" {
		t.Fatalf("target/domain missing: target=%q domain=%q", rm.Target, rm.Domain)
	}
}

func TestManifestClaimStatuses(t *testing.T) {
	rm := loadResearchMetadata(t)
	want := map[string]string{
		"ZEH-Q-01": StatusLiteratureClaim,
		"ZEH-Q-02": StatusCandidate,
		"ZEH-Q-03": StatusOpenConjecture,
		"ZEH-Q-04": StatusDeferred,
	}
	if len(rm.Claims) != len(want) {
		t.Fatalf("claims = %d, want %d", len(rm.Claims), len(want))
	}
	for _, c := range rm.Claims {
		w, ok := want[c.ID]
		if !ok {
			t.Fatalf("unknown claim %s", c.ID)
		}
		if c.Status != w {
			t.Fatalf("claim %s status = %q, want %q", c.ID, c.Status, w)
		}
		if len(c.EvidenceRefs) != 0 {
			t.Fatalf("claim %s cites evidence it does not have", c.ID)
		}
	}
}

func TestManifestObligations(t *testing.T) {
	rm := loadResearchMetadata(t)
	ids := map[string]bool{}
	for _, o := range rm.Obligations {
		if o.ID == "" || o.Statement == "" || o.EvidenceRequired == "" || len(o.Blocks) == 0 {
			t.Fatalf("obligation under-specified: %+v", o)
		}
		ids[o.ID] = true
	}
	for _, want := range []string{"O1", "O2", "O3", "O4", "O5", "O6", "O7"} {
		if !ids[want] {
			t.Fatalf("obligation %s missing", want)
		}
	}
}

func TestManifestLimitationsHonest(t *testing.T) {
	m := loadManifest(t)
	need := map[string]bool{
		"averaging_unexecuted": false, "status_is_metadata": false,
		"no_pair_constructed":         false,
		"research_metadata_relocated": false,
	}
	for _, l := range m.Limitations {
		if _, ok := need[l.ID]; ok {
			need[l.ID] = true
		}
	}
	for id, found := range need {
		if !found {
			t.Fatalf("limitation %s missing from manifest", id)
		}
	}
}
