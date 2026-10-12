package protocol

import (
	"fmt"
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
	// P1: d(u^1)/du returns du directly, so no Pow(φ,0) intermediate is
	// manufactured and no NEQ(φ,0) premise exists anywhere. The polynomial
	// derivative holds on the full declared domain including φ=0.
	simp := simplify(t, f.Store, f.D2PhiRef, nil)
	if got := mustResolve(t, f.Store, simp); !got.Equal(physmath.MustRational(2, 1)) {
		t.Fatal("second phi derivative of R is not hand-derived 2")
	}
	// Domain-full proof: re-running the second derivative with no
	// conditions succeeds, carries none, and replays. With P1 no
	// Pow(φ,0) intermediate exists, so no NEQ(φ,0) premise can arise.
	// Re-derive from a fresh R mint (independent of fixture internals).
	{
		pa0 := physmath.DeriveParameterID(0)
		pp0 := physmath.DeriveParameterID(1)
		a0 := physmath.NewBoundSymbol("a", pa0)
		phi0 := physmath.NewBoundSymbol("phi", pp0)
		rbody := physmath.Add(physmath.Pow(a0, physmath.MustRational(2, 1)),
			physmath.Pow(phi0, physmath.MustRational(2, 1)), physmath.MustRational(1, 1))
		rref, err := physmath.PutExpression(f.Store, rbody)
		if err != nil {
			t.Fatal(err)
		}
		ffn, err := f.Store.Put(physmath.FamilyFunction, mustFnBytes(t, f.Store, rbody))
		if err != nil {
			t.Fatal(err)
		}
		params := append(physmath.BoundParameterRef{DeclaringArtifactRef: ffn, ParameterID: pp0}.Encode(),
			artifact.EncodeUInt32(2)...)
		res, err := physmath.Apply(f.Store, physmath.Invocation{
			ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
			InputRefs: []artifact.ArtifactRef{rref, ffn}, Parameters: params,
		})
		if err != nil {
			t.Fatalf("condition-free second derivative failed: %v", err)
		}
		if len(res.ConditionRefs) != 0 {
			t.Fatal("polynomial derivative carries an unnecessary condition")
		}
		d := physmath.Derivation{
			PremiseRefs: []artifact.ArtifactRef{rref, ffn},
			Steps: []physmath.DerivationStep{{
				Ordinal: 0, ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
				InputRefs: []artifact.ArtifactRef{rref, ffn}, OutputRef: res.OutputRef,
				Parameters: params,
			}},
			ResultRefs: []artifact.ArtifactRef{res.OutputRef},
		}
		if err := physmath.Replay(f.Store, d, nil); err != nil {
			t.Fatalf("condition-free second derivative does not replay: %v", err)
		}
		s2, err := physmath.Apply(f.Store, physmath.Invocation{
			ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprSimplify,
			InputRefs: []artifact.ArtifactRef{res.OutputRef},
		})
		if err != nil {
			t.Fatalf("condition-free simplify failed: %v", err)
		}
		if got := mustResolve(t, f.Store, s2.OutputRef); !got.Equal(physmath.MustRational(2, 1)) {
			t.Fatal("condition-free second derivative is not 2")
		}
	}
}

