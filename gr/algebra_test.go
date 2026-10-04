package gr

import (
	"errors"
	"testing"

	"github.com/PithomLabs/phys-math"
)

// Category-B vector: GR-local algebra behaves on copied legacy shapes.
// Category-C vectors: projection success for the representable subset and
// explicit failure for Sin/Cos/Unknown (D9/R7 — no synthetic keys).
func TestProjectionSuccessSubset(t *testing.T) {
	r, err := NewGRational(3, 2)
	if err != nil {
		t.Fatal(err)
	}
	g := NewGMul(NewGSymbol("r"), r)
	got, err := g.ToPhysMath()
	if err != nil {
		t.Fatal(err)
	}
	want := physmath.Mul(physmath.NewEphemeralSymbol("r"), physmath.MustRational(3, 2))
	if !got.Equal(want) {
		t.Fatal("representable projection mismatch")
	}
	// Integer powers project.
	pw, err := NewGPow(NewGSymbol("r"), -1).ToPhysMath()
	if err != nil {
		t.Fatal(err)
	}
	wantPw := physmath.Pow(physmath.NewEphemeralSymbol("r"), physmath.MustRational(-1, 1))
	if !pw.Equal(wantPw) {
		t.Fatal("Pow projection mismatch")
	}
	// Neg/Add nest correctly.
	nested, err := NewGAdd(NewGNeg(NewGSymbol("a")), NewGSymbol("b")).ToPhysMath()
	if err != nil {
		t.Fatal(err)
	}
	wantNested := physmath.Add(
		physmath.Neg(physmath.NewEphemeralSymbol("a")),
		physmath.NewEphemeralSymbol("b"))
	if !nested.Equal(wantNested) {
		t.Fatal("nested projection mismatch")
	}
}

func TestProjectionFailures(t *testing.T) {
	theta := NewGSymbol("theta")
	for _, g := range []GRExpr{
		NewGSin(theta),
		NewGCos(theta),
		NewGUnknown("A", NewGSymbol("r")),
		NewGMul(NewGSymbol("r"), NewGSin(theta)),
	} {
		if _, err := g.ToPhysMath(); !errors.Is(err, ErrUnrepresentableProjection) {
			t.Fatalf("%s projected without error", g.Kind)
		}
	}
	// Legacy spherical shape r^2·sin²θ is GR-local only: never projected.
	r := NewGSymbol("r")
	spherical, err := NewGMul(
		NewGPow(r, 2),
		NewGPow(NewGSin(theta), 2),
	).ToPhysMath()
	_ = spherical
	if err == nil {
		t.Fatal("spherical sin² term projected — trig leaked toward PhysMath")
	} else if !errors.Is(err, ErrUnrepresentableProjection) {
		t.Fatalf("wrong error class: %v", err)
	}
	if _, err := NewGRational(1, 0); err == nil {
		t.Fatal("zero denominator accepted in GR-local rational")
	}
}
