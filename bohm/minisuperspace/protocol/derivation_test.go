package protocol

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	ms "github.com/PithomLabs/phys-lib/bohm/minisuperspace"
	"github.com/PithomLabs/phys-math"
)

func mustResolve(t *testing.T, s physmath.Store, ref artifact.ArtifactRef) physmath.Expr {
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

func simplify(t *testing.T, s physmath.Store, ref artifact.ArtifactRef, conds []artifact.ArtifactRef) artifact.ArtifactRef {
	t.Helper()
	res, err := physmath.Apply(s, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprSimplify,
		InputRefs: []artifact.ArtifactRef{ref}, ConditionRefs: conds,
	})
	if err != nil {
		t.Fatal(err)
	}
	return res.OutputRef
}

func boundSym(t *testing.T, s *physmath.Store, display string, pid physmath.ParameterID) physmath.Expr {
	t.Helper()
	_ = s
	return physmath.NewBoundSymbol(display, pid)
}

// TestWDWSecondPhiDerivative: ∂²_φR = 2 for R = a²+φ²+1 (paper 74, first piece).
func TestWDWSecondPhiDerivative(t *testing.T) {
	f := ms.Build()
	// Higher-order differentiation manufactures Pow(φ,0) intermediates, so
	// the exact NEQ(φ,0) condition is required for full simplification.
	simp := simplify(t, f.Store, f.D2PhiRef,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["phiNonzero"]})
	if got := mustResolve(t, f.Store, simp); !got.Equal(physmath.MustRational(2, 1)) {
		t.Fatal("second phi derivative of R is not hand-derived 2")
	}
}

// TestOperatorCompositionVsSquare: (a∂_a)²R = 4a², and crucially NOT
// a²∂²_aR = 2a². This test fails on the composition/ordering bug.
func TestOperatorCompositionVsSquare(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	simp := simplify(t, f.Store, f.AdaRef,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["phiNonzero"]})
	// Hand-derived factored form: (a∂_a)²R = a·(∂_a(a·∂_aR)) with
	// ∂_aR = 2a gives a·(2a+2a) — the two identical summands are the two
	// product-rule terms. Since 2a+2a = 4a (rational fold inside the Add
	// is NOT performed: no x+x→2x rule exists), a·(2a+2a) = 4a² holds by
	// hand algebra requiring distributivity, which is deliberately absent
	// (same boundary class as the FRW a^6 pin). The golden pins the exact
	// factored bytes; the comment records the hand-verified equality.
	twoA := physmath.Mul(a, physmath.MustRational(2, 1))
	want, err := physmath.Canonicalize(physmath.Mul(a, physmath.Add(twoA, twoA)))
	if err != nil {
		t.Fatal(err)
	}
	if got := mustResolve(t, f.Store, simp); !got.Equal(want) {
		t.Fatal("(a∂_a)²R is not hand-derived a·(2a+2a)")
	}
	// Anti-golden: the a²∂²_a misimplementation yields Mul(a²,2), a
	// structurally distinct shape that must never match.
	wrong, err := physmath.Canonicalize(physmath.Mul(physmath.MustRational(2, 1),
		physmath.Pow(a, physmath.MustRational(2, 1))))
	if err != nil {
		t.Fatal(err)
	}
	if mustResolve(t, f.Store, simp).Equal(wrong) {
		t.Fatal("result matches the a²∂²_a misimplementation")
	}
	// The factored golden must not already be the collapsed 4a² form:
	// collapse requires distributivity, absent by firewall design.
	collapsed, err := physmath.Canonicalize(physmath.Mul(physmath.MustRational(4, 1),
		physmath.Pow(a, physmath.MustRational(2, 1))))
	if err != nil {
		t.Fatal(err)
	}
	if mustResolve(t, f.Store, simp).Equal(collapsed) {
		t.Fatal("factored form unexpectedly collapsed (distributivity leaked in)")
	}
}

// TestGuidancePartials: ∂_φS = a and ∂_aS = φ for S = a·φ (paper 75 RHS).
func TestGuidancePartials(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	pp := physmath.DeriveParameterID(1)
	a := physmath.NewBoundSymbol("a", pa)
	phi := physmath.NewBoundSymbol("phi", pp)
	_ = boundSym(t, &f.Store, "x", pa)
	if got := mustResolve(t, f.Store, simplify(t, f.Store, f.DPhiSRef, nil)); !got.Equal(a) {
		t.Fatal("phase phi-partial is not hand-derived a")
	}
	if got := mustResolve(t, f.Store, simplify(t, f.Store, f.DaSRef, nil)); !got.Equal(phi) {
		t.Fatal("phase a-partial is not hand-derived phi")
	}
}

