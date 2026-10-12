package minisuperspace

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// TestFRWCovariantFragment exercises the complete metric → general
// inverse → Levi-Civita → covariant-derivative path on the flat FRW
// spatial metric h_ij = a²δ_ij in a Cartesian chart. The scale factor a
// carries a parameter identity disjoint from the coordinates (spatially
// constant by the FRW symmetries), so all spatial partials vanish and the
// connection is honestly zero — asserted, not assumed.
func TestFRWCovariantFragment(t *testing.T) {
	s := physmath.NewMemStore()
	px := physmath.DeriveParameterID(0)
	py := physmath.DeriveParameterID(1)
	pz := physmath.DeriveParameterID(2)
	pa := physmath.DeriveParameterID(3)
	x := physmath.NewBoundSymbol("x", px)
	y := physmath.NewBoundSymbol("y", py)
	z := physmath.NewBoundSymbol("z", pz)
	a := physmath.NewBoundSymbol("a", pa)
	declFn := mintFn(s, []physmath.ParameterID{px, py, pz}, x)
	// Chart over a fresh space.
	desig, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol("frw_cartesian_space"))
	if err != nil {
		t.Fatal(err)
	}
	spaceRaw, err := physmath.Collection{ElementRefs: []artifact.ArtifactRef{desig}}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	space, err := s.Put(physmath.FamilyCollection, spaceRaw)
	if err != nil {
		t.Fatal(err)
	}
	coordRef := func(e physmath.Expr) artifact.ArtifactRef {
		t.Helper()
		r, err := physmath.PutExpression(s, e)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	chart := physmath.CoordinateSystem{Dimension: 3,
		CoordinateRefs: []artifact.ArtifactRef{coordRef(x), coordRef(y), coordRef(z)},
		SpaceRef:       space}
	if err := chart.Validate(); err != nil {
		t.Fatal(err)
	}
	chartRef, err := s.Put(physmath.FamilyCoordinateSystem, chart.Encode())
	if err != nil {
		t.Fatal(err)
	}
	mapping := physmath.CoordParamMap{Entries: []physmath.BoundParameterRef{
		{DeclaringArtifactRef: declFn, ParameterID: px},
		{DeclaringArtifactRef: declFn, ParameterID: py},
		{DeclaringArtifactRef: declFn, ParameterID: pz},
	}}
	// Metric diag(a²,a²,a²) via the symmetric-bilinear form.
	a2 := physmath.Pow(a, physmath.MustRational(2, 1))
	mref := diagTensor(s, []physmath.Expr{a2, a2, a2})
	// General inverse (never a diagonal shortcut): det + NEQ condition.
	detRef, invComps, _, err := generalInverseForTest(t, s, mref)
	if err != nil {
		t.Fatal(err)
	}
	_ = invComps
	detExpr := resolveTestExpr(t, s, detRef)
	cond := condNEQ(s, detExpr)
	// Levi-Civita: flat chart implies vanishing coefficients, verified
	// below rather than assumed.
	connRef, connDeriv, err := physmath.LeviCivita(s, mref, chartRef, mapping, declFn,
		[]artifact.ArtifactRef{cond})
	if err != nil {
		t.Fatalf("FRW levi-civita failed: %v", err)
	}
	_, cpayload, err := s.Lookup(connRef)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := physmath.DecodeAffineConnection(cpayload)
	if err != nil {
		t.Fatal(err)
	}
	if !conn.ClaimSymmetric {
		t.Fatal("flat-chart symmetry not established")
	}
	for i, r := range conn.CoefficientRefs {
		if got := resolveTestExpr(t, s, r); !got.Equal(physmath.MustRational(0, 1)) {
			t.Fatalf("FRW coefficient %d is not exactly zero", i)
		}
	}
	if err := physmath.Replay(s, connDeriv, nil); err != nil {
		t.Fatalf("FRW connection derivation does not replay: %v", err)
	}
	if err := physmath.CheckAffineConnection(s, connRef, &physmath.ConnectionSources{
		MetricRef: mref, Mapping: mapping, DeclaringFnRef: declFn,
		Conds: []artifact.ArtifactRef{cond},
	}); err != nil {
		t.Fatalf("FRW connection check failed: %v", err)
	}
	// Test covector ω = [x, y, z]: ∇_μ ω_ν = δ_μν by hand (partials only,
	// connection terms vanish against zero coefficients).
	mkIdx := func(label string) artifact.ArtifactRef {
		p, err := physmath.Index{Label: label, Variance: physmath.VarianceCovariant, Space: &space}.Encode()
		if err != nil {
			t.Fatal(err)
		}
		r, err := s.Put(physmath.FamilyIndex, p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	omega := mkTensorRef(s, t, []artifact.ArtifactRef{mkIdx("nu")}, []physmath.Expr{x, y, z})
	omegaField, err := s.Put(physmath.FamilyTensorField,
		physmath.TensorField{TensorRef: omega, CoordinateSystemRef: chartRef}.Encode())
	if err != nil {
		t.Fatal(err)
	}
	wres, wderiv, err := physmath.EvaluateCovariantDerivative(s, omegaField, connRef, mapping, declFn, nil)
	if err != nil {
		t.Fatalf("FRW covector derivative failed: %v", err)
	}
	wfield := decodeTensorField(s, t, wres)
	if wfield.CoordinateSystemRef != chartRef {
		t.Fatal("FRW result chart not preserved")
	}
	wtensor := decodeTensor(s, t, wfield.TensorRef)
	if wtensor.Rank != 2 || len(wtensor.ComponentRefs) != 9 {
		t.Fatal("FRW result shape wrong")
	}
	// Hand golden: ∇_μ ω_ν = δ_μν (partials of [x,y,z]; Γ = 0).
	one := physmath.MustRational(1, 1)
	zero := physmath.MustRational(0, 1)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := zero
			if i == j {
				want = one
			}
			if got := resolveTestExpr(t, s, wtensor.ComponentRefs[i*3+j]); !got.Equal(want) {
				t.Fatalf("FRW component (%d,%d) is not hand-derived Kronecker", i, j)
			}
		}
	}
	// Certificate binds field, connection, result, derivation.
	wderivRef, err := s.Put(physmath.FamilyDerivation, mustDerivationBytes(t, wderiv))
	if err != nil {
		t.Fatal(err)
	}
	rec := physmath.CovariantEvaluation{FieldRef: omegaField, ConnectionRef: connRef,
		ResultRef: wres, DerivationRef: wderivRef}
	if err := rec.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := physmath.CheckCovariantEvaluation(s, rec, mapping, declFn); err != nil {
		t.Fatalf("FRW record rejected: %v", err)
	}
}