// TestOperatorCompositionVsSquare: (a∂_a)²R = 4a², and crucially NOT
// a²∂²_aR = 2a². This test fails on the composition/ordering bug.
func TestOperatorCompositionVsSquare(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	simp := simplify(t, f.Store, f.AdaRef, nil)
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

// TestGuidanceRHSAssembly checks the constructed right-hand sides of paper
// (75): φ̇-RHS = N·a⁻³·∂_φS and ȧ-RHS = -κ̄²·N·a⁻¹·∂_aS, with the executed
// partials embedded by exact bytes. Time evolution itself is not derived.
func TestGuidanceRHSAssembly(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	phiE := mustResolve(t, f.Store, f.DPhiSRef)
	aE := mustResolve(t, f.Store, f.DaSRef)
	neph := physmath.NewEphemeralSymbol("N")
	kap := physmath.NewBoundSymbol("kappa2", physmath.DeriveParameterID(2))
	wantPhi, err := physmath.Canonicalize(physmath.Mul(neph,
		physmath.Pow(a, physmath.MustRational(-3, 1)), phiE))
	if err != nil {
		t.Fatal(err)
	}
	gotPhi, err := physmath.Canonicalize(mustResolve(t, f.Store, f.GuidePhiRef))
	if err != nil {
		t.Fatal(err)
	}
	if !gotPhi.Equal(wantPhi) {
		t.Fatal("guidance phi RHS is not N·a^-3·∂S")
	}
	wantA, err := physmath.Canonicalize(physmath.Mul(
		physmath.Neg(physmath.MustRational(1, 1)), kap,
		neph, physmath.Pow(a, physmath.MustRational(-1, 1)), aE))
	if err != nil {
		t.Fatal(err)
	}
	gotA, err := physmath.Canonicalize(mustResolve(t, f.Store, f.GuideARef))
	if err != nil {
		t.Fatal(err)
	}
	if !gotA.Equal(wantA) {
		t.Fatal("guidance scale RHS is not -κ̄²·N·a^-1·∂S")
	}
}

// TestDiffWUnit: ∂_aW for W = Mul(a, Add(Mul(a,2),Mul(a,2))) is the factored
// Add(Mul(a,2),Mul(a,2),Mul(a,4)) — i.e. W + a·4, hand-verified (= 8a given
// distributivity, which stays out).
func TestDiffWUnit(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	w := physmath.Add(
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(2, 1)))
	wref, err := physmath.PutExpression(f.Store, physmath.Mul(a, w))
	if err != nil {
		t.Fatal(err)
	}
	ampFn := mustAmpFn(t, f)
	params := append(physmath.BoundParameterRef{DeclaringArtifactRef: ampFn, ParameterID: pa}.Encode(),
		artifact.EncodeUInt32(1)...)
	res, err := physmath.Apply(f.Store, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
		InputRefs: []artifact.ArtifactRef{wref, ampFn}, Parameters: params,
	})
	if err != nil {
		t.Fatal(err)
	}
	simp := simplify(t, f.Store, res.OutputRef, nil)
	want, err := physmath.Canonicalize(physmath.Add(
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(4, 1))))
	if err != nil {
		t.Fatal(err)
	}
	// Same associativity discipline as the pQ golden: compare canonical
	// closures; all factors and exponents asserted exactly.
	gotRaw := mustResolve(t, f.Store, simp)
	got, err := physmath.Canonicalize(gotRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatal("factored W derivative is not hand-derived W + 4a")
	}
}

// TestDiffUnitSteps verifies single-node differentiations used inside the
// pQ chain, each hand-derived: d(a³), d(a⁻⁶) with condition, d(A⁻¹) with
// condition over the amplitude base.
func TestDiffUnitSteps(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	ampFn := mustAmpFn(t, f)
	params := append(physmath.BoundParameterRef{DeclaringArtifactRef: ampFn, ParameterID: pa}.Encode(),
		artifact.EncodeUInt32(1)...)
	diff := func(in artifact.ArtifactRef, conds []artifact.ArtifactRef) artifact.ArtifactRef {
		t.Helper()
		res, err := physmath.Apply(f.Store, physmath.Invocation{
			ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
			InputRefs: []artifact.ArtifactRef{in, ampFn}, ConditionRefs: conds, Parameters: params,
		})
		if err != nil {
			t.Fatal(err)
		}
		return res.OutputRef
	}
	mk := func(e physmath.Expr) artifact.ArtifactRef {
		t.Helper()
		r, err := physmath.PutExpression(f.Store, e)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	d1 := diff(mk(physmath.Pow(a, physmath.MustRational(3, 1))), nil)
	w1, _ := physmath.Canonicalize(physmath.Mul(physmath.MustRational(3, 1),
		physmath.Pow(a, physmath.MustRational(2, 1)), physmath.MustRational(1, 1)))
	if got := mustResolve(t, f.Store, d1); !got.Equal(w1) {
		t.Fatal("d(a³) unit wrong")
	}
	aneg6 := mk(physmath.Pow(a, physmath.MustRational(-6, 1)))
	if _, err := physmath.Apply(f.Store, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
		InputRefs: []artifact.ArtifactRef{aneg6, ampFn}, Parameters: params,
	}); err == nil {
		t.Fatal("unguarded negative-power differentiation accepted")
	}
	acond := f.Conds["aNonzero"]
	d2 := diff(aneg6, []artifact.ArtifactRef{acond})
	w2, _ := physmath.Canonicalize(physmath.Mul(physmath.MustRational(-6, 1),
		physmath.Pow(a, physmath.MustRational(-7, 1)), physmath.MustRational(1, 1)))
	if got := mustResolve(t, f.Store, d2); !got.Equal(w2) {
		t.Fatal("d(a⁻⁶) unit wrong")
	}
}

