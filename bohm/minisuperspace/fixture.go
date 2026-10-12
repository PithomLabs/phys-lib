// Package minisuperspace is the Dürr–Struyve quantum Einstein equations
// Section 6 (pp.11–12, eqs. 67–78) representation and partial-execution
// fixture for arXiv:2003.03839v2.
//
// Status discipline (see manifest.json): the Wheeler–DeWitt operator,
// guidance-equation right-hand sides, and quantum density/pressure pieces
// are EXECUTED on explicit test wavefunctions with replayed evidence.
// Time-derivative left-hand sides, equation solving, the full
// Wheeler–DeWitt theory, functional variation, and curvature are
// REPRESENTED or UNSUPPORTED as recorded — never claimed executed.
// Minisuperspace is the paper's symmetry-reduced toy model, not proof of
// the general theory.
package minisuperspace

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// Fixture holds every minted artifact of the minisuperspace evidence chain.
type Fixture struct {
	Store physmath.Store
	// Test wavefunction R(a,φ) = a²+φ²+1 and its declaring function.
	RRef    artifact.ArtifactRef // expression/1 test wavefunction
	RFunRef artifact.ArtifactRef // function/1 declaring (a,φ)
	// WDW operator pieces (74) applied to R.
	D2PhiRef artifact.ArtifactRef // expression/1 ∂²_φR = 2
	AdaRef   artifact.ArtifactRef // expression/1 (a∂_a)²R = 4a²
	Residual artifact.ArtifactRef // expression/1 ĤR residual, k=0 Λ=0
	// Test phase S(a,φ) = a·φ and guidance RHS pieces (75).
	SRef     artifact.ArtifactRef
	DPhiSRef artifact.ArtifactRef // expression/1 ∂_φS = a
	DaSRef   artifact.ArtifactRef // expression/1 ∂_aS = φ
	// Test amplitude A(a) = a²+1 with quantum density/pressure (76-77).
	Amplitude artifact.ArtifactRef // expression/1 test |ψ|
	RhoQRef   artifact.ArtifactRef // expression/1 ρQ piece
	PQRef     artifact.ArtifactRef // expression/1 pQ
	// Quantum fluid assembly (78) and comoving velocity.
	TQ00Ref artifact.ArtifactRef // expression/1 T_Q^00 component
	U0Ref   artifact.ArtifactRef // expression/1 U^0 = 1/N component
	// FRW spatial determinant + volume factor (67) via existing paths.
	FRWMetricRef artifact.ArtifactRef // symmetric_bilinear/1 diag(a²,a²,a²), flat chart
	FRWDetRef    artifact.ArtifactRef // expression/1 canonical product
	FRWDerivRef  artifact.ArtifactRef // derivation/1 CANON+SIMPLIFY evidence
	FRWVolRef    artifact.ArtifactRef // volume_density/1 a³-gated factor
	Derivs       []artifact.ArtifactRef
	Symbols      map[string]artifact.ArtifactRef
	Conds        map[string]artifact.ArtifactRef
}

func sym(s physmath.Store, name string) artifact.ArtifactRef {
	ref, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol(name))
	if err != nil {
		panic(err)
	}
	return ref
}

func condNEQ(s physmath.Store, e physmath.Expr) artifact.ArtifactRef {
	body, err := physmath.EncodeRelationBody(physmath.OpNEQ, e, physmath.MustRational(0, 1))
	if err != nil {
		panic(err)
	}
	rel, err := s.Put(physmath.FamilyRelation, body)
	if err != nil {
		panic(err)
	}
	cond, err := s.Put(physmath.FamilyCondition, rel.Encode())
	if err != nil {
		panic(err)
	}
	return cond
}

func condGTE(s physmath.Store, e physmath.Expr) artifact.ArtifactRef {
	body, err := physmath.EncodeRelationBody(physmath.OpGTE, e, physmath.MustRational(0, 1))
	if err != nil {
		panic(err)
	}
	rel, err := s.Put(physmath.FamilyRelation, body)
	if err != nil {
		panic(err)
	}
	cond, err := s.Put(physmath.FamilyCondition, rel.Encode())
	if err != nil {
		panic(err)
	}
	return cond
}

func applyReplay(s physmath.Store, rule string, inputs []artifact.ArtifactRef, conds []artifact.ArtifactRef, params []byte) artifact.ArtifactRef {
	res, err := physmath.Apply(s, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: rule,
		InputRefs: inputs, ConditionRefs: conds, Parameters: params,
	})
	if err != nil {
		panic(err)
	}
	return res.OutputRef
}