// TestQuantumPressureChain: p_Q for amplitude A(a) = a²+1 replays and
// simplifies to the hand-derived quotient form (paper 76-77).
func TestQuantumPressureChain(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	kap := physmath.NewEphemeralSymbol("kappa2")
	ampRaw := physmath.Add(physmath.Pow(a, physmath.MustRational(2, 1)), physmath.MustRational(1, 1))
	// Hand-derived Q = κ̄²/2 · a^-6 · W · A^-1 with W the factored
	// (a∂_a)²A form. ∂_aQ by the product rule over the four factors:
	//   T1 = κ̄²/2 · ∂_aW · a^-6 · A^-1          (W differentiated)
	//   T2 = κ̄²/2 · W · (-6a^-7) · A^-1        (a^-6 differentiated, W1 rule)
	//   T3 = κ̄²/2 · W · a^-6 · (-A^-2·∂_aA)    (A^-1 differentiated, W1 rule)
	// with Wd = ∂_aW the simplified factored derivative below. Each term
	// is traceable to one product-rule factor; Wd itself is covered by the
	// ADA factored golden, and negative powers by the W1 rule goldens.
	// Canonical child order matters for byte equality: Add sorts the
	// rational 1 before Pow (tag order), matching engine output.
	amp, err := physmath.Canonicalize(ampRaw)
	if err != nil {
		t.Fatal(err)
	}
	neg6 := physmath.Pow(a, physmath.MustRational(-6, 1))
	neg1 := physmath.Pow(amp, physmath.MustRational(-1, 1))
	khalf := physmath.Mul(physmath.MustRational(1, 2), kap)
	wd := physmath.Add(
		physmath.Mul(a, physmath.MustRational(4, 1)),
		physmath.Add(
			physmath.Mul(a, physmath.MustRational(2, 1)),
			physmath.Mul(a, physmath.MustRational(2, 1))),
	)
	w := physmath.Add(
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(2, 1)))
	t1 := physmath.Mul(khalf, wd, neg6, neg1)
	t2 := physmath.Mul(a, khalf,
		physmath.MustRational(-6, 1), w,
		physmath.Pow(a, physmath.MustRational(-7, 1)), neg1)
	t3 := physmath.Mul(a, khalf,
		physmath.MustRational(-1, 1),
		physmath.Mul(a, physmath.MustRational(2, 1)), w,
		neg6, physmath.Pow(amp, physmath.MustRational(-2, 1)))
	dq, err := physmath.Canonicalize(physmath.Add(t1, t2, t3))
	if err != nil {
		t.Fatal(err)
	}
	want, err := physmath.Canonicalize(physmath.Mul(
		physmath.MustRational(-1, 3),
		physmath.Pow(a, physmath.MustRational(-2, 1)), dq))
	if err != nil {
		t.Fatal(err)
	}
	simp := simplify(t, f.Store, f.PQRef,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["phiNonzero"]})
	// Associativity normalization lives on both sides: SIMPLIFY preserves
	// rewrite-history nesting (it sorts but never flattens), so compare
	// canonical closures. All factors, exponents, bases, and conditions are
	// still asserted exactly; only Add/Mul associativity is normalized.
	gotRaw := mustResolve(t, f.Store, simp)
	got, err := physmath.Canonicalize(gotRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatal("pQ chain is not the hand-derived quotient expansion")
	}
	// Determinism: rebuild produces the identical reference.
	f2 := ms.Build()
	simp2 := simplify(t, f2.Store, f2.PQRef,
		[]artifact.ArtifactRef{f2.Conds["aNonzero"], f2.Conds["phiNonzero"]})
	got2 := mustResolve(t, f2.Store, simp2)
	rb1, _ := mustResolve(t, f.Store, simp).Encode()
	rb2, _ := got2.Encode()
	if string(rb1) != string(rb2) {
		t.Fatal("pQ chain nondeterministic across fixtures")
	}
}

