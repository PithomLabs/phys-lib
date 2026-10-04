package protocol_test

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-lib/gr/equivariance"
	"github.com/PithomLabs/phys-math"
)

// TestClaimTracePasses establishes the positive side of the H-3 invariant:
// the unmodified fixture's original sides trace to the conclusion sides.
func TestClaimTracePasses(t *testing.T) {
	fx, err := equivariance.Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := equivariance.VerifyClaim(fx); err != nil {
		t.Fatalf("claim trace failed on valid fixture: %v", err)
	}
	if err := equivariance.Verify(fx); err != nil {
		t.Fatalf("Verify failed on valid fixture: %v", err)
	}
}

func mutantFixture(t *testing.T, mutate func(fx *equivariance.Fixture)) *equivariance.Fixture {
	t.Helper()
	fx, err := equivariance.Build()
	if err != nil {
		t.Fatal(err)
	}
	cp := *fx
	mutate(&cp)
	return &cp
}

func mintExpr(t *testing.T, fx *equivariance.Fixture, e physmath.Expr) artifact.ArtifactRef {
	t.Helper()
	ref, err := physmath.PutExpression(fx.Store, e)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// Plan10 H-3 adversarial test: replace LHS0 and/or RHS0 with different valid
// expressions while Step2/Step5 remain unchanged. Verification MUST fail.
// If it passes, the original proposition is semantically dead → PROTOCOL GAP.
func TestOriginalSideMutantsFail(t *testing.T) {
	p0 := physmath.DeriveParameterID(0)
	p1 := physmath.DeriveParameterID(1)
	zzz := physmath.NewEphemeralSymbol("zzz")
	cases := map[string]func(fx *equivariance.Fixture){
		// LHS0 replaced by a different valid expression.
		"mutate-LHS0-only": func(fx *equivariance.Fixture) {
			fx.LHS0 = mintExpr(t, fx, physmath.Call(fx.F, zzz))
		},
		// RHS0 replaced by a different valid expression.
		"mutate-RHS0-only": func(fx *equivariance.Fixture) {
			fx.RHS0 = mintExpr(t, fx,
				physmath.Call(fx.RhoY, zzz, physmath.NewBoundSymbol("w", p1)))
		},
		// LHS0 disconnected: same head declaration, different inner call
		// (y schematic instead of x) — linkage to Step 0 breaks.
		"disconnect-LHS0-arg": func(fx *equivariance.Fixture) {
			otherInner := physmath.Call(fx.RhoX,
				physmath.NewBoundSymbol("g", p0), physmath.NewBoundSymbol("y", p1))
			fx.LHS0 = mintExpr(t, fx, physmath.Call(fx.F, otherInner))
		},
		// Declaration replaced: LHS0 references a different Function.
		"replace-declaration": func(fx *equivariance.Fixture) {
			altBody := physmath.Mul(physmath.MustRational(3, 1),
				physmath.NewBoundSymbol("v", p0))
			altBytes, err := altBody.Encode()
			if err != nil {
				t.Fatal(err)
			}
			altBodyRef, err := fx.Store.Put(physmath.FamilyExpression, altBytes)
			if err != nil {
				t.Fatal(err)
			}
			altFn := physmath.Function{ParameterIDs: []physmath.ParameterID{p0}, BodyRef: altBodyRef}
			fnBytes, err := altFn.Encode()
			if err != nil {
				t.Fatal(err)
			}
			altRef, err := fx.Store.Put(physmath.FamilyFunction, fnBytes)
			if err != nil {
				t.Fatal(err)
			}
			fx.LHS0 = mintExpr(t, fx, physmath.Call(altRef, zzz))
		},
	}
	_ = p0
	for name, mutate := range cases {
		mut := mutantFixture(t, mutate)
		// Step2/Step5 genuinely unchanged in every mutant.
		base, err := equivariance.Build()
		if err != nil {
			t.Fatal(err)
		}
		if mut.Step[2] != base.Step[2] || mut.Step[5] != base.Step[5] {
			t.Fatalf("%s: steps changed (mutant invalid)", name)
		}
		if err := equivariance.Verify(mut); err == nil {
			t.Fatalf("mutant %q verified: original sides are semantically dead", name)
		}
		if err := equivariance.VerifyClaim(mut); err == nil {
			t.Fatalf("mutant %q passed claim trace", name)
		}
	}
}

func mintResolve(t *testing.T, fx *equivariance.Fixture, ref artifact.ArtifactRef) physmath.Expr {
	t.Helper()
	_, payload, err := fx.Store.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
	if err != nil {
		t.Fatal(err)
	}
	return e
}