func diffParams(fn artifact.ArtifactRef, pid physmath.ParameterID, order uint32) []byte {
	return append(physmath.BoundParameterRef{DeclaringArtifactRef: fn, ParameterID: pid}.Encode(),
		artifact.EncodeUInt32(order)...)
}

// Build constructs the minisuperspace evidence chain. It panics on
// programmer error; replay/validation failures of crafted artifacts surface
// through Verify in protocol tests.
func Build() *Fixture {
	s := physmath.NewMemStore()
	f := &Fixture{Store: s, Symbols: map[string]artifact.ArtifactRef{}, Conds: map[string]artifact.ArtifactRef{}}
	pa := physmath.DeriveParameterID(0)
	pp := physmath.DeriveParameterID(1)
	a := physmath.NewBoundSymbol("a", pa)
	phi := physmath.NewBoundSymbol("phi", pp)
	for _, n := range []string{"N", "Uphi", "kappa2", "xbar", "T00", "rhoQ", "pQ"} {
		f.Symbols[n] = sym(s, n)
	}
	f.Conds["aNonzero"] = condNEQ(s, a)
	f.Conds["phiNonzero"] = condNEQ(s, phi)
	f.Conds["aPos"] = condGTE(s, a)
	f.Conds["nNonzero"] = condNEQ(s, physmath.NewEphemeralSymbol("N"))
	rfn := mintFn(s, []physmath.ParameterID{pa, pp},
		physmath.Add(physmath.Pow(a, physmath.MustRational(2, 1)),
			physmath.Pow(phi, physmath.MustRational(2, 1)), physmath.MustRational(1, 1)))
	f.RFunRef = rfn
	var err error
	f.RRef, err = refOf(s, rfn)
	if err != nil {
		panic(err)
	}
	// ∂²_φR via Order-2 DIFFERENTIATE; replayed.
	d2phi, d2phiDeriv := diffReplay(s, f.RRef, rfn, pp, 2, nil)
	f.D2PhiRef = d2phi
	f.Derivs = append(f.Derivs, d2phiDeriv)
	// (a∂_a)²R by composition: d1=∂_aR; m1=a·d1; d2=∂_a(m1); out=a·d2.
	d1, d1Deriv := diffReplay(s, f.RRef, rfn, pa, 1, nil)
	m1, err := physmath.PutExpression(s, physmath.Mul(a, mustResolve(s, d1)))
	if err != nil {
		panic(err)
	}
	d2, d2Deriv := diffReplay(s, m1, rfn, pa, 1, nil)
	out, outDeriv := diffReplay2(s, d2, rfn, pa, a)
	f.AdaRef = out
	f.Derivs = append(f.Derivs, d1Deriv, d2Deriv, outDeriv)
	// Residual ĤR with k=0, Λ=0, U opaque: -1/2a³·∂²R + κ̄²/2a³·(a∂_a)²R + a³U.
	ainv3 := physmath.Pow(a, physmath.MustRational(-3, 1))
	kap := physmath.NewEphemeralSymbol("kappa2")
	u := physmath.NewEphemeralSymbol("Uphi")
	resid := physmath.Add(
		physmath.Mul(physmath.MustRational(-1, 2), ainv3, mustResolve(s, f.D2PhiRef)),
		physmath.Mul(physmath.Mul(physmath.MustRational(1, 2), kap), ainv3, mustResolve(s, f.AdaRef)),
		physmath.Mul(physmath.Pow(a, physmath.MustRational(3, 1)), u),
	)
	f.Residual, err = physmath.PutExpression(s, resid)
	if err != nil {
		panic(err)
	}
	// Canon replay of the residual construction (no differentiation inside).
	canonRes := applyReplay(s, physmath.RuleExprCanon, []artifact.ArtifactRef{f.Residual}, nil, nil)
	_ = canonRes
	// Test phase S = a·φ: guidance RHS partials (75).
	sfn := mintFn(s, []physmath.ParameterID{pa, pp}, physmath.Mul(a, phi))
	f.SRef, err = refOf(s, sfn)
	if err != nil {
		panic(err)
	}
	dphis, dphisDeriv := diffReplay(s, f.SRef, sfn, pp, 1, nil)
	f.DPhiSRef = dphis
	das, dasDeriv := diffReplay(s, f.SRef, sfn, pa, 1, nil)
	f.DaSRef = das
	f.Derivs = append(f.Derivs, dphisDeriv, dasDeriv)
	// Test amplitude A(a) = a²+1: ρQ/pQ chain (76-77).
	amp := physmath.Add(physmath.Pow(a, physmath.MustRational(2, 1)), physmath.MustRational(1, 1))
	ampFn := mintFn(s, []physmath.ParameterID{pa}, amp)
	ampRef, err := refOf(s, ampFn)
	if err != nil {
		panic(err)
	}
	f.Amplitude = ampRef
	conds := []artifact.ArtifactRef{f.Conds["aNonzero"]}
	ampCond := condNEQ(s, amp)
	conds = append(conds, ampCond)
	// ∂²_φA = 0 (φ-independent); (a∂_a)²A = 4a² by the same composition.
	ad1, ad1Deriv := diffReplay(s, ampRef, ampFn, pa, 1, nil)
	am1, err := physmath.PutExpression(s, physmath.Mul(a, mustResolve(s, ad1)))
	if err != nil {
		panic(err)
	}
	ad2, ad2Deriv := diffReplay(s, am1, ampFn, pa, 1, nil)
	aout, aoutDeriv := diffReplay2(s, ad2, ampFn, pa, a)
	f.Derivs = append(f.Derivs, ad1Deriv, ad2Deriv, aoutDeriv)
	// Q = κ̄²/2a⁶ · 4a²/A(a): Mul(κ̄²/2, a^-6, 4a²-expr, A^-1).
	qexpr := physmath.Mul(
		physmath.Mul(physmath.MustRational(1, 2), kap),
		physmath.Pow(a, physmath.MustRational(-6, 1)),
		mustResolve(s, aout),
		physmath.Pow(mustResolve(s, ampRef), physmath.MustRational(-1, 1)),
	)
	qref, err := physmath.PutExpression(s, qexpr)
	if err != nil {
		panic(err)
	}
	// p_Q = -1/3a² · ∂_aQ via extended DIFFERENTIATE (quotient-capable).
	dq, dqDeriv := diffReplay(s, qref, ampFn, pa, 1, conds)
	f.Derivs = append(f.Derivs, dqDeriv)
	pq := physmath.Mul(physmath.MustRational(-1, 3),
		physmath.Pow(a, physmath.MustRational(-2, 1)), mustResolve(s, dq))
	f.PQRef, err = physmath.PutExpression(s, pq)
	if err != nil {
		panic(err)
	}
	// ρQ = Q/a³: represented alongside (same Q evidence).
	rhoq := physmath.Mul(physmath.Pow(a, physmath.MustRational(-3, 1)), mustResolve(s, qref))
	f.RhoQRef, err = physmath.PutExpression(s, rhoq)
	if err != nil {
		panic(err)
	}
	// Comoving U^0 = 1/N via the Kronecker selector (δ^μ_0, dim 4, pos 0).
	comoving := spaceCollection(s, "comoving_space")
	muIdx := spaceIndex(s, "mu", physmath.VarianceContravariant, comoving)
	sel, err := physmath.KroneckerSelector(s, 4, 0, muIdx)
	if err != nil {
		panic(err)
	}
	if err := physmath.ValidateKroneckerSelector(s, sel, 0); err != nil {
		panic(err)
	}
	u0 := physmath.Mul(mustResolve(s, mustComp(s, sel, 0)),
		physmath.Pow(physmath.NewEphemeralSymbol("N"), physmath.MustRational(-1, 1)))
	f.U0Ref, err = physmath.PutExpression(s, u0)
	if err != nil {
		panic(err)
	}
	// T_Q^00 = (ρQ+pQ)·U^0·U^0 − pQ·g^00 with g^00 = 1/N² represented.
	g00 := physmath.Pow(physmath.NewEphemeralSymbol("N"), physmath.MustRational(-2, 1))
	tq00 := physmath.Add(
		physmath.Mul(physmath.Add(mustResolve(s, f.RhoQRef), mustResolve(s, f.PQRef)),
			mustResolve(s, f.U0Ref), mustResolve(s, f.U0Ref)),
		physmath.Neg(physmath.Mul(mustResolve(s, f.PQRef), g00)),
	)
	f.TQ00Ref, err = physmath.PutExpression(s, tq00)
	if err != nil {
		panic(err)
	}
	// FRW spatial diag(a²,a²,a²), flat chart: determinant + a>0 volume.
	a2 := physmath.Pow(a, physmath.MustRational(2, 1))
	frw := diagTensor(s, []physmath.Expr{a2, a2, a2})
	f.FRWMetricRef = frw
	det, deriv, err := physmath.DiagonalDeterminant(s, frw)
	if err != nil {
		panic(err)
	}
	f.FRWDetRef = det
	f.FRWDerivRef = mustMintDerivation(s, deriv)
	absa3 := sym(s, "abs_a_cubed")
	redRef := mustStorePow(s, a, 3)
	red := redRef
	vol := physmath.VolumeDensity{MetricRef: frw, DeterminantRef: det,
		DeterminantDerivationRef: f.FRWDerivRef,
		SignatureClass:           physmath.SignaturePositiveDefinite, FactorRef: absa3,
		ReductionRef:           &red,
		ReductionConditionRefs: []artifact.ArtifactRef{f.Conds["aPos"]},
		DomainAssumptionRefs:   []artifact.ArtifactRef{f.Conds["aPos"]}}
	if err := vol.Validate(); err != nil {
		panic(err)
	}
	if err := physmath.CheckVolumeDensity(s, vol); err != nil {
		panic(err)
	}
	payload, err := vol.Encode()
	if err != nil {
		panic(err)
	}
	f.FRWVolRef, err = s.Put(physmath.FamilyVolumeDensity, payload)
	if err != nil {
		panic(err)
	}
	return f
}

