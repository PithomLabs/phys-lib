package minisuperspace

import (
	"strings"
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// TestFRWCovariantFragment exercises the fixture-integrated covariant
// evidence: Build() constructs the chart, metric, connection, covector,
// evaluation, and certificate in the fixture store, and Verify() resolves,
// replays, and structurally checks them. There is no side store: the
// fixture is the single unambiguous evidence path.
//
// The flat chart makes all spatial partials vanish, so the Levi-Civita
// computation below returns vanishing coefficients. The test asserts those
// computed zeros against explicit expected values; vanishing is the
// verified computation result for this chart, not an assumed input.
func TestFRWCovariantFragment(t *testing.T) {
	f := Build()
	if err := f.Verify(); err != nil {
		t.Fatalf("fixture Verify failed: %v", err)
	}
	// Connection coefficients computed to exactly zero by the fixture.
	_, cpayload, err := f.Store.Lookup(f.FRWConnRef)
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
		if got := resolveTestExpr(t, f.Store, r); !got.Equal(physmath.MustRational(0, 1)) {
			t.Fatalf("FRW coefficient %d is not exactly zero", i)
		}
	}
	// Hand golden already enforced inside Verify (Kronecker components);
	// here assert the record round-trips through the store.
	fam, payload, err := f.Store.Lookup(f.FRWCovEvalRef)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyCovariantEvaluation {
		t.Fatal("evaluation record is not covariant_evaluation/1")
	}
	if _, err := physmath.DecodeCovariantEvaluation(payload); err != nil {
		t.Fatalf("evaluation record does not decode: %v", err)
	}
}

