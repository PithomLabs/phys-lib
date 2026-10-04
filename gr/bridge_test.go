package gr

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// D2 bridge: a rule-surfaced MathCondition is captured by reference into a
// framework-owned assumption. Rules generate nothing; the framework links.
func TestCaptureConditionBridge(t *testing.T) {
	s := testStore()
	x := physmath.NewEphemeralSymbol("x")
	relBytes, err := physmath.EncodeRelationBody(physmath.OpNEQ, x, physmath.MustRational(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	relRef, err := s.Put(physmath.FamilyRelation, relBytes)
	if err != nil {
		t.Fatal(err)
	}
	condRef, err := s.Put(physmath.FamilyCondition, relRef.Encode())
	if err != nil {
		t.Fatal(err)
	}
	a, err := CaptureCondition(s, condRef, "regularity_x_nonzero", CategoryRegularity,
		"Coordinate regularity: x != 0 on the domain.")
	if err != nil {
		t.Fatal(err)
	}
	if len(a.ConditionRefs) != 1 || a.ConditionRefs[0] != condRef {
		t.Fatal("bridge did not preserve the exact condition reference")
	}
	if a.Category != CategoryRegularity {
		t.Fatal("bridge category wrong")
	}
	if _, err := a.Mint(s); err != nil {
		t.Fatal(err)
	}
	// Non-condition references are rejected.
	exprRef, err := physmath.PutExpression(s, x)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CaptureCondition(s, exprRef, "bad", CategoryRegularity, "d"); err == nil {
		t.Fatal("non-condition captured as assumption condition")
	}
	// Missing references are rejected.
	ghost, _ := artifact.ComputeRef("physmath", "1", "condition/1", []byte("ghost"))
	if _, err := CaptureCondition(s, ghost, "bad", CategoryRegularity, "d"); err == nil {
		t.Fatal("unresolvable condition captured")
	}
}

func TestSeedConventions(t *testing.T) {
	s := testStore()
	refs, err := SeedConventions(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"metric_signature", "curvature_convention", "cosmological_constant"} {
		ref, ok := refs[id]
		if !ok {
			t.Fatalf("missing seed %q", id)
		}
		fam, _, err := s.Lookup(ref)
		if err != nil {
			t.Fatal(err)
		}
		if fam != FamilyAssumption {
			t.Fatalf("seed %q under wrong family", id)
		}
	}
}