// Verify replays every stored evidence derivation. It establishes that
// the recorded computations reproduce from their premises; it does not
// solve equations, evolve time, or verify the full Wheeler-DeWitt theory.
func (f *Fixture) Verify() error {
	for i, dref := range f.Derivs {
		_, payload, err := f.Store.Lookup(dref)
		if err != nil {
			return fmt.Errorf("evidence derivation %d unresolvable: %w", i, err)
		}
		d, err := physmath.DecodeDerivation(payload)
		if err != nil {
			return fmt.Errorf("evidence derivation %d undecodable: %w", i, err)
		}
		if err := physmath.Replay(f.Store, d, nil); err != nil {
			return fmt.Errorf("evidence derivation %d does not replay: %w", i, err)
		}
	}
	return nil
}

// mintFn mints a function/1 artifact with a validated body.
func mintFn(s physmath.Store, ids []physmath.ParameterID, body physmath.Expr) artifact.ArtifactRef {
	bodyRef, err := physmath.PutExpression(s, body)
	if err != nil {
		panic(err)
	}
	fn := physmath.Function{ParameterIDs: ids, BodyRef: bodyRef}
	if err := physmath.ValidateFunctionBody(fn, body); err != nil {
		panic(err)
	}
	payload, err := fn.Encode()
	if err != nil {
		panic(err)
	}
	ref, err := s.Put(physmath.FamilyFunction, payload)
	if err != nil {
		panic(err)
	}
	return ref
}