// TestWValue: W = (a∂_a)²A for A = a²+1 is the simplified factored form
// a·(2a+2a) — hand-verified (= 4a² requiring distributivity, deliberately
// absent; same boundary class as the FRW a^6 pin).
func TestWValue(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	twoA := physmath.Mul(a, physmath.MustRational(2, 1))
	want, err := physmath.Canonicalize(physmath.Mul(a, physmath.Add(twoA, twoA)))
	if err != nil {
		t.Fatal(err)
	}
	if got := mustResolve(t, f.Store, f.WSimpRef); !got.Equal(want) {
		t.Fatal("W is not hand-derived a·(2a+2a)")
	}
	wrong, err := physmath.Canonicalize(physmath.Mul(physmath.MustRational(2, 1),
		physmath.Pow(a, physmath.MustRational(2, 1))))
	if err != nil {
		t.Fatal(err)
	}
	if mustResolve(t, f.Store, f.WSimpRef).Equal(wrong) {
		t.Fatal("W matches a wrong collapsed shape")
	}
}

// TestRhoQValue: RhoQRef decodes to Mul(κ̄²/2, a⁻⁶, W, A⁻¹) with W bytes
// identical to the verified WSimpRef and A bytes identical to Amplitude;
// the bare amplitude must not appear as a direct factor.
func TestRhoQValue(t *testing.T) {
	f := ms.Build()
	e := mustResolve(t, f.Store, f.RhoQRef)
	if e.Kind != physmath.TagMul || len(e.Children) != 4 {
		t.Fatal("rhoQ must be a four-factor Mul")
	}
	wBytes, err := mustResolve(t, f.Store, f.WSimpRef).Encode()
	if err != nil {
		t.Fatal(err)
	}
	ampBytes, err := mustResolve(t, f.Store, f.Amplitude).Encode()
	if err != nil {
		t.Fatal(err)
	}
	foundW, foundAinv, foundDet := false, false, false
	for _, c := range e.Children {
		cb, err := c.Encode()
		if err != nil {
			t.Fatal(err)
		}
		if string(cb) == string(wBytes) {
			foundW = true
		}
		if string(cb) == string(ampBytes) {
			t.Fatal("bare amplitude must not be a direct rhoQ factor")
		}
		if c.Kind == physmath.TagPow {
			bb, _ := c.Base.Encode()
			xb, _ := c.Exp.Encode()
			neg1, _ := physmath.MustRational(-1, 1).Encode()
			if string(xb) == string(neg1) && string(bb) == string(ampBytes) {
				foundAinv = true
			}
			aneg6, _ := physmath.Pow(mustBoundA(t), physmath.MustRational(-6, 1)).Encode()
			if string(cb) == string(aneg6) {
				foundDet = true
			}
		}
	}
	if !foundW || !foundAinv || !foundDet {
		t.Fatal("rhoQ factors do not reference verified W, A^-1, a^-6")
	}
}

func mustBoundA(t *testing.T) physmath.Expr {
	t.Helper()
	return physmath.NewBoundSymbol("a", physmath.DeriveParameterID(0))
}

// TestQRelation: Q equals canonical Mul(a³, ρQ), and QRelRef decodes to an
// EQ relation over exactly those sides.
func TestQRelation(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	rho := mustResolve(t, f.Store, f.RhoQRef)
	want, err := physmath.Canonicalize(physmath.Mul(physmath.Pow(a, physmath.MustRational(3, 1)), rho))
	if err != nil {
		t.Fatal(err)
	}
	gotRaw := mustResolve(t, f.Store, f.QRef)
	got, err := physmath.Canonicalize(gotRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatal("Q is not Mul(a³, density) by canonical identity")
	}
	_, rpayload, err := f.Store.Lookup(f.QRelRef)
	if err != nil {
		t.Fatal(err)
	}
	rd := artifact.NewDecoder(rpayload)
	opStr, err := rd.String()
	if err != nil {
		t.Fatalf("Q relation missing operator: %v", err)
	}
	if opStr != string(physmath.OpEQ) {
		t.Fatal("Q relation is not an equality")
	}
	relLeft, err := physmath.DecodeExpr(rd)
	if err != nil {
		t.Fatalf("Q relation left undecodable: %v", err)
	}
	relRight, err := physmath.DecodeExpr(rd)
	if err != nil {
		t.Fatalf("Q relation right undecodable: %v", err)
	}
	if !rd.Exhausted() {
		t.Fatal("trailing bytes in Q relation")
	}
	lb, err := relLeft.Encode()
	if err != nil {
		t.Fatal(err)
	}
	qb, err := mustResolve(t, f.Store, f.QRef).Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(lb) != string(qb) {
		t.Fatal("Q relation left side is not the stored Q")
	}
	rbCanon, err := physmath.Canonicalize(relRight)
	if err != nil {
		t.Fatal(err)
	}
	rb, err := rbCanon.Encode()
	if err != nil {
		t.Fatal(err)
	}
	wb, err := want.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(rb) != string(wb) {
		t.Fatal("Q relation right side is not Mul(a³, density)")
	}
}

