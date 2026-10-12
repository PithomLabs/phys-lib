package gr

import (
	"errors"
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// fluxSetup builds a complete valid evidence set: divergence relation,
// integral over domain, boundary lemma equation, stationarity condition,
// bulk/caps terms, assumptions, framework.
func fluxSetup(t *testing.T, s physmath.Store) (div, integral, domain artifact.ArtifactRef, ev FluxEvidence) {
	t.Helper()
	put := func(f physmath.FamilyID, p []byte) artifact.ArtifactRef {
		r, err := s.Put(f, p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	expr := func(e physmath.Expr) artifact.ArtifactRef {
		p, err := e.Encode()
		if err != nil {
			t.Fatal(err)
		}
		return put(physmath.FamilyExpression, p)
	}
	divBody, err := physmath.EncodeRelationBody(physmath.OpEQ,
		physmath.NewEphemeralSymbol("div_j"), physmath.NewEphemeralSymbol("bulk_t"))
	if err != nil {
		t.Fatal(err)
	}
	div = put(physmath.FamilyRelation, divBody)
	tau := expr(physmath.NewEphemeralSymbol("tau"))
	dm := physmath.Domain{Kind: physmath.DomainHypersurface, DefiningRefs: []artifact.ArtifactRef{tau}}
	dmP, err := dm.Encode()
	if err != nil {
		t.Fatal(err)
	}
	domain = put(physmath.FamilyDomain, dmP)
	in := physmath.Integral{DomainRef: domain, IntegrandRefs: []artifact.ArtifactRef{expr(physmath.NewEphemeralSymbol("T"))}}
	inP, err := in.Encode()
	if err != nil {
		t.Fatal(err)
	}
	integral = put(physmath.FamilyIntegral, inP)
	lemBody, err := physmath.EncodeRelationBody(physmath.OpEQ,
		physmath.NewEphemeralSymbol("jN_walls"), physmath.MustRational(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	lemRel := put(physmath.FamilyRelation, lemBody)
	lemEq := physmath.Equation{RelationRef: lemRel}.Encode()
	lemma := put(physmath.FamilyEquation, lemEq)
	statBody, err := physmath.EncodeRelationBody(physmath.OpEQ,
		physmath.NewEphemeralSymbol("mom"), physmath.MustRational(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	statRel := put(physmath.FamilyRelation, statBody)
	stat := put(physmath.FamilyCondition, statRel.Encode())
	fw, err := NewFramework("general_relativity")
	if err != nil {
		t.Fatal(err)
	}
	fwRef, err := fw.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	asm := PhysicalAssumption{AssumptionID: "stationarity_zero_momentum",
		Category: CategoryRegularity, Description: "test stationarity"}
	if err := asm.Validate(); err != nil {
		t.Fatal(err)
	}
	asmRef, err := asm.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	ev = FluxEvidence{
		BoundaryLemmaRefs: []artifact.ArtifactRef{lemma},
		StationarityRef:   stat,
		BulkTermRef:       expr(physmath.NewEphemeralSymbol("bulk")),
		CapsTermRef:       expr(physmath.NewEphemeralSymbol("caps")),
		AssumptionRefs:    []artifact.ArtifactRef{asmRef},
		FrameworkRef:      fwRef,
	}
	return div, integral, domain, ev
}

func TestFluxConserveHappyPath(t *testing.T) {
	s := physmath.NewMemStore()
	div, integral, domain, ev := fluxSetup(t, s)
	cert, err := CheckFluxConservation(s, div, integral, domain, ev)
	if err != nil {
		t.Fatalf("valid inference rejected: %v", err)
	}
	// Conclusion independently recomputed: EQ(bulk, caps).
	bulk, err := physmath.DecodeExpr(artifact.NewDecoder(mustPayload(t, s, ev.BulkTermRef)))
	if err != nil {
		t.Fatal(err)
	}
	caps, err := physmath.DecodeExpr(artifact.NewDecoder(mustPayload(t, s, ev.CapsTermRef)))
	if err != nil {
		t.Fatal(err)
	}
	want, err := physmath.EncodeRelationBody(physmath.OpEQ, bulk, caps)
	if err != nil {
		t.Fatal(err)
	}
	gotFam, gotPayload, err := s.Lookup(cert.ConclusionRef)
	if err != nil {
		t.Fatal(err)
	}
	if gotFam != physmath.FamilyRelation {
		t.Fatal("conclusion is not relation/1")
	}
	if string(gotPayload) != string(want) {
		t.Fatal("conclusion not assembled from validated terms")
	}
	// Determinism: identical inputs re-mint the identical reference.
	cert2, err := CheckFluxConservation(s, div, integral, domain, ev)
	if err != nil {
		t.Fatal(err)
	}
	if cert.ConclusionRef != cert2.ConclusionRef || cert.CertDigest() != cert2.CertDigest() {
		t.Fatal("inference nondeterministic")
	}
	// ExpectedRef postcondition accepts the true conclusion...
	ev.ExpectedRef = &cert.ConclusionRef
	if _, err := CheckFluxConservation(s, div, integral, domain, ev); err != nil {
		t.Fatalf("true postcondition rejected: %v", err)
	}
	// ...and rejects a tampered one.
	other := ev.BulkTermRef
	ev.ExpectedRef = &other
	if _, err := CheckFluxConservation(s, div, integral, domain, ev); err == nil {
		t.Fatal("tampered expectation accepted")
	}
}

func mustPayload(t *testing.T, s physmath.Store, r artifact.ArtifactRef) []byte {
	t.Helper()
	_, p, err := s.Lookup(r)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFluxConserveFailClosed(t *testing.T) {
	cases := map[string]func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence){
		"P1 wrong family": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			return ev.BulkTermRef, integral, domain, ev
		},
		"P2 domain mismatch": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			other := physmath.Domain{Kind: physmath.DomainRegion, DefiningRefs: []artifact.ArtifactRef{ev.BulkTermRef}}
			p, _ := other.Encode()
			r, _ := s.Put(physmath.FamilyDomain, p)
			return div, integral, r, ev
		},
		"P2 integral wrong family": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			return div, domain, domain, ev
		},
		"P3 stationarity not condition": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			ev.StationarityRef = ev.BulkTermRef
			return div, integral, domain, ev
		},
		"P4 zero boundary lemmas": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			ev.BoundaryLemmaRefs = nil
			return div, integral, domain, ev
		},
		"P4 lemma wrong family": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			ev.BoundaryLemmaRefs = []artifact.ArtifactRef{ev.BulkTermRef}
			return div, integral, domain, ev
		},
		"P4 zero assumptions": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			ev.AssumptionRefs = nil
			return div, integral, domain, ev
		},
		"P4 assumption wrong family": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			ev.AssumptionRefs = []artifact.ArtifactRef{ev.BulkTermRef}
			return div, integral, domain, ev
		},
		"P5 bulk wrong family": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			ev.BulkTermRef = domain
			return div, integral, domain, ev
		},
		"P1 unresolvable": func(s physmath.Store, div, integral, domain artifact.ArtifactRef, ev FluxEvidence) (artifact.ArtifactRef, artifact.ArtifactRef, artifact.ArtifactRef, FluxEvidence) {
			ghost := ev.BulkTermRef
			ghost.Digest = [32]byte{9, 9, 9}
			return ghost, integral, domain, ev
		},
	}
	for name, mutate := range cases {
		s := physmath.NewMemStore()
		div, integral, domain, ev := fluxSetup(t, s)
		d, i, dm, e := mutate(s, div, integral, domain, ev)
		if _, err := CheckFluxConservation(s, d, i, dm, e); err == nil {
			t.Fatalf("%s accepted", name)
		} else if !errors.Is(err, ErrFluxPrecondition) {
			t.Fatalf("%s failed without sentinel: %v", name, err)
		}
	}
}

func TestFluxConserveOrderMatters(t *testing.T) {
	// Swapped bulk/caps terms mint a different conclusion: term roles are
	// load-bearing, and ExpectedRef catches the swap.
	s := physmath.NewMemStore()
	div, integral, domain, ev := fluxSetup(t, s)
	cert, err := CheckFluxConservation(s, div, integral, domain, ev)
	if err != nil {
		t.Fatal(err)
	}
	ev2 := ev
	ev2.BulkTermRef, ev2.CapsTermRef = ev.CapsTermRef, ev.BulkTermRef
	ev2.ExpectedRef = &cert.ConclusionRef
	if _, err := CheckFluxConservation(s, div, integral, domain, ev2); err == nil {
		t.Fatal("swapped terms satisfied the original postcondition")
	}
}