// refOf resolves a function artifact to its body expression reference.
func refOf(s physmath.Store, fnRef artifact.ArtifactRef) (artifact.ArtifactRef, error) {
	fam, payload, err := s.Lookup(fnRef)
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	if fam != physmath.FamilyFunction {
		return artifact.ArtifactRef{}, fmt.Errorf("not function/1")
	}
	fn, err := physmath.DecodeFunction(artifact.NewDecoder(payload))
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return fn.BodyRef, nil
}

func mustResolve(s physmath.Store, ref artifact.ArtifactRef) physmath.Expr {
	fam, payload, err := s.Lookup(ref)
	if err != nil {
		panic(err)
	}
	if fam != physmath.FamilyExpression {
		panic("not expression/1")
	}
	e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
	if err != nil {
		panic(err)
	}
	return e
}

// diffReplay runs DIFFERENTIATE-001 with replay evidence, returning the
// output ref plus a minted derivation/1 ref for protocol verification.
func diffReplay(s physmath.Store, in, fn artifact.ArtifactRef, pid physmath.ParameterID, order uint32, conds []artifact.ArtifactRef) (artifact.ArtifactRef, artifact.ArtifactRef) {
	res, err := physmath.Apply(s, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
		InputRefs: []artifact.ArtifactRef{in, fn}, ConditionRefs: conds,
		Parameters: diffParams(fn, pid, order),
	})
	if err != nil {
		panic(err)
	}
	d := physmath.Derivation{
		PremiseRefs: []artifact.ArtifactRef{in, fn},
		Steps: []physmath.DerivationStep{{
			Ordinal: 0, ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleDifferentiate,
			InputRefs: []artifact.ArtifactRef{in, fn}, OutputRef: res.OutputRef,
			ConditionRefs: res.ConditionRefs, Parameters: diffParams(fn, pid, order),
		}},
		ResultRefs: []artifact.ArtifactRef{res.OutputRef},
	}
	if err := physmath.Replay(s, d, nil); err != nil {
		panic(err)
	}
	return res.OutputRef, mustMintDerivation(s, d)
}