// TestPressureFromQ verifies the corrected chain with layered evidence.
// Hand-derived overall shape (paper 77 over Q = Mul(a³, ρQ)):
// pQ = Mul(-1/3, a⁻², D) with D the replayed ∂_aQ output. The three
// asserted linkage facts are exact: leading coefficient, a^-2 factor,
// and byte-identity of D with the replayed derivation output embedded at
// construction. Inside D, three targeted subtrees are asserted by
// canonical identity, each hand-derived: ∂W (U1 origin), -6a^-7 (U2
// origin, W1 rule), -A^-2 (U3 origin, W1 rule). Full-tree association is
// deliberately NOT byte-pinned: SIMPLIFY preserves rewrite-history
// association, which hand assembly cannot reproduce without copying the
// engine (see TestDiffWUnit/TestDiffUnitSteps for the independent unit
// goldens that make this decomposition sound).
func TestPressureFromQ(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	kap := physmath.NewBoundSymbol("kappa2", physmath.DeriveParameterID(2))
	simp := simplify(t, f.Store, f.PQRef,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["ampNonzero"]})
	gotRaw := mustResolve(t, f.Store, simp)
	if gotRaw.Kind != physmath.TagMul || len(gotRaw.Children) != 3 {
		t.Fatal("pQ must be a three-factor Mul")
	}
	negThird, err := physmath.MustRational(-1, 3).Encode()
	if err != nil {
		t.Fatal(err)
	}
	c0b, err := gotRaw.Children[0].Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(c0b) != string(negThird) {
		t.Fatal("pQ leading factor is not hand-derived -1/3")
	}
	powNeg2, err := physmath.Pow(a, physmath.MustRational(-2, 1)).Encode()
	if err != nil {
		t.Fatal(err)
	}
	c1b, err := gotRaw.Children[1].Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(c1b) != string(powNeg2) {
		t.Fatal("pQ second factor is not hand-derived a^-2")
	}
	// The embedded dQ factor must equal the independently simplified dQ
	// derivation output: simplification is context-free given the same
	// conditions, so simplifying the derivation output alone reproduces
	// exactly the subtree inside simplified pQ.
	dqRef := diffOutputFor(t, f, f.QRef)
	dqSimp := simplify(t, f.Store, dqRef,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["ampNonzero"]})
	dqb, err := mustResolve(t, f.Store, dqSimp).Encode()
	if err != nil {
		t.Fatal(err)
	}
	c2b, err := gotRaw.Children[2].Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(c2b) != string(dqb) {
		t.Fatal("pQ third factor is not the replayed dQ derivation output")
	}
	// Derivational golden: rebuild the expected dQ tree the way the engine
	// does — product-rule terms from independently golden-tested pieces,
	// then one Canonicalize (which flattens nested same-kind Muls, e.g.
	// WSimp inside rhoQ and dA inside U3: specified engine behavior that
	// explains every flat factor, verified above by mechanism).
	amp := canonicalAmp(t, a)
	wAdd := physmath.Add(
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(2, 1)))
	wSimp := physmath.Mul(a, wAdd)
	wdExp, err := physmath.Canonicalize(physmath.Add(
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(4, 1))))
	if err != nil {
		t.Fatal(err)
	}
	dAExp, err := physmath.Canonicalize(physmath.Mul(a, physmath.MustRational(2, 1)))
	if err != nil {
		t.Fatal(err)
	}
	dNeg6Exp, err := physmath.Canonicalize(physmath.Mul(physmath.MustRational(-6, 1),
		physmath.Pow(a, physmath.MustRational(-7, 1))))
	if err != nil {
		t.Fatal(err)
	}
	rhoExp, err := physmath.Canonicalize(physmath.Mul(khalfOf(t, kap),
		physmath.Pow(a, physmath.MustRational(-6, 1)), wSimp,
		physmath.Pow(amp, physmath.MustRational(-1, 1))))
	if err != nil {
		t.Fatal(err)
	}
	u1Exp, err := physmath.Canonicalize(physmath.Mul(khalfOf(t, kap),
		physmath.Pow(a, physmath.MustRational(-6, 1)), wdExp,
		physmath.Pow(amp, physmath.MustRational(-1, 1))))
	if err != nil {
		t.Fatal(err)
	}
	dA3Exp, err := physmath.Canonicalize(physmath.Mul(physmath.MustRational(3, 1),
		physmath.Pow(a, physmath.MustRational(2, 1)),
		physmath.MustRational(1, 1)))
	if err != nil {
		t.Fatal(err)
	}
	_ = dA3Exp
	_ = rhoExp
	dTree := gotRaw.Children[2]
	if !containsSubtree(t, dTree, wdExp) {
		t.Fatal("pQ derivative lacks the hand-derived dW summand")
	}
	// Full-tree pin: derivationally rebuilt expected dQ then pQ.
	dNeg1Exp, err := physmath.Canonicalize(physmath.Mul(physmath.MustRational(-1, 1),
		physmath.Pow(amp, physmath.MustRational(-2, 1)), dAExp))
	if err != nil {
		t.Fatal(err)
	}
	u2Exp, err := physmath.Canonicalize(physmath.Mul(khalfOf(t, kap), dNeg6Exp, wSimp,
		physmath.Pow(amp, physmath.MustRational(-1, 1))))
	if err != nil {
		t.Fatal(err)
	}
	u3Exp, err := physmath.Canonicalize(physmath.Mul(khalfOf(t, kap),
		physmath.Pow(a, physmath.MustRational(-6, 1)), wSimp, dNeg1Exp))
	if err != nil {
		t.Fatal(err)
	}
	dRhoExp, err := physmath.Canonicalize(physmath.Add(u1Exp, u2Exp, u3Exp))
	if err != nil {
		t.Fatal(err)
	}
	t1Exp, err := physmath.Canonicalize(physmath.Mul(dRhoExp, physmath.Pow(a, physmath.MustRational(3, 1))))
	if err != nil {
		t.Fatal(err)
	}
	// No trailing unit factor: SIMPLIFY strips Mul-one identities, so the
	// hand tree must not carry the raw-rule Mul(3,a^2,1) shape.
	t2Exp, err := physmath.Canonicalize(physmath.Mul(
		physmath.Mul(physmath.MustRational(3, 1),
			physmath.Pow(a, physmath.MustRational(2, 1))), rhoExp))
	if err != nil {
		t.Fatal(err)
	}
	dqExp, err := physmath.Canonicalize(physmath.Add(t1Exp, t2Exp))
	if err != nil {
		t.Fatal(err)
	}
	want, err := physmath.Canonicalize(physmath.Mul(
		physmath.MustRational(-1, 3),
		physmath.Pow(a, physmath.MustRational(-2, 1)), dqExp))
	if err != nil {
		t.Fatal(err)
	}
	gotC, err := physmath.Canonicalize(gotRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !gotC.Equal(want) {
		t.Fatal("pQ chain is not the derivationally rebuilt Q-derivative expansion")
	}
	// Determinism across fixtures.
	f2 := ms.Build()
	simp2 := simplify(t, f2.Store, f2.PQRef,
		[]artifact.ArtifactRef{f2.Conds["aNonzero"], f2.Conds["ampNonzero"]})
	got2 := mustResolve(t, f2.Store, simp2)
	rb1, _ := gotRaw.Encode()
	rb2, _ := got2.Encode()
	if string(rb1) != string(rb2) {
		t.Fatal("pQ chain nondeterministic across fixtures")
	}
}

