package ltb

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

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
	ProvenanceNote string               `json:"provenance_note"`
	Regime         string               `json:"regime"`
	Conventions    map[string]string    `json:"conventions"`
	Target         string               `json:"target"`
	Domain         string               `json:"domain"`
	Claims         []manifestClaim      `json:"claims"`
	Obligations    []manifestObligation `json:"obligations"`
	Assumptions    []string             `json:"assumptions"`
	Verification   struct {
		Executed        string `json:"executed"`
		RepresentedOnly string `json:"represented_only"`
	} `json:"verification"`
	Limitations []struct {
		ID        string `json:"id"`
		Statement string `json:"statement"`
	} `json:"limitations"`
	Literature []struct {
		Cite        string `json:"cite"`
		Establishes string `json:"establishes"`
		Status      string `json:"status"`
	} `json:"literature"`
}

func loadManifest(t *testing.T) manifest {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(file), "manifest.json")
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
	if m.ProvenanceNote == "" {
		t.Fatal("provenance note missing: plan2_z origin must be recorded")
	}
	if m.Conventions["radial_gauge"] == "" {
		t.Fatal("radial gauge convention missing")
	}
}

func TestManifestClaimStatuses(t *testing.T) {
	m := loadManifest(t)
	want := map[string]string{
		"ZEH-Q-01": StatusLiteratureClaim,
		"ZEH-Q-02": StatusCandidate,
		"ZEH-Q-03": StatusOpenConjecture,
		"ZEH-Q-04": StatusDeferred,
	}
	if len(m.Claims) != len(want) {
		t.Fatalf("claims = %d, want %d", len(m.Claims), len(want))
	}
	for _, c := range m.Claims {
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
	m := loadManifest(t)
	ids := map[string]bool{}
	for _, o := range m.Obligations {
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
		"no_pair_constructed": false,
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