// TestFRWCovariantIdentities guards the parameter-identity discipline:
// the covariant fragment uses its own metric artifact over a disjoint
// scale-factor identity (same display label "a", different PID), so chart
// differentiation never confuses the scale factor with a coordinate. The
// determinant/volume path keeps the original metric.
func TestFRWCovariantIdentities(t *testing.T) {
	f := Build()
	if f.FRWCovMetricRef == f.FRWMetricRef {
		t.Fatal("covariant fragment reuses the volume-path metric identity")
	}
	// Both metrics decode to diagonal form with the same display labels
	// but different parameter identities: read the (0,0) components.
	compOf := func(ref artifact.ArtifactRef) physmath.Expr {
		t.Helper()
		_, payload, err := f.Store.Lookup(ref)
		if err != nil {
			t.Fatal(err)
		}
		m, err := physmath.DecodeSymmetricBilinear(payload)
		if err != nil {
			t.Fatal(err)
		}
		_, cp, err := f.Store.Lookup(m.ComponentRefs[0])
		if err != nil {
			t.Fatal(err)
		}
		e, err := physmath.DecodeExpr(artifact.NewDecoder(cp))
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	covComp := compOf(f.FRWCovMetricRef)
	volComp := compOf(f.FRWMetricRef)
	if covComp.Equal(volComp) {
		t.Fatal("fragment and volume metrics share component identities")
	}
	// The fragment scale factor is independent of every chart coordinate:
	// differentiating its square along x yields exactly zero (spatial
	// constancy through an independent parameter identity) — which is
	// what makes the flat-chart connection vanish. A PID collision would
	// produce 2a here instead.
	xPid := physmath.DeriveParameterID(0)
	params := append(physmath.BoundParameterRef{DeclaringArtifactRef: f.FRWDeclFnRef, ParameterID: xPid}.Encode(),
		artifact.EncodeUInt32(1)...)
	res, err := physmath.Apply(f.Store, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
		InputRefs:  []artifact.ArtifactRef{compOfRef(t, f, f.FRWCovMetricRef), f.FRWDeclFnRef},
		Parameters: params,
	})
	if err != nil {
		t.Fatal(err)
	}
	// The rule output is canonical, not simplified: Mul(2,a,0)-shaped.
	// Simplification (no conditions needed) annihilates it to bare zero.
	simp, err := physmath.Apply(f.Store, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprSimplify,
		InputRefs: []artifact.ArtifactRef{res.OutputRef},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := resolveTestExpr(t, f.Store, simp.OutputRef); !got.Equal(physmath.MustRational(0, 1)) {
		t.Fatal("fragment scale factor is not independent of chart coordinates")
	}
}

func compOfRef(t *testing.T, f *Fixture, ref artifact.ArtifactRef) artifact.ArtifactRef {
	t.Helper()
	_, payload, err := f.Store.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	m, err := physmath.DecodeSymmetricBilinear(payload)
	if err != nil {
		t.Fatal(err)
	}
	return m.ComponentRefs[0]
}

// TestFRWCovariantIdentityGuard proves the disjointness contract on the
// fixture's own chart: a tensor component binding a chart PID under a
// different display ("q" on x's PID) is rejected at the identity gate
// before any artifact is minted, while the committed fixture still
// verifies clean.
func TestFRWCovariantIdentityGuard(t *testing.T) {
	f := Build()
	if err := f.Verify(); err != nil {
		t.Fatalf("fixture Verify failed: %v", err)
	}
	// Resolve the fixture chart's coordinate PIDs through its declaring
	// function: position 0 is the x identity.
	xPid := physmath.DeriveParameterID(0)
	// Mint a rank-1 tensor over the chart space whose component reuses
	// x's PID for a distinct quantity "q".
	space := chartSpaceOf(t, f)
	ixPayload, err := physmath.Index{Label: "v", Variance: physmath.VarianceContravariant, Space: &space}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	ix, err := f.Store.Put(physmath.FamilyIndex, ixPayload)
	if err != nil {
		t.Fatal(err)
	}
	qComp, err := physmath.PutExpression(f.Store, physmath.NewBoundSymbol("q", xPid))
	if err != nil {
		t.Fatal(err)
	}
	tn := physmath.Tensor{Rank: 1, Dimension: 3,
		IndexRefs:     []artifact.ArtifactRef{ix},
		ComponentRefs: []artifact.ArtifactRef{qComp, qComp, qComp}}
	if err := tn.Validate(); err != nil {
		t.Fatal(err)
	}
	tref, err := f.Store.Put(physmath.FamilyTensor, tn.Encode())
	if err != nil {
		t.Fatal(err)
	}
	field, err := f.Store.Put(physmath.FamilyTensorField,
		physmath.TensorField{TensorRef: tref, CoordinateSystemRef: f.FRWChartRef}.Encode())
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = physmath.EvaluateCovariantDerivative(f.Store, field, f.FRWConnRef,
		f.FRWMapping, f.FRWDeclFnRef, nil)
	if err == nil {
		t.Fatal("collided identity accepted by evaluator")
	}
	if !strings.Contains(err.Error(), "reserved for coordinate") {
		t.Fatalf("wrong rejection boundary (expected identity gate): %v", err)
	}
	// The committed fixture is unaffected by the rejected attempt.
	if err := f.Verify(); err != nil {
		t.Fatalf("fixture Verify failed after rejected attempt: %v", err)
	}
}

func chartSpaceOf(t *testing.T, f *Fixture) artifact.ArtifactRef {
	t.Helper()
	_, payload, err := f.Store.Lookup(f.FRWChartRef)
	if err != nil {
		t.Fatal(err)
	}
	chart, err := physmath.DecodeCoordinateSystem(payload)
	if err != nil {
		t.Fatal(err)
	}
	return chart.SpaceRef
}

// TestFRWCovariantNegatives proves each failure boundary with fixture
// copies whose stores are shared (content-addressed) but whose record
// under test is mutated.
func TestFRWCovariantNegatives(t *testing.T) {
	f := Build()
	// Chart-mismatched connection: build a second chart and rewire the
	// record's connection through a same-shape connection on it.
	otherChart, otherConn := otherChartAndConn(t, f)
	badRec := mustDecodeEval(t, f, f.FRWCovEvalRef)
	badRec.ConnectionRef = otherConn
	raw := badRec.Encode()
	cref, err := f.Store.Put(physmath.FamilyCovariantEvaluation, raw)
	if err != nil {
		t.Fatal(err)
	}
	_ = cref
	_ = otherChart
	bad := *f
	bad.FRWCovEvalRef = cref
	if err := bad.Verify(); err == nil {
		t.Fatal("chart-mismatched connection accepted by Verify")
	}
	// Swapped result tensor.
	f2 := Build()
	alt, err := f2.Store.Put(physmath.FamilyTensor, mustAltTensor(t, f2))
	if err != nil {
		t.Fatal(err)
	}
	altField, err := f2.Store.Put(physmath.FamilyTensorField,
		physmath.TensorField{TensorRef: alt, CoordinateSystemRef: f2.FRWChartRef}.Encode())
	if err != nil {
		t.Fatal(err)
	}
	rec2 := mustDecodeEval(t, f2, f2.FRWCovEvalRef)
	rec2.ResultRef = altField
	raw2 := rec2.Encode()
	cref2, err := f2.Store.Put(physmath.FamilyCovariantEvaluation, raw2)
	if err != nil {
		t.Fatal(err)
	}
	bad2 := *f2
	bad2.FRWCovEvalRef = cref2
	if err := bad2.Verify(); err == nil {
		t.Fatal("substituted result accepted by Verify")
	}
	// Corrupted evaluation derivation.
	f3 := Build()
	ghost, err := artifact.ComputeRef("physmath", "1", "derivation/1", []byte("ghost-payload"))
	if err != nil {
		t.Fatal(err)
	}
	rec3 := mustDecodeEval(t, f3, f3.FRWCovEvalRef)
	rec3.DerivationRef = ghost
	raw3 := rec3.Encode()
	cref3, err := f3.Store.Put(physmath.FamilyCovariantEvaluation, raw3)
	if err != nil {
		t.Fatal(err)
	}
	bad3 := *f3
	bad3.FRWCovEvalRef = cref3
	if err := bad3.Verify(); err == nil {
		t.Fatal("corrupted derivation accepted by Verify")
	}
}

func mustDecodeEval(t *testing.T, f *Fixture, ref artifact.ArtifactRef) physmath.CovariantEvaluation {
	t.Helper()
	_, payload, err := f.Store.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := physmath.DecodeCovariantEvaluation(payload)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

// otherChartAndConn builds a second chart plus a zero connection over it,
// reusing the same dimension so only chart identity differs.
func otherChartAndConn(t *testing.T, f *Fixture) (artifact.ArtifactRef, artifact.ArtifactRef) {
	t.Helper()
	desig, err := physmath.PutExpression(f.Store, physmath.NewEphemeralSymbol("other_space"))
	if err != nil {
		t.Fatal(err)
	}
	spaceRaw, err := physmath.Collection{ElementRefs: []artifact.ArtifactRef{desig}}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	space, err := f.Store.Put(physmath.FamilyCollection, spaceRaw)
	if err != nil {
		t.Fatal(err)
	}
	px := physmath.DeriveParameterID(7)
	py := physmath.DeriveParameterID(8)
	pz := physmath.DeriveParameterID(9)
	coord := func(name string, pid physmath.ParameterID) artifact.ArtifactRef {
		r, err := physmath.PutExpression(f.Store, physmath.NewBoundSymbol(name, pid))
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	chart := physmath.CoordinateSystem{Dimension: 3,
		CoordinateRefs: []artifact.ArtifactRef{coord("u", px), coord("v", py), coord("w", pz)},
		SpaceRef:       space}
	if err := chart.Validate(); err != nil {
		t.Fatal(err)
	}
	chartRef, err := f.Store.Put(physmath.FamilyCoordinateSystem, chart.Encode())
	if err != nil {
		t.Fatal(err)
	}
	zero, err := physmath.PutExpression(f.Store, physmath.MustRational(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	coeffs := make([]artifact.ArtifactRef, 27)
	for i := range coeffs {
		coeffs[i] = zero
	}
	conn := physmath.AffineConnection{CoordRef: chartRef, Dimension: 3, CoefficientRefs: coeffs}
	raw := conn.Encode()
	connRef, err := f.Store.Put(physmath.FamilyAffineConnection, raw)
	if err != nil {
		t.Fatal(err)
	}
	return chartRef, connRef
}

// mustAltTensor builds a non-Kronecker rank-2 tensor for substitution tests.
func mustAltTensor(t *testing.T, f *Fixture) []byte {
	t.Helper()
	_, cpayload, err := f.Store.Lookup(f.FRWChartRef)
	if err != nil {
		t.Fatal(err)
	}
	chart, err := physmath.DecodeCoordinateSystem(cpayload)
	if err != nil {
		t.Fatal(err)
	}
	seven, err := physmath.PutExpression(f.Store, physmath.MustRational(7, 1))
	if err != nil {
		t.Fatal(err)
	}
	crefs := make([]artifact.ArtifactRef, 9)
	for i := range crefs {
		crefs[i] = seven
	}
	tn := physmath.Tensor{Rank: 2, Dimension: 3,
		IndexRefs:     []artifact.ArtifactRef{mustDerivIdx(t, f, chart), mustDerivIdx(t, f, chart)},
		ComponentRefs: crefs}
	if err := tn.Validate(); err != nil {
		t.Fatal(err)
	}
	return tn.Encode()
}

func mustDerivIdx(t *testing.T, f *Fixture, chart physmath.CoordinateSystem) artifact.ArtifactRef {
	t.Helper()
	p, err := physmath.Index{Label: "d", Variance: physmath.VarianceCovariant, Space: &chart.SpaceRef}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	ref, err := f.Store.Put(physmath.FamilyIndex, p)
	if err != nil {
		t.Fatal(err)
	}
	return ref
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