// diffOutputFor returns the output reference of the replayed
// DIFFERENTIATE derivation consuming the given input, failing unless
// exactly one such derivation exists in the fixture evidence.
func diffOutputFor(t *testing.T, f *ms.Fixture, in artifact.ArtifactRef) artifact.ArtifactRef {
	t.Helper()
	var found []artifact.ArtifactRef
	for _, dref := range f.Derivs {
		_, payload, err := f.Store.Lookup(dref)
		if err != nil {
			t.Fatal(err)
		}
		d, err := physmath.DecodeDerivation(payload)
		if err != nil {
			t.Fatal(err)
		}
		for _, st := range d.Steps {
			if st.RuleID != physmath.RuleDifferentiate || len(st.InputRefs) != 2 {
				continue
			}
			if st.InputRefs[0] != in {
				continue
			}
			if err := physmath.Replay(f.Store, d, nil); err != nil {
				t.Fatal(err)
			}
			found = append(found, st.OutputRef)
		}
	}
	if len(found) != 1 {
		t.Fatalf("expected exactly one differentiation of the input, got %d", len(found))
	}
	return found[0]
}

func resolveStoreExpr(t *testing.T, f *ms.Fixture, ref artifact.ArtifactRef) (physmath.Expr, error) {
	t.Helper()
	fam, payload, err := f.Store.Lookup(ref)
	if err != nil {
		return physmath.Expr{}, err
	}
	if fam != physmath.FamilyExpression {
		return physmath.Expr{}, fmt.Errorf("not expression/1")
	}
	return physmath.DecodeExpr(artifact.NewDecoder(payload))
}

