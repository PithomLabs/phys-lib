package maxwellbox

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

func TestFixtureContractionEvidence(t *testing.T) {
	f := Build()
	if err := RequireContractionEvaluation(f.Store, f.WEvalRef); err != nil {
		t.Fatalf("fixture w certificate rejected: %v", err)
	}
	if err := RequireNegatedCurrent(f.Store, f.WRef, f.JRef, f.JDerivRefs); err != nil {
		t.Fatalf("fixture j negation rejected: %v", err)
	}
	// s2 stays UNSUPPORTED: the integral half is unexecuted by design.
	unsupported := false
	for _, id := range f.Sketch.UnsupportedSteps() {
		if id == "s2" {
			unsupported = true
		}
	}
	if !unsupported {
		t.Fatal("s2 must remain UNSUPPORTED while the energy integral is unexecuted")
	}
}

func TestConsumerRejectsCrossOperands(t *testing.T) {
	f := Build()
	fam, payload, err := f.Store.Lookup(f.WEvalRef)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyContractionEvaluation {
		t.Fatal("fixture record is not contraction_evaluation/1")
	}
	rec, err := physmath.DecodeContractionEvaluation(payload)
	if err != nil {
		t.Fatal(err)
	}
	// Honest second tensor over the same space/labels/dimensions, edited
	// diagonal content; reuse the old derivation and result.
	other, err := physmath.PutExpression(f.Store, physmath.NewEphemeralSymbol("U00"))
	if err != nil {
		t.Fatal(err)
	}
	_ = other
	ia, ib := spaceIndexRefs(t, f)
	alt := honestTensor(t, f, ia, ib, []string{"U00", "T01", "T10", "T11"})
	forged := rec
	forged.TensorRef = alt
	raw := forged.Encode()
	cref, err := f.Store.Put(physmath.FamilyContractionEvaluation, raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireContractionEvaluation(f.Store, cref); err == nil {
		t.Fatal("consumer gate accepted cross-operand evidence")
	}
}

func TestConsumerRejectsJAsW(t *testing.T) {
	f := Build()
	fam, payload, err := f.Store.Lookup(f.WEvalRef)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyContractionEvaluation {
		t.Fatal("fixture record is not contraction_evaluation/1")
	}
	rec, err := physmath.DecodeContractionEvaluation(payload)
	if err != nil {
		t.Fatal(err)
	}
	// Claim the negated j tensor as the evaluation output: the w-only
	// boundary must reject it.
	forged := rec
	forged.ResultTensorRef = f.JRef
	raw := forged.Encode()
	cref, err := f.Store.Put(physmath.FamilyContractionEvaluation, raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireContractionEvaluation(f.Store, cref); err == nil {
		t.Fatal("consumer gate accepted the negated current as evaluation output")
	}
}

func TestConsumerRejectsBrokenNegation(t *testing.T) {
	f := Build()
	// Drop one negation derivation: linkage must fail closed.
	if err := RequireNegatedCurrent(f.Store, f.WRef, f.JRef, f.JDerivRefs[:1]); err == nil {
		t.Fatal("negation gate accepted a missing component derivation")
	}
	// Swap w for an unrelated tensor: premise construction must fail.
	other, err := physmath.PutExpression(f.Store, physmath.NewEphemeralSymbol("elsewhere"))
	if err != nil {
		t.Fatal(err)
	}
	_ = other
	ia, ib := spaceIndexRefs(t, f)
	alt := honestTensor(t, f, ia, ib, []string{"Q00", "Q01", "Q10", "Q11"})
	if err := RequireNegatedCurrent(f.Store, alt, f.JRef, f.JDerivRefs); err == nil {
		t.Fatal("negation gate accepted premises built from another tensor")
	}
}

// spaceIndexRefs decodes the fixture honest tensor to reuse its
// Space-bearing index artifacts in mutation tests.
func spaceIndexRefs(t *testing.T, f *Fixture) (artifact.ArtifactRef, artifact.ArtifactRef) {
	t.Helper()
	fam, payload, err := f.Store.Lookup(f.HonestTRef)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyTensor {
		t.Fatal("honest tensor is not tensor/1")
	}
	tn, err := physmath.DecodeTensor(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(tn.IndexRefs) != 2 {
		t.Fatal("honest tensor must carry two index refs")
	}
	return tn.IndexRefs[0], tn.IndexRefs[1]
}

// honestTensor mints an alternate rank-2 tensor over the given indices with
// the named display-symbol components.
func honestTensor(t *testing.T, f *Fixture, ia, ib artifact.ArtifactRef, names []string) artifact.ArtifactRef {
	t.Helper()
	if len(names) != 4 {
		t.Fatal("need four component names")
	}
	crefs := make([]artifact.ArtifactRef, 4)
	for i, n := range names {
		r, err := physmath.PutExpression(f.Store, physmath.NewEphemeralSymbol(n))
		if err != nil {
			t.Fatal(err)
		}
		crefs[i] = r
	}
	tn := physmath.Tensor{Rank: 2, Dimension: 2,
		IndexRefs: []artifact.ArtifactRef{ia, ib}, ComponentRefs: crefs}
	if err := tn.Validate(); err != nil {
		t.Fatal(err)
	}
	ref, err := f.Store.Put(physmath.FamilyTensor, tn.Encode())
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