func mkTensorRef(s physmath.Store, t *testing.T, idx []artifact.ArtifactRef, comps []physmath.Expr) artifact.ArtifactRef {
	t.Helper()
	crefs := make([]artifact.ArtifactRef, len(comps))
	for i, e := range comps {
		r, err := physmath.PutExpression(s, e)
		if err != nil {
			t.Fatal(err)
		}
		crefs[i] = r
	}
	tn := physmath.Tensor{Rank: 1, Dimension: 3, IndexRefs: idx, ComponentRefs: crefs}
	if err := tn.Validate(); err != nil {
		t.Fatal(err)
	}
	ref, err := s.Put(physmath.FamilyTensor, tn.Encode())
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func decodeTensorField(s physmath.Store, t *testing.T, ref artifact.ArtifactRef) physmath.TensorField {
	t.Helper()
	_, payload, err := s.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	tf, err := physmath.DecodeTensorField(payload)
	if err != nil {
		t.Fatal(err)
	}
	return tf
}

func decodeTensor(s physmath.Store, t *testing.T, ref artifact.ArtifactRef) physmath.Tensor {
	t.Helper()
	_, payload, err := s.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	tn, err := physmath.DecodeTensor(payload)
	if err != nil {
		t.Fatal(err)
	}
	return tn
}

func mustDerivationBytes(t *testing.T, d physmath.Derivation) []byte {
	t.Helper()
	raw, err := d.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// generalInverseForTest runs InverseMatrix with an auto-minted exact
// NEQ(det,0) condition (an explicitly recorded assumption).
func generalInverseForTest(t *testing.T, s physmath.Store, mref artifact.ArtifactRef) (artifact.ArtifactRef, []artifact.ArtifactRef, physmath.Derivation, error) {
	t.Helper()
	detRef, _, err := physmath.GeneralDeterminant(s, mref)
	if err != nil {
		t.Fatal(err)
	}
	detExpr := resolveTestExpr(t, s, detRef)
	cond := condNEQ(s, detExpr)
	return physmath.InverseMatrix(s, mref, []artifact.ArtifactRef{cond})
}

func resolveTestExpr(t *testing.T, s physmath.Store, ref artifact.ArtifactRef) physmath.Expr {
	t.Helper()
	fam, payload, err := s.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyExpression {
		t.Fatalf("not expression/1: %v", fam)
	}
	e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
	if err != nil {
		t.Fatal(err)
	}
	return e
}