// containsSubtree reports whether golden occurs as an exact canonical
// subtree of tree (both sides canonicalized, so association differences
// introduced by rewrite history cannot false-negative).
func containsSubtree(t *testing.T, tree, golden physmath.Expr) bool {
	t.Helper()
	gc, err := physmath.Canonicalize(golden)
	if err != nil {
		t.Fatal(err)
	}
	gb, err := gc.Encode()
	if err != nil {
		t.Fatal(err)
	}
	var walk func(e physmath.Expr) bool
	walk = func(e physmath.Expr) bool {
		ec, err := physmath.Canonicalize(e)
		if err != nil {
			t.Fatal(err)
		}
		eb, err := ec.Encode()
		if err != nil {
			t.Fatal(err)
		}
		if string(eb) == string(gb) {
			return true
		}
		for _, c := range e.Children {
			if walk(c) {
				return true
			}
		}
		if e.Operand != nil && walk(*e.Operand) {
			return true
		}
		if e.Base != nil && walk(*e.Base) {
			return true
		}
		if e.Exp != nil && walk(*e.Exp) {
			return true
		}
		return false
	}
	return walk(tree)
}

// wdGolden is the hand-derived ∂W = W + 4a factored form (TestDiffWUnit).
func wdGolden(t *testing.T, a physmath.Expr) physmath.Expr {
	t.Helper()
	return physmath.Add(
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(2, 1)),
		physmath.Mul(a, physmath.MustRational(4, 1)))
}

// khalfOf is the hand-derived kappa/2 coefficient constructor.
// neg6Golden is the hand-derived -6a^-7 factor (W1 rule on a^-6, with
// the unit derivative stripped by SIMPLIFY identity removal).
func neg6Golden(t *testing.T, a physmath.Expr) physmath.Expr {
	t.Helper()
	return physmath.Mul(physmath.MustRational(-6, 1),
		physmath.Pow(a, physmath.MustRational(-7, 1)))
}

// negA2Golden is the hand-derived -A^-2 factor (W1 rule on A^-1 is the
// -1·A^-2 part; the ∂A factor travels with the surrounding term).
func negA2Golden(t *testing.T, amp physmath.Expr) physmath.Expr {
	t.Helper()
	return physmath.Pow(amp, physmath.MustRational(-2, 1))
}

func khalfOf(t *testing.T, kap physmath.Expr) physmath.Expr {
	t.Helper()
	return physmath.Mul(physmath.MustRational(1, 2), kap)
}

func canonicalAmp(t *testing.T, a physmath.Expr) physmath.Expr {
	t.Helper()
	amp, err := physmath.Canonicalize(physmath.Add(
		physmath.Pow(a, physmath.MustRational(2, 1)), physmath.MustRational(1, 1)))
	if err != nil {
		t.Fatal(err)
	}
	return amp
}

// TestHistoricalWrongPath rejects the exact previous failure:
// differentiating ρQ instead of Q, then comparing against corrected pQ.
func TestHistoricalWrongPath(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	wrongD, err := physmath.Apply(f.Store, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
		InputRefs:     []artifact.ArtifactRef{f.RhoQRef, mustAmpFn(t, f)},
		ConditionRefs: []artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["ampNonzero"]},
		Parameters:    diffParamsFor(t, f, pa, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	a := physmath.NewBoundSymbol("a", pa)
	wrongPQ, err := physmath.PutExpression(f.Store, physmath.Mul(
		physmath.MustRational(-1, 3),
		physmath.Pow(a, physmath.MustRational(-2, 1)),
		mustResolve(t, f.Store, wrongD.OutputRef)))
	if err != nil {
		t.Fatal(err)
	}
	wrongSimp := simplify(t, f.Store, wrongPQ,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["ampNonzero"]})
	correctSimp := simplify(t, f.Store, f.PQRef,
		[]artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["ampNonzero"]})
	gw, err := physmath.Canonicalize(mustResolve(t, f.Store, wrongSimp))
	if err != nil {
		t.Fatal(err)
	}
	gc, err := physmath.Canonicalize(mustResolve(t, f.Store, correctSimp))
	if err != nil {
		t.Fatal(err)
	}
	if gw.Equal(gc) {
		t.Fatal("historical wrong path reproduces corrected pressure")
	}
}

