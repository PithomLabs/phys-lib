// Package protocol_test is the executable derivation/protocol test for the
// equivariance reference fixture. It exercises the formal content through
// Build + Replay + conclusion gate; it never merely checks static files.
// The manifest is metadata and is never a derivation premise: this file
// reads manifest.json, while fixture.go does not.
package protocol_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-lib/gr/equivariance"
	"github.com/PithomLabs/phys-math"
)

// manifestJSON is read from the fixture package directory at test time via
// runtime.Caller sibling resolution (mirroring the GR precedent's packageFile
// approach: hermetic to CWD, never a derivation premise — fixture.go does not
// reference this file or the manifest).
func loadManifestJSON(t *testing.T) []byte {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	path := filepath.Join(filepath.Dir(thisFile), "..", "manifest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

type manifestPremise struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
}

type manifestNote struct {
	ID          string `json:"id"`
	Description string `json:"description"`
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
	Regime      string   `json:"regime"`
	Coordinates []string `json:"coordinates"`
	Conventions struct {
		MetricSignature      string `json:"metric_signature"`
		CurvatureConvention  string `json:"curvature_convention"`
		CosmologicalConstant string `json:"cosmological_constant"`
	} `json:"conventions"`
	Premises       []manifestPremise `json:"premises"`
	Assumptions    []string          `json:"assumptions"`
	ExpectedResult struct {
		LHS         string `json:"LHS"`
		RHS         string `json:"RHS"`
		Description string `json:"description"`
	} `json:"expected_result"`
	Verification struct {
		PrimaryDerivationTrace string `json:"primary_derivation_trace"`
		PrimaryStages          int    `json:"primary_stages"`
		SecondaryVerification  string `json:"secondary_verification"`
	} `json:"verification"`
	Limitations             []manifestNote `json:"limitations"`
	Anomalies               []manifestNote `json:"anomalies"`
	FalsificationConditions []manifestNote `json:"falsification_conditions"`
}

func decodeManifestStrict(t *testing.T) manifest {
	t.Helper()
	var m manifest
	dec := json.NewDecoder(bytes.NewReader(loadManifestJSON(t)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		t.Fatalf("strict manifest decode: %v", err)
	}
	return m
}

func TestManifestExactValues(t *testing.T) {
	m := decodeManifestStrict(t)
	if m.SchemaVersion != "1" || m.ArtifactType != "derivation" {
		t.Fatal("manifest identity keys wrong")
	}
	if m.DerivationID != "equivariance" || m.FrameworkID != "general_relativity" {
		t.Fatal("manifest derivation/framework ids wrong")
	}
	if m.EntryPoints.Primary != "equivariance.Build" || m.EntryPoints.Secondary != "equivariance.Verify" {
		t.Fatal("manifest entry points wrong")
	}
	if m.Verification.PrimaryStages != 6 {
		t.Fatal("manifest stages != 6")
	}
	if len(m.Premises) != 3 || len(m.Assumptions) != 1 {
		t.Fatal("manifest premises/assumptions wrong")
	}
	if m.Conventions.MetricSignature != "-+++" || m.Conventions.CurvatureConvention != "MTW" ||
		m.Conventions.CosmologicalConstant != "0" {
		t.Fatal("manifest conventions wrong")
	}
	if m.ExpectedResult.Description == "" {
		t.Fatal("manifest prose description missing")
	}
	// Expected hex matches live derivation finals (metadata documents execution).
	fx, err := equivariance.Build()
	if err != nil {
		t.Fatal(err)
	}
	liveL := hex.EncodeToString(fx.Step[2].Digest[:])
	liveR := hex.EncodeToString(fx.Step[5].Digest[:])
	if m.ExpectedResult.LHS != liveL || m.ExpectedResult.RHS != liveR {
		t.Fatal("manifest expected_result does not match live derivation")
	}
	if m.ExpectedResult.LHS != m.ExpectedResult.RHS {
		t.Fatal("manifest sides differ")
	}
}

func TestManifestProseIsNotMath(t *testing.T) {
	m := decodeManifestStrict(t)
	if m.ExpectedResult.Description == "" || m.Name == "" {
		t.Fatal("prose fields missing")
	}
	// Prose must not parse as canonical expression bytes.
	if _, err := physmath.DecodeExpr(artifact.NewDecoder([]byte(m.ExpectedResult.Description))); err == nil {
		t.Fatal("prose parsed as math")
	}
}

func orderedRules(t *testing.T, fx *equivariance.Fixture) []string {
	t.Helper()
	_, payload, err := fx.Store.Lookup(fx.DerivationRef)
	if err != nil {
		t.Fatal(err)
	}
	drv, err := equivariance.DecodeDerivationPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	rules := make([]string, len(drv.Steps))
	for i, s := range drv.Steps {
		if s.Ordinal != uint32(i) {
			t.Fatalf("step ordinal %d out of order", s.Ordinal)
		}
		rules[i] = s.RuleID
	}
	return rules
}

func TestEquivarianceTrace(t *testing.T) {
	fx, err := equivariance.Build()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		physmath.RuleExprApply, physmath.RuleExprApply, physmath.RuleExprCanon,
		physmath.RuleExprApply, physmath.RuleExprApply, physmath.RuleExprCanon,
	}
	got := orderedRules(t, fx)
	if len(got) != len(want) {
		t.Fatalf("want 6 steps, got %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d rule = %s, want %s", i, got[i], want[i])
		}
	}
	// Staged-call linkage: LHS_1 arg == Step 0 bytes; RHS_1 arg2 == Step 3 bytes.
	_, lhs1Payload, _ := fx.Store.Lookup(fx.LHS1)
	lhs1, err := physmath.DecodeExpr(artifact.NewDecoder(lhs1Payload))
	if err != nil {
		t.Fatal(err)
	}
	_, s0Payload, _ := fx.Store.Lookup(fx.Step[0])
	s0, err := physmath.DecodeExpr(artifact.NewDecoder(s0Payload))
	if err != nil {
		t.Fatal(err)
	}
	if len(lhs1.Args) != 1 || !lhs1.Args[0].Equal(s0) {
		t.Fatal("LHS_1 linkage to Step 0 broken")
	}
	_, rhs1Payload, _ := fx.Store.Lookup(fx.RHS1)
	rhs1, _ := physmath.DecodeExpr(artifact.NewDecoder(rhs1Payload))
	_, s3Payload, _ := fx.Store.Lookup(fx.Step[3])
	s3, _ := physmath.DecodeExpr(artifact.NewDecoder(s3Payload))
	if len(rhs1.Args) != 2 || !rhs1.Args[1].Equal(s3) {
		t.Fatal("RHS_1 linkage to Step 3 broken")
	}
	// Display firewall: only g/x/v appear; no foreign content markers.
	var checkDisplays func(e physmath.Expr)
	checkDisplays = func(e physmath.Expr) {
		if e.Kind == physmath.TagSymbol {
			switch e.DisplayName {
			case "g", "x", "v":
			default:
				t.Fatalf("foreign display %q in fixture", e.DisplayName)
			}
		}
		for _, c := range e.Children {
			checkDisplays(c)
		}
		for _, a := range e.Args {
			checkDisplays(a)
		}
		if e.Operand != nil {
			checkDisplays(*e.Operand)
		}
		if e.Base != nil {
			checkDisplays(*e.Base)
		}
		if e.Exp != nil {
			checkDisplays(*e.Exp)
		}
		if e.Left != nil {
			checkDisplays(*e.Left)
		}
		if e.Right != nil {
			checkDisplays(*e.Right)
		}
	}
	checkDisplays(s0)
	checkDisplays(s3)
	finalPayload := s0Payload
	_ = finalPayload
	// Secondary verification entry point passes on the fixture.
	if err := equivariance.Verify(fx); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	// Replay determinism: identical refs across independent builds.
	fx2, err := equivariance.Build()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		if fx.Step[i] != fx2.Step[i] {
			t.Fatalf("step %d not deterministic across builds", i)
		}
	}
	if fx.TheoremRef != fx2.TheoremRef || fx.RelationRef != fx2.RelationRef {
		t.Fatal("theorem/relation refs not deterministic")
	}
}

