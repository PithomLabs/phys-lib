package gr

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

func testStoreFW(t *testing.T) (*physmath.MemStore, artifact.ArtifactRef) {
	t.Helper()
	s := physmath.NewMemStore()
	fw, err := NewFramework("general_relativity")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := fw.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	return s, ref
}

func TestFrameDefinition(t *testing.T) {
	s, fwRef := testStoreFW(t)
	e, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol("xbar"))
	if err != nil {
		t.Fatal(err)
	}
	d, err := NewFrameDefinition(fwRef, "rindler_frame", "accelerated_box_frame",
		[]artifact.ArtifactRef{e}, nil, nil)
	if err != nil {
		t.Fatalf("valid frame rejected: %v", err)
	}
	if _, err := d.Mint(s); err != nil {
		t.Fatalf("frame mint failed: %v", err)
	}
	if _, err := NewFrameDefinition(fwRef, "", "t", nil, nil, nil); err == nil {
		t.Fatal("empty definition ID accepted")
	}
	if _, err := NewChartAssumption("rindler_chart", "Rindler (6)-(7) adopted.", nil); err != nil {
		t.Fatalf("valid chart assumption rejected: %v", err)
	}
	if _, err := NewChartAssumption("BAD ID", "x", nil); err == nil {
		t.Fatal("bad chart ID accepted")
	}
}

func TestMaxwellBoxAssumptions(t *testing.T) {
	s, _ := testStoreFW(t)
	got, err := SeedMaxwellBoxAssumptions(s)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	for _, k := range []string{"stationarity_zero_momentum", "perfectly_conducting_walls", "rindler_chart"} {
		if _, ok := got[k]; !ok {
			t.Fatalf("missing assumption %s", k)
		}
	}
}

func TestProtocolSketchValidation(t *testing.T) {
	s, _ := testStoreFW(t)
	e, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol("x"))
	if err != nil {
		t.Fatal(err)
	}
	good := ProtocolSketch{SketchID: "k1", UnitID: "d1", Steps: []SketchStep{
		{StepID: "s1", Description: "setup", Status: SketchUnsupported, Refs: []artifact.ArtifactRef{e}},
		{StepID: "s2", Description: "endpoint", Status: SketchSupported, Refs: []artifact.ArtifactRef{e}},
	}}
	if err := good.Validate(); err != nil {
		t.Fatalf("valid sketch rejected: %v", err)
	}
	if got := good.UnsupportedSteps(); len(got) != 1 || got[0] != "s1" {
		t.Fatalf("unsupported steps wrong: %v", got)
	}
	empty := ProtocolSketch{SketchID: "k1", UnitID: "d1"}
	if err := empty.Validate(); err == nil {
		t.Fatal("empty sketch accepted")
	}
	dup := ProtocolSketch{SketchID: "k1", UnitID: "d1", Steps: []SketchStep{
		{StepID: "s1", Description: "a", Status: SketchSupported},
		{StepID: "s1", Description: "b", Status: SketchSupported},
	}}
	if err := dup.Validate(); err == nil {
		t.Fatal("duplicate step accepted")
	}
	bad := ProtocolSketch{SketchID: "k1", UnitID: "d1", Steps: []SketchStep{
		{StepID: "s1", Description: "a", Status: "EXECUTED"},
	}}
	if err := bad.Validate(); err == nil {
		t.Fatal("open status vocabulary accepted")
	}
	nodesc := ProtocolSketch{SketchID: "k1", UnitID: "d1", Steps: []SketchStep{
		{StepID: "s1", Description: "", Status: SketchSupported},
	}}
	if err := nodesc.Validate(); err == nil {
		t.Fatal("undescribed step accepted")
	}
}

// R3 volume-assumption wiring: the Rindler |xbar| -> xbar reduction is
// licensed only by a cited xbar > 0 condition captured into a framework
// assumption. The math layer gates on condition presence; the gr layer
// records whose convention it is.
func TestRindlerDomainAssumptionWiring(t *testing.T) {
	s, _ := testStoreFW(t)
	xbar := physmath.NewEphemeralSymbol("xbar")
	body, err := physmath.EncodeRelationBody(physmath.OpGTE, xbar, physmath.MustRational(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	relRef, err := s.Put(physmath.FamilyRelation, body)
	if err != nil {
		t.Fatal(err)
	}
	condRef, err := s.Put(physmath.FamilyCondition, relRef.Encode())
	if err != nil {
		t.Fatal(err)
	}
	asm, err := CaptureCondition(s, condRef, "rindler_positive_radius",
		CategoryModel, "Rindler chart domain xbar > 0 licensing |xbar| -> xbar.")
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if len(asm.ConditionRefs) != 1 || asm.ConditionRefs[0] != condRef {
		t.Fatal("captured assumption does not cite the domain condition")
	}
	asmRef, err := asm.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	// The captured assumption carries the convention; the math-layer
	// VolumeDensity gates the reduction on the condition ref itself
	// (tested in phys-math density tests). Record the linkage here.
	factor, err := physmath.PutExpression(s, xbar)
	if err != nil {
		t.Fatal(err)
	}
	_ = factor
	_ = asmRef
}
