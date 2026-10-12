package gr

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// Rindler (6): ct = xbar*sinh(tau), x = xbar*cosh(tau), represented with
// opaque hyperbolic nodes. Projection of either component must fail
// closed; the map itself validates shape only.
func TestHyperbolicOpaqueProjection(t *testing.T) {
	tau := NewGSymbol("tau")
	xbar := NewGSymbol("xbar")
	ct := NewGMul(xbar, NewGSinh(tau))
	x := NewGMul(xbar, NewGCosh(tau))
	for name, g := range map[string]GRExpr{"ct": ct, "x": x} {
		if _, err := g.ToPhysMath(); err == nil {
			t.Fatalf("%s with hyperbolic node projected", name)
		}
	}
	// Standalone nodes also fail, with identity preserved in the error path.
	if _, err := NewGSinh(tau).ToPhysMath(); err == nil {
		t.Fatal("sinh projected")
	}
	if _, err := NewGCosh(tau).ToPhysMath(); err == nil {
		t.Fatal("cosh projected")
	}
	// No lossy rewriting: the kind survives.
	if NewGSinh(tau).Kind != GSinh || NewGCosh(tau).Kind != GCosh {
		t.Fatal("hyperbolic identity lost")
	}
}

func TestTransformMapValidation(t *testing.T) {
	s := physmath.NewMemStore()
	mkMetric := func() artifact.ArtifactRef {
		e, _ := physmath.PutExpression(s, physmath.MustRational(1, 1))
		m := physmath.SymmetricBilinear{Dimension: 1,
			ComponentRefs: []artifact.ArtifactRef{e}, InverseRef: nil}
		p := m.Encode()
		r, err := s.Put(physmath.FamilySymmetricBilinear, p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	src, tgt := mkMetric(), mkMetric()
	tau := NewGSymbol("tau")
	m := TransformMap{
		MapID: "rindler_chart", SourceChart: "inertial_minkowski",
		TargetChart:    "rindler_accelerated",
		Components:     []GRExpr{NewGMul(NewGSymbol("xbar"), NewGSinh(tau))},
		SourceMetric:   src,
		TargetMetric:   tgt,
		AssumptionRefs: nil,
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("valid map rejected: %v", err)
	}
	if got := m.OpaqueComponents(); len(got) != 1 || got[0] != 0 {
		t.Fatalf("opaque detection wrong: %v", got)
	}
	// Pure-algebra map has no opaque components (still not executable —
	// the type has no execution method by construction).
	alg := m
	alg.Components = []GRExpr{NewGMul(NewGSymbol("xbar"), tau)}
	if got := alg.OpaqueComponents(); len(got) != 0 {
		t.Fatalf("algebra map flagged opaque: %v", got)
	}
	empty := m
	empty.Components = nil
	if err := empty.Validate(); err == nil {
		t.Fatal("component-free map accepted")
	}
	nochart := m
	nochart.TargetChart = ""
	if err := nochart.Validate(); err == nil {
		t.Fatal("chartless map accepted")
	}
	badns := m
	badns.SourceMetric = artifact.ArtifactRef{Namespace: "BAD", SchemaVersion: "1", ArtifactType: "x/1"}
	if err := badns.Validate(); err == nil {
		t.Fatal("bad metric namespace accepted")
	}
	dup := m
	dup.AssumptionRefs = []artifact.ArtifactRef{src, src}
	if err := dup.Validate(); err == nil {
		t.Fatal("duplicate assumption refs accepted")
	}
}

// Anti-vacuity: a validated map proves nothing about the transformation.
// There is deliberately no Execute/Verify/Replay method on TransformMap;
// this test pins the schematic status alongside validation.
func TestTransformMapProvesNothing(t *testing.T) {
	s := physmath.NewMemStore()
	e, _ := physmath.PutExpression(s, physmath.MustRational(1, 1))
	mk := func() artifact.ArtifactRef {
		m := physmath.SymmetricBilinear{Dimension: 1,
			ComponentRefs: []artifact.ArtifactRef{e}, InverseRef: nil}
		r, err := s.Put(physmath.FamilySymmetricBilinear, m.Encode())
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	m := TransformMap{
		MapID: "rindler_chart", SourceChart: "a", TargetChart: "b",
		Components:   []GRExpr{NewGSymbol("x")},
		SourceMetric: mk(), TargetMetric: mk(),
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(m.OpaqueComponents()) != 0 {
		t.Fatal("pure-symbol map flagged opaque")
	}
	// Validation success coexists with zero execution semantics: the
	// type exposes no execution entry point (compile-time property,
	// asserted here as documentation of the anti-vacuity contract).
}