// TestWDWResidual: ĤR with k=0, Λ=0, U opaque simplifies to the
// hand-derived three-summand form (paper 74 applied, not solved).
func TestWDWResidual(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	kap := physmath.NewEphemeralSymbol("kappa2")
	u := physmath.NewEphemeralSymbol("Uphi")
	twoA := physmath.Mul(a, physmath.MustRational(2, 1))
	sW := physmath.Mul(a, physmath.Add(twoA, twoA))
	want, err := physmath.Canonicalize(physmath.Add(
		physmath.Mul(physmath.MustRational(-1, 2),
			physmath.Pow(a, physmath.MustRational(-3, 1)),
			physmath.MustRational(2, 1)),
		physmath.Mul(
			physmath.Mul(physmath.MustRational(1, 2), kap),
			physmath.Pow(a, physmath.MustRational(-3, 1)), sW),
		physmath.Mul(physmath.Pow(a, physmath.MustRational(3, 1)), u),
	))
	if err != nil {
		t.Fatal(err)
	}
	simp := simplify(t, f.Store, f.Residual,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["phiNonzero"]})
	gotRaw := mustResolve(t, f.Store, simp)
	got, err := physmath.Canonicalize(gotRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatal("WDW residual is not the hand-derived three-summand form")
	}
}

// TestTQAssembly: T_Q^00 builds from ρQ, pQ, U^0, g^00 with exact
// component references (paper 78, comoving frame).
func TestTQAssembly(t *testing.T) {
	f := ms.Build()
	e := mustResolve(t, f.Store, f.TQ00Ref)
	if e.Kind != physmath.TagAdd || len(e.Children) != 2 {
		t.Fatal("TQ00 must be a two-summand Add")
	}
	// Second summand is the negated -pQ·g^00 term.
	if e.Children[1].Kind != physmath.TagNeg {
		t.Fatal("TQ00 second summand must be the negated pressure term")
	}
	// U^0 is Mul(1, N^-1): selector component times inverse lapse.
	ue := mustResolve(t, f.Store, f.U0Ref)
	if ue.Kind != physmath.TagMul {
		t.Fatal("U0 must be a Mul construction")
	}
	foundPow := false
	for _, c := range ue.Children {
		if c.Kind == physmath.TagPow {
			foundPow = true
		}
	}
	if !foundPow {
		t.Fatal("U0 must carry the inverse-lapse power")
	}
}

