package protocol_test

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-lib/gr/equivariance"
	"github.com/PithomLabs/phys-math"
)

// TestFreshVerifierPasses proves the persisted artifacts suffice: a verifier
// that never touches Fixture semantic fields reconstructs the proposition
// and its trace to the conclusion.
func TestFreshVerifierPasses(t *testing.T) {
	store, drvRef, thRef := buildPersisted(t)
	if err := freshVerify(t, store, drvRef, thRef); err != nil {
		t.Fatalf("fresh verifier failed on valid fixture: %v", err)
	}
}

// mutantDerivation decodes the persisted derivation, applies a persisted-
// representation mutation, and re-encodes it. The mutant derivation is new
// persisted bytes; finals stay fixed where the mutation permits.
func mutantDerivation(t *testing.T, store physmath.Store, drvRef artifact.ArtifactRef,
	mutate func(*physmath.Derivation)) artifact.ArtifactRef {
	t.Helper()
	_, payload, err := store.Lookup(drvRef)
	if err != nil {
		t.Fatal(err)
	}
	drv, err := equivariance.DecodeDerivationPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	mutate(&drv)
	if err := drv.Validate(); err != nil {
		t.Fatalf("mutant structurally invalid (mutation too coarse): %v", err)
	}
	enc, err := drv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	mutRef, err := store.Put(physmath.FamilyDerivation, enc)
	if err != nil {
		t.Fatal(err)
	}
	return mutRef
}

// TestFreshVerifierMutations: every persisted-representation source mutation
// must fail the fresh verifier while designated finals stay fixed.
func TestFreshVerifierMutations(t *testing.T) {
	mintCall := func(t *testing.T, store physmath.Store, fn artifact.ArtifactRef, args ...physmath.Expr) artifact.ArtifactRef {
		t.Helper()
		ref, err := physmath.PutExpression(store, physmath.Call(fn, args...))
		if err != nil {
			t.Fatal(err)
		}
		return ref
	}
	newFunction := func(t *testing.T, store physmath.Store, body physmath.Expr) artifact.ArtifactRef {
		t.Helper()
		p0 := physmath.DeriveParameterID(0)
		bodyBytes, err := body.Encode()
		if err != nil {
			t.Fatal(err)
		}
		bodyRef, err := store.Put(physmath.FamilyExpression, bodyBytes)
		if err != nil {
			t.Fatal(err)
		}
		fn := physmath.Function{ParameterIDs: []physmath.ParameterID{p0}, BodyRef: bodyRef}
		fnBytes, err := fn.Encode()
		if err != nil {
			t.Fatal(err)
		}
		ref, err := store.Put(physmath.FamilyFunction, fnBytes)
		if err != nil {
			t.Fatal(err)
		}
		return ref
	}
	cases := map[string]func(t *testing.T, store physmath.Store, d *physmath.Derivation){
		// Corrupt the Step-0 input call (the persisted original-side material).
		"mutated-step0-input": func(t *testing.T, store physmath.Store, d *physmath.Derivation) {
			fDecl := d.Steps[1].InputRefs[1]
			zzz := physmath.NewEphemeralSymbol("zzz")
			d.Steps[0].InputRefs[0] = mintCall(t, store, fDecl, zzz)
		},
		// Corrupt the Step-3 input call (right original side).
		"mutated-step3-input": func(t *testing.T, store physmath.Store, d *physmath.Derivation) {
			fDecl := d.Steps[3].InputRefs[1]
			zzz := physmath.NewEphemeralSymbol("zzz")
			d.Steps[3].InputRefs[0] = mintCall(t, store, fDecl, zzz)
		},
		// Break staged LHS linkage: Step-1 input is no longer f(Step-0 output).
		"broken-lhs-staged-linkage": func(t *testing.T, store physmath.Store, d *physmath.Derivation) {
			fDecl := d.Steps[1].InputRefs[1]
			zzz := physmath.NewEphemeralSymbol("zzz")
			d.Steps[1].InputRefs[0] = mintCall(t, store, fDecl, zzz)
		},
		// Break staged RHS linkage similarly.
		"broken-rhs-staged-linkage": func(t *testing.T, store physmath.Store, d *physmath.Derivation) {
			rhoDecl := d.Steps[4].InputRefs[1]
			g := resolveExpr(t, store, d.Steps[0].InputRefs[0])
			_ = g
			zzz := physmath.NewEphemeralSymbol("zzz")
			inner := resolveExpr(t, store, d.Steps[3].InputRefs[0])
			_ = inner
			d.Steps[4].InputRefs[0] = mintCall(t, store, rhoDecl, zzz, zzz)
		},
		// Changed declaration: Step-1 references a different Function.
		"changed-declaration": func(t *testing.T, store physmath.Store, d *physmath.Derivation) {
			alt := newFunction(t, store, physmath.Mul(physmath.MustRational(9, 1),
				physmath.NewBoundSymbol("v", physmath.DeriveParameterID(0))))
			d.Steps[1].InputRefs[1] = alt
		},
	}
	for name, mutate := range cases {
		store, drvRef, thRef := buildPersisted(t)
		mutRef := mutantDerivation(t, store, drvRef,
			func(d *physmath.Derivation) { mutate(t, store, d) })
		if mutRef == drvRef {
			t.Fatalf("%s: mutation did not change persisted bytes", name)
		}
		if err := freshVerify(t, store, mutRef, thRef); err == nil {
			t.Fatalf("mutant %q passed fresh verification: gap unproven", name)
		}
	}
}