// diffReplay2 multiplies a differentiated result by a and returns the
// product ref plus its CANON replay derivation (composition helper for
// (a∂_a)² chains).
func diffReplay2(s physmath.Store, dref, fn artifact.ArtifactRef, pid physmath.ParameterID, avar physmath.Expr) (artifact.ArtifactRef, artifact.ArtifactRef) {
	_ = pid
	_ = fn
	mref, err := physmath.PutExpression(s, physmath.Mul(avar, mustResolve(s, dref)))
	if err != nil {
		panic(err)
	}
	res, err := physmath.Apply(s, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprCanon,
		InputRefs: []artifact.ArtifactRef{mref},
	})
	if err != nil {
		panic(err)
	}
	d := physmath.Derivation{
		PremiseRefs: []artifact.ArtifactRef{mref},
		Steps: []physmath.DerivationStep{{
			Ordinal: 0, ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{mref}, OutputRef: res.OutputRef,
		}},
		ResultRefs: []artifact.ArtifactRef{res.OutputRef},
	}
	if err := physmath.Replay(s, d, nil); err != nil {
		panic(err)
	}
	return res.OutputRef, mustMintDerivation(s, d)
}

func mustMintDerivation(s physmath.Store, d physmath.Derivation) artifact.ArtifactRef {
	raw, err := d.Encode()
	if err != nil {
		panic(err)
	}
	ref, err := s.Put(physmath.FamilyDerivation, raw)
	if err != nil {
		panic(err)
	}
	return ref
}

func spaceCollection(s physmath.Store, name string) artifact.ArtifactRef {
	desig, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol(name))
	if err != nil {
		panic(err)
	}
	raw, err := physmath.Collection{ElementRefs: []artifact.ArtifactRef{desig}}.Encode()
	if err != nil {
		panic(err)
	}
	ref, err := s.Put(physmath.FamilyCollection, raw)
	if err != nil {
		panic(err)
	}
	return ref
}

func spaceIndex(s physmath.Store, label string, v physmath.IndexVariance, space artifact.ArtifactRef) artifact.ArtifactRef {
	p, err := physmath.Index{Label: label, Variance: v, Space: &space}.Encode()
	if err != nil {
		panic(err)
	}
	ref, err := s.Put(physmath.FamilyIndex, p)
	if err != nil {
		panic(err)
	}
	return ref
}

func mustComp(s physmath.Store, tensor artifact.ArtifactRef, i int) artifact.ArtifactRef {
	fam, payload, err := s.Lookup(tensor)
	if err != nil {
		panic(err)
	}
	if fam != physmath.FamilyTensor {
		panic("not tensor/1")
	}
	tn, err := physmath.DecodeTensor(payload)
	if err != nil {
		panic(err)
	}
	return tn.ComponentRefs[i]
}

func mustStorePow(s physmath.Store, base physmath.Expr, n int64) artifact.ArtifactRef {
	ref, err := physmath.PutExpression(s, physmath.Pow(base, physmath.MustRational(n, 1)))
	if err != nil {
		panic(err)
	}
	return ref
}

// diagTensor mints a diagonal symmetric_bilinear/1 metric from diagonal
// component expressions (flat-chart representation only).
func diagTensor(s physmath.Store, diags []physmath.Expr) artifact.ArtifactRef {
	n := len(diags)
	zero, err := physmath.PutExpression(s, physmath.MustRational(0, 1))
	if err != nil {
		panic(err)
	}
	refs := make([]artifact.ArtifactRef, n*n)
	for i := range refs {
		refs[i] = zero
	}
	for i, e := range diags {
		r, err := physmath.PutExpression(s, e)
		if err != nil {
			panic(err)
		}
		refs[i*n+i] = r
	}
	m := physmath.SymmetricBilinear{Dimension: uint16(n), ComponentRefs: refs}
	if err := m.Validate(); err != nil {
		panic(err)
	}
	ref, err := s.Put(physmath.FamilySymmetricBilinear, m.Encode())
	if err != nil {
		panic(err)
	}
	return ref
}