// TestFRWVolumeProvenance: diag(a²,a²,a²) determinant + a>0 volume factor
// carry verified provenance (paper 67 spatial factor).
func TestFRWVolumeProvenance(t *testing.T) {
	f := ms.Build()
	fam, _, err := f.Store.Lookup(f.FRWVolRef)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyVolumeDensity {
		t.Fatal("FRW volume is not volume_density/1")
	}
	v, err := physmath.DecodeVolumeDensity(mustPayload(t, f, f.FRWVolRef))
	if err != nil {
		t.Fatal(err)
	}
	if err := physmath.CheckVolumeDensity(f.Store, v); err != nil {
		t.Fatalf("FRW volume provenance failed: %v", err)
	}
	// Product form preserved: det is not collapsed to a^6.
	sixth, err := physmath.PutExpression(f.Store,
		physmath.Pow(physmath.NewEphemeralSymbol("a"), physmath.MustRational(6, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if f.FRWDetRef == sixth {
		t.Fatal("FRW determinant collapsed to a^6 without an exponent rule")
	}
}

func mustPayload(t *testing.T, f *ms.Fixture, ref artifact.ArtifactRef) []byte {
	t.Helper()
	_, payload, err := f.Store.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// TestVerifyReplaysAllEvidence: the Verify entry replays every stored
// derivation; failure anywhere fails the protocol.
func TestVerifyReplaysAllEvidence(t *testing.T) {
	f := ms.Build()
	if err := f.Verify(); err != nil {
		t.Fatalf("fixture Verify failed: %v", err)
	}
}

// TestLimitationRecord: equation-solving, time evolution, full WDW, and
// functional variation have no executable evidence in this fixture.
func TestLimitationRecord(t *testing.T) {
	f := ms.Build()
	// Every stored derivation replays (executed evidence is real)...
	for i, dref := range f.Derivs {
		_, payload, err := f.Store.Lookup(dref)
		if err != nil {
			t.Fatal(err)
		}
		d, err := physmath.DecodeDerivation(payload)
		if err != nil {
			t.Fatal(err)
		}
		if err := physmath.Replay(f.Store, d, nil); err != nil {
			t.Fatalf("fixture derivation %d does not replay: %v", i, err)
		}
	}
	// ...but no derivation solves for ψ, evolves time, or varies functionally:
	// assert the limitation by checking no SELECT-BRANCH or unknown rule
	// appears in any fixture derivation (executable vocabulary audit).
	for i, dref := range f.Derivs {
		_, payload, err := f.Store.Lookup(dref)
		if err != nil {
			t.Fatal(err)
		}
		d, err := physmath.DecodeDerivation(payload)
		if err != nil {
			t.Fatal(err)
		}
		for _, st := range d.Steps {
			switch st.RuleID {
			case physmath.RuleExprCanon, physmath.RuleExprSimplify, physmath.RuleDifferentiate:
			default:
				t.Fatalf("fixture derivation %d uses unexpected rule %s", i, st.RuleID)
			}
		}
	}
}

// TestReducedVariationAnsatz: the γ-contraction projection executes on a
// 2x2 identity ansatz with hand-derived golden; the spatial integral and
// fiducial-volume normalization stay explicit limitations (no integration
// rule exists).
func TestReducedVariationAnsatz(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	ia, ib := ansatzIndices(t, f)
	gamma := ansatzTensor(t, f, ia, ib, []int64{1, 0, 0, 1})
	vary := ansatzTensor(t, f, ia, ib, []int64{1, 2, 3, 4})
	res, d, err := physmath.ReduceVariation(f.Store, gamma, vary, a)
	if err != nil {
		t.Fatalf("reduced variation failed: %v", err)
	}
	// Hand-derived: 2a·1·1 + 0 + 0 + 2a·1·4.
	want, err := physmath.Canonicalize(physmath.Add(
		physmath.Mul(physmath.MustRational(2, 1), a),
		physmath.Mul(physmath.MustRational(2, 1), a, physmath.MustRational(4, 1)),
	))
	if err != nil {
		t.Fatal(err)
	}
	if got := mustResolve(t, f.Store, res); !got.Equal(want) {
		t.Fatal("reduced integrand is not hand-derived 2a + 2a·4")
	}
	if len(d.Steps) != 2 {
		t.Fatal("reduction derivation is not the two-step shape")
	}
	dref, err := f.Store.Put(physmath.FamilyDerivation, mustDerivationBytes(t, d))
	if err != nil {
		t.Fatal(err)
	}
	rec := physmath.VariationReduction{AnsatzRef: gamma, VariationRef: vary,
		ReducedRef: res, DerivationRef: dref}
	if err := rec.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := physmath.CheckVariationReduction(f.Store, rec, a); err != nil {
		t.Fatalf("fresh reduction record rejected: %v", err)
	}
	// Cross-metric analog: edited variation density reuses old evidence.
	edited := ansatzTensor(t, f, ia, ib, []int64{1, 2, 3, 5})
	stale := rec
	stale.VariationRef = edited
	if err := physmath.CheckVariationReduction(f.Store, stale, a); err == nil {
		t.Fatal("stale variation evidence accepted")
	}
}

func ansatzIndices(t *testing.T, f *ms.Fixture) (artifact.ArtifactRef, artifact.ArtifactRef) {
	t.Helper()
	desig, err := physmath.PutExpression(f.Store, physmath.NewEphemeralSymbol("var_space"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := physmath.Collection{ElementRefs: []artifact.ArtifactRef{desig}}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	sp, err := f.Store.Put(physmath.FamilyCollection, raw)
	if err != nil {
		t.Fatal(err)
	}
	mk := func(label, v string) artifact.ArtifactRef {
		var variance physmath.IndexVariance
		if v == "co" {
			variance = physmath.VarianceCovariant
		} else {
			variance = physmath.VarianceContravariant
		}
		p, err := physmath.Index{Label: label, Variance: variance, Space: &sp}.Encode()
		if err != nil {
			t.Fatal(err)
		}
		r, err := f.Store.Put(physmath.FamilyIndex, p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	return mk("i", "co"), mk("j", "co")
}

func ansatzTensor(t *testing.T, f *ms.Fixture, ia, ib artifact.ArtifactRef, vals []int64) artifact.ArtifactRef {
	t.Helper()
	crefs := make([]artifact.ArtifactRef, len(vals))
	for i, v := range vals {
		r, err := physmath.PutExpression(f.Store, physmath.MustRational(v, 1))
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

func mustDerivationBytes(t *testing.T, d physmath.Derivation) []byte {
	t.Helper()
	raw, err := d.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