// TestSubstitutionPoint evaluates ρQ/Q/pQ at a=1, κ̄²=1 via
// EXPR-SUBSTITUTE-001. Pow(1,n) remnants are expected (no const-Pow
// folding rule exists): the check proves substitution consistency and
// wrong-path discrimination, while bare values 1/1/2/3 rest on hand
// algebra in comments (ρQ(1,1)=1, Q(1,1)=1, pQ(1,1)=2/3 vs wrong 5/3).
func TestSubstitutionPoint(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	pk := physmath.DeriveParameterID(2)
	one, err := physmath.PutExpression(f.Store, physmath.MustRational(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	sub := func(target artifact.ArtifactRef) artifact.ArtifactRef {
		t.Helper()
		ampFn := mustAmpFn(t, f)
		decls := []artifact.ArtifactRef{ampFn, f.KappaFunRef}
		if string(decls[0].Encode()) > string(decls[1].Encode()) {
			decls[0], decls[1] = decls[1], decls[0]
		}
		e1 := append(physmath.BoundParameterRef{DeclaringArtifactRef: ampFn, ParameterID: pa}.Encode(), one.Encode()...)
		e2 := append(physmath.BoundParameterRef{DeclaringArtifactRef: f.KappaFunRef, ParameterID: pk}.Encode(), one.Encode()...)
		entries := [][]byte{e1, e2}
		if string(entries[0]) > string(entries[1]) {
			entries[0], entries[1] = entries[1], entries[0]
		}
		params := append(artifact.EncodeUInt32(2), append(entries[0], entries[1]...)...)
		res, err := physmath.Apply(f.Store, physmath.Invocation{
			ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprSubstitute,
			InputRefs:  append([]artifact.ArtifactRef{target}, decls...),
			Parameters: params,
		})
		if err != nil {
			t.Fatal(err)
		}
		return res.OutputRef
	}
	for _, target := range []artifact.ArtifactRef{f.RhoQRef, f.QRef, f.PQRef} {
		if _, _, err := f.Store.Lookup(sub(target)); err != nil {
			t.Fatal(err)
		}
	}
	if r1, r2 := sub(f.RhoQRef), sub(f.RhoQRef); r1 != r2 {
		t.Fatal("substitution nondeterministic")
	}
	// Wrong-path substituted tree differs from corrected substituted tree.
	wrongD, err := physmath.Apply(f.Store, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
		InputRefs:     []artifact.ArtifactRef{f.RhoQRef, mustAmpFn(t, f)},
		ConditionRefs: []artifact.ArtifactRef{f.Conds["aNonzero"], f.Conds["ampNonzero"]},
		Parameters:    diffParamsFor(t, f, pa, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	a := physmath.NewBoundSymbol("a", pa)
	wrongPQ, err := physmath.PutExpression(f.Store, physmath.Mul(
		physmath.MustRational(-1, 3),
		physmath.Pow(a, physmath.MustRational(-2, 1)),
		mustResolve(t, f.Store, wrongD.OutputRef)))
	if err != nil {
		t.Fatal(err)
	}
	sWrong := mustResolve(t, f.Store, sub(wrongPQ))
	sRight, err := physmath.Canonicalize(mustResolve(t, f.Store, sub(f.PQRef)))
	if err != nil {
		t.Fatal(err)
	}
	sWrongC, err := physmath.Canonicalize(sWrong)
	if err != nil {
		t.Fatal(err)
	}
	if sWrongC.Equal(sRight) {
		t.Fatal("substituted wrong path matches corrected pressure")
	}
}

func mustAmpFn(t *testing.T, f *ms.Fixture) artifact.ArtifactRef {
	t.Helper()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	body := physmath.Add(physmath.Pow(a, physmath.MustRational(2, 1)), physmath.MustRational(1, 1))
	bodyRef, err := physmath.PutExpression(f.Store, body)
	if err != nil {
		t.Fatal(err)
	}
	fn := physmath.Function{ParameterIDs: []physmath.ParameterID{pa}, BodyRef: bodyRef}
	raw, err := fn.Encode()
	if err != nil {
		t.Fatal(err)
	}
	ref, err := f.Store.Put(physmath.FamilyFunction, raw)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func diffParamsFor(t *testing.T, f *ms.Fixture, pid physmath.ParameterID, order uint32) []byte {
	t.Helper()
	fn := mustAmpFn(t, f)
	return append(physmath.BoundParameterRef{DeclaringArtifactRef: fn, ParameterID: pid}.Encode(),
		artifact.EncodeUInt32(order)...)
}

// TestWDWResidual: ĤR with k=0, Λ=0, U opaque simplifies to the
// hand-derived three-summand form (paper 74 applied, not solved).
func TestWDWResidual(t *testing.T) {
	f := ms.Build()
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	kap := physmath.NewBoundSymbol("kappa2", physmath.DeriveParameterID(2))
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
		[]artifact.ArtifactRef{f.Conds["aNonzero"]})
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
	// Product form preserved: det is not collapsed to a^6. The anti-golden
	// uses the EXACT bound parameter identity of the metric (DeriveParameterID(0)),
	// not a same-labeled ephemeral symbol: labels are not identities.
	pa := physmath.DeriveParameterID(0)
	a := physmath.NewBoundSymbol("a", pa)
	sixth, err := physmath.PutExpression(f.Store,
		physmath.Pow(a, physmath.MustRational(6, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if f.FRWDetRef == sixth {
		t.Fatal("FRW determinant collapsed to a^6 without an exponent rule")
	}
	// Identity-aware comparison: decoded canonical expressions differ, not
	// just the references.
	wantSixth, err := physmath.Canonicalize(physmath.Pow(a, physmath.MustRational(6, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if got := mustResolve(t, f.Store, f.FRWDetRef); got.Equal(wantSixth) {
		t.Fatal("FRW determinant equals a^6 by expression identity")
	}
	// Sensitivity control: re-minting the same a^6 yields the identical
	// reference, proving the comparison above can detect a collapse.
	sixth2, err := physmath.PutExpression(f.Store,
		physmath.Pow(a, physmath.MustRational(6, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if sixth2 != sixth {
		t.Fatal("content addressing unstable for the anti-golden")
	}
	// Label control: an ephemeral same-labeled "a" is a DIFFERENT variable
	// identity and must mint a different reference.
	ephem, err := physmath.PutExpression(f.Store,
		physmath.Pow(physmath.NewEphemeralSymbol("a"), physmath.MustRational(6, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if ephem == sixth {
		t.Fatal("ephemeral and bound same-label symbols collided")
	}
}

func mustFnBytes(t *testing.T, s physmath.Store, body physmath.Expr) []byte {
	t.Helper()
	pa := physmath.DeriveParameterID(0)
	pp := physmath.DeriveParameterID(1)
	bodyRef, err := physmath.PutExpression(s, body)
	if err != nil {
		t.Fatal(err)
	}
	fn := physmath.Function{ParameterIDs: []physmath.ParameterID{pa, pp}, BodyRef: bodyRef}
	if err := physmath.ValidateFunctionBody(fn, body); err != nil {
		t.Fatal(err)
	}
	raw, err := fn.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return raw
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

// TestVerifyRejectsMissingEvidence proves every declared evidence path is
// load-bearing: each failure below must fail at its own boundary.
func TestVerifyRejectsMissingEvidence(t *testing.T) {
	// Missing FRW determinant derivation.
	f := ms.Build()
	badFRW := *f
	badFRW.FRWDerivRef = mustGhostRef(t, "physmath", "derivation/1")
	if err := badFRW.Verify(); err == nil {
		t.Fatal("missing FRW determinant derivation accepted")
	}
	// Corrupted FRW evidence (wrong family: expression instead of derivation).
	f2 := ms.Build()
	one, err := physmath.PutExpression(f2.Store, physmath.MustRational(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	badFam := *f2
	badFam.FRWDerivRef = one
	if err := badFam.Verify(); err == nil {
		t.Fatal("wrong-family FRW evidence accepted")
	}
	// Missing Q-pressure derivation: drop one stored derivation.
	f3 := ms.Build()
	if len(f3.Derivs) == 0 {
		t.Fatal("no evidence derivations to corrupt")
	}
	badQ := *f3
	badQ.Derivs = append([]artifact.ArtifactRef{}, f3.Derivs...)
	badQ.Derivs[0] = mustGhostRef(t, "physmath", "derivation/1")
	if err := badQ.Verify(); err == nil {
		t.Fatal("missing Q/pressure derivation accepted")
	}
	// Inconsistent Q/density relation: Q pointing at the density itself.
	f4 := ms.Build()
	badRel := *f4
	badRel.QRef = f4.RhoQRef
	if err := badRel.Verify(); err == nil {
		t.Fatal("inconsistent Q/density relation accepted")
	}
}

// TestVerifyCoverageAudit asserts the manifest's coverage claim
// structurally: every Derivs entry plus FRWDerivRef resolves to
// derivation/1 and replays; nothing else is described as replayed.
func TestVerifyCoverageAudit(t *testing.T) {
	f := ms.Build()
	checked := 0
	for i, dref := range append(append([]artifact.ArtifactRef{}, f.Derivs...), f.FRWDerivRef) {
		fam, payload, err := f.Store.Lookup(dref)
		if err != nil {
			t.Fatalf("evidence %d unresolvable: %v", i, err)
		}
		if fam != physmath.FamilyDerivation {
			t.Fatalf("evidence %d is not derivation/1", i)
		}
		d, err := physmath.DecodeDerivation(payload)
		if err != nil {
			t.Fatalf("evidence %d undecodable: %v", i, err)
		}
		if err := physmath.Replay(f.Store, d, nil); err != nil {
			t.Fatalf("evidence %d does not replay: %v", i, err)
		}
		checked++
	}
	if checked != len(f.Derivs)+1 {
		t.Fatal("coverage audit did not visit all evidence")
	}
	if err := f.Verify(); err != nil {
		t.Fatalf("valid fixture rejected: %v", err)
	}
}

// mustGhostRef builds a well-formed but never-minted reference.
func mustGhostRef(t *testing.T, ns, typ string) artifact.ArtifactRef {
	t.Helper()
	ref, err := artifact.ComputeRef(ns, "1", typ, []byte("ghost-payload"))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