func rebuildWith(t *testing.T, fx *equivariance.Fixture, mutate func(*physmath.Derivation)) physmath.Derivation {
	t.Helper()
	_, payload, err := fx.Store.Lookup(fx.DerivationRef)
	if err != nil {
		t.Fatal(err)
	}
	drv, err := equivariance.DecodeDerivationPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	mutate(&drv)
	return drv
}

func TestMutationNegatives(t *testing.T) {
	fx, err := equivariance.Build()
	if err != nil {
		t.Fatal(err)
	}
	fixtures := map[artifact.ArtifactRef]bool{fx.LHS1: true, fx.RHS1: true}
	// Replay is content-addressed: order-only mutants still verify. Each
	// mutant below breaks bytes or applicability, and each must fail.
	cases := map[string]func(*physmath.Derivation){
		// Wrong input: Step 2 canonically reduces g·x, not the declared 2·g·x.
		"wrong-input": func(d *physmath.Derivation) {
			d.Steps[2].InputRefs = []artifact.ArtifactRef{d.Steps[0].OutputRef}
		},
		// Wrong function: LHS_1's Call references f, not rho.
		"wrong-function": func(d *physmath.Derivation) {
			d.Steps[1].InputRefs[1] = d.Steps[0].InputRefs[1]
		},
		// Wrong output: tampered Step-5 ref.
		"tampered-output": func(d *physmath.Derivation) {
			d.Steps[5].OutputRef = d.Steps[0].OutputRef
		},
	}
	for name, mutate := range cases {
		mut := rebuildWith(t, fx, mutate)
		if err := physmath.Replay(fx.Store, mut, fixtures); err == nil {
			t.Fatalf("mutant %q replayed without error", name)
		}
	}
	// Foreign conclusion bytes fail the gate.
	other, err := fx.Store.Put(physmath.FamilyExpression,
		mustEncodeExpr(t, physmath.NewEphemeralSymbol("zzz")))
	if err != nil {
		t.Fatal(err)
	}
	_, thPayload, _ := fx.Store.Lookup(fx.TheoremRef)
	th, err := equivariance.DecodeTheoremPayload(thPayload)
	if err != nil {
		t.Fatal(err)
	}
	if err := physmath.CheckTheoremConclusion(fx.Store, th, fx.Step[2], other); err == nil {
		t.Fatal("foreign conclusion bytes accepted")
	}
}

func mustEncodeExpr(t *testing.T, e physmath.Expr) []byte {
	t.Helper()
	b, err := e.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return b
}
