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
	SRef        artifact.ArtifactRef
	DPhiSRef    artifact.ArtifactRef // expression/1 ∂_φS = a
	DaSRef      artifact.ArtifactRef // expression/1 ∂_aS = φ
	GuidePhiRef artifact.ArtifactRef // expression/1 φ̇ RHS = N/a³·∂_φS (constructed)
	GuideARef   artifact.ArtifactRef // expression/1 ȧ RHS = -κ̄²·N/a·∂_aS (constructed)
	// Test amplitude A(a) = a²+1 with quantum density/pressure (76-77).
	// Dependency order: Amplitude → W → RhoQRef → QRef → dQ → PQRef.
	Amplitude   artifact.ArtifactRef // expression/1 test |psi|
	AmpFunRef   artifact.ArtifactRef // function/1 declaring the amplitude parameter |ψ|
	WSimpRef    artifact.ArtifactRef // expression/1 simplified factored W
	RhoQRef     artifact.ArtifactRef // expression/1 ρQ = (76) density
	QRef        artifact.ArtifactRef // expression/1 Q = a³ρQ quantum potential
	QRelRef     artifact.ArtifactRef // relation/1 EQ(Q, a³ρQ) defining relation
	KappaFunRef artifact.ArtifactRef // function/1 declaring the κ̄² parameter
	PQRef       artifact.ArtifactRef // expression/1 pQ = -1/3a²·∂_aQ
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
	// FRW covariant fragment: chart-bound metric connection and covector
	// derivative evidence, verified by Verify() (no side stores).
	FRWChartRef     artifact.ArtifactRef   // coordinate_system/1 Cartesian chart
	FRWCovMetricRef artifact.ArtifactRef   // symmetric_bilinear/1 diag(a3²×3), covariant-fragment scale factor
	FRWDeclFnRef    artifact.ArtifactRef   // function/1 declaring chart coordinates
	FRWMapping      physmath.CoordParamMap // ordered coordinate-to-parameter binding
	FRWConnRef      artifact.ArtifactRef   // affine_connection/1 Levi-Civita of FRWMetricRef
	FRWConnDerivRef artifact.ArtifactRef   // derivation/1 connection evidence
	FRWCovFieldRef  artifact.ArtifactRef   // tensor_field/1 test covector
	FRWCovResultRef artifact.ArtifactRef   // tensor_field/1 covariant-derivative result
	FRWCovEvalRef   artifact.ArtifactRef   // covariant_evaluation/1 certificate
	FRWCovDerivRef  artifact.ArtifactRef   // derivation/1 evaluation evidence
	// Stable shared identities for repeated quantities.
	NRef    artifact.ArtifactRef // expression/1 lapse N, single identity
	UPhiRef artifact.ArtifactRef // expression/1 potential U(phi), single identity
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
	pk := physmath.DeriveParameterID(2)
	a := physmath.NewBoundSymbol("a", pa)
	phi := physmath.NewBoundSymbol("phi", pp)
	// κ̄² is a declared model parameter (own PID, disjoint from pa/pp), so
	// it differentiates to 0 as an independent quantity and remains
	// substitutable via EXPR-SUBSTITUTE-001, unlike an ephemeral symbol.
	kap := physmath.NewBoundSymbol("kappa2", pk)
	f.KappaFunRef = mintFn(s, []physmath.ParameterID{pa, pp, pk}, kap)
	// Stable shared identities: one ref per repeated quantity, reused at
	// every site below. The map keeps only genuinely shared refs.
	f.NRef = sym(s, "N")
	f.Symbols["N"] = f.NRef
	f.UPhiRef = sym(s, "Uphi")
	f.Symbols["Uphi"] = f.UPhiRef
	f.Conds["aNonzero"] = condNEQ(s, a)
	// No phiNonzero condition is minted: post-P1 polynomial differentiation
	// manufactures no Pow(phi,0) node, so the WDW pieces hold on the full
	// declared domain including phi = 0.
	f.Conds["aPos"] = condGTE(s, a)
	f.Conds["nNonzero"] = condNEQ(s, mustResolve(s, f.NRef))
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
	u := mustResolve(s, f.UPhiRef)
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
	// Full RHS assembly per paper (75) from the executed partials. The
	// time-derivative left-hand sides stay represented relations; only the
	// right-hand sides are constructed here, from replayed partials.
	neph := mustResolve(s, f.NRef)
	f.GuidePhiRef, err = physmath.PutExpression(s, physmath.Mul(neph,
		physmath.Pow(a, physmath.MustRational(-3, 1)), mustResolve(s, f.DPhiSRef)))
	if err != nil {
		panic(err)
	}
	f.GuideARef, err = physmath.PutExpression(s, physmath.Mul(
		physmath.Neg(physmath.MustRational(1, 1)), kap,
		neph, physmath.Pow(a, physmath.MustRational(-1, 1)), mustResolve(s, f.DaSRef)))
	if err != nil {
		panic(err)
	}
	// Test amplitude A(a) = a²+1: ρQ/pQ chain (76-77).
	amp := physmath.Add(physmath.Pow(a, physmath.MustRational(2, 1)), physmath.MustRational(1, 1))
	ampFn := mintFn(s, []physmath.ParameterID{pa}, amp)
	f.AmpFunRef = ampFn
	ampRef, err := refOf(s, ampFn)
	if err != nil {
		panic(err)
	}
	f.Amplitude = ampRef
	conds := []artifact.ArtifactRef{f.Conds["aNonzero"]}
	ampCond := condNEQ(s, amp)
	f.Conds["ampNonzero"] = ampCond
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
	// Stage the composed result through SIMPLIFY (replayed): downstream
	// assemblies consume the simplified factored form, keeping goldens
	// hand-traceable. Post-P1 no Pow(·,0) node can arise here, so no
	// conditions are needed for this step.
	simpW, simpWDeriv := simpReplay(s, aout, nil)
	f.WSimpRef = simpW
	f.Derivs = append(f.Derivs, simpWDeriv)
	// ρQ from paper (76): for φ-independent A the first term vanishes, so
	// ρQ = κ̄²/2 · a^-6 · W · A^-1 with W = (a∂_a)²A. This record is the
	// density, not the potential: see QRef below.
	rhoqExpr := physmath.Mul(
		physmath.Mul(physmath.MustRational(1, 2), kap),
		physmath.Pow(a, physmath.MustRational(-6, 1)),
		mustResolve(s, f.WSimpRef),
		physmath.Pow(mustResolve(s, ampRef), physmath.MustRational(-1, 1)),
	)
	f.RhoQRef, err = physmath.PutExpression(s, rhoqExpr)
	if err != nil {
		panic(err)
	}
	// Q = a³ρQ as a first-class artifact plus its defining relation.
	qexpr := physmath.Mul(
		physmath.Pow(a, physmath.MustRational(3, 1)),
		mustResolve(s, f.RhoQRef),
	)
	f.QRef, err = physmath.PutExpression(s, qexpr)
	if err != nil {
		panic(err)
	}
	qrel, err := physmath.EncodeRelationBody(physmath.OpEQ, mustResolve(s, f.QRef), qexpr)
	if err != nil {
		panic(err)
	}
	f.QRelRef, err = s.Put(physmath.FamilyRelation, qrel)
	if err != nil {
		panic(err)
	}
	// p_Q = -1/3a² · ∂_aQ via extended DIFFERENTIATE (quotient-capable).
	// The derivative acts on Q, never on ρQ directly.
	dq, dqDeriv := diffReplay(s, f.QRef, ampFn, pa, 1, conds)
	f.Derivs = append(f.Derivs, dqDeriv)
	pq := physmath.Mul(physmath.MustRational(-1, 3),
		physmath.Pow(a, physmath.MustRational(-2, 1)), mustResolve(s, dq))
	f.PQRef, err = physmath.PutExpression(s, pq)
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
		physmath.Pow(mustResolve(s, f.NRef), physmath.MustRational(-1, 1)))
	f.U0Ref, err = physmath.PutExpression(s, u0)
	if err != nil {
		panic(err)
	}
	// T_Q^00 = (ρQ+pQ)·U^0·U^0 − pQ·g^00 with g^00 = 1/N² represented.
	g00 := physmath.Pow(mustResolve(s, f.NRef), physmath.MustRational(-2, 1))
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
	f.buildFRWCovariant(s)
	return f
}

// buildFRWCovariant constructs the FRW covariant-derivative fragment inside
// the fixture store: Cartesian chart, general inverse, Levi-Civita
// connection, test covector, evaluated derivative, and certificate. All
// derivation evidence is minted into f.Derivs for Verify(). The scale
// factor keeps its fixture-wide bound identity (parameter 0), disjoint
// from the chart coordinates, so spatial partials vanish by the FRW
// symmetries and the connection coefficients compute to zero.
func (f *Fixture) buildFRWCovariant(s physmath.Store) {
	// Parameter-identity discipline: chart coordinates must occupy ordinal
	// PIDs 0-2 of their declaring function (spec §8.3), which the fixture's
	// scale factor `a` (PID 0) already uses. A shared PID would make
	// differentiation confuse the scale factor with a coordinate, so the
	// covariant fragment uses a distinct scale-factor identity `a` at
	// PID 3: same display label, different mathematical variable, recorded
	// here and in the manifest. The determinant/volume path keeps the
	// original FRWMetricRef untouched.
	px := physmath.DeriveParameterID(0)
	py := physmath.DeriveParameterID(1)
	pz := physmath.DeriveParameterID(2)
	ps := physmath.DeriveParameterID(3)
	x := physmath.NewBoundSymbol("x", px)
	y := physmath.NewBoundSymbol("y", py)
	z := physmath.NewBoundSymbol("z", pz)
	ascale := physmath.NewBoundSymbol("a", ps)
	declFn := mintFn(s, []physmath.ParameterID{px, py, pz}, x)
	f.FRWDeclFnRef = declFn
	desig, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol("frw_cartesian_space"))
	if err != nil {
		panic(err)
	}
	spaceRaw, err := physmath.Collection{ElementRefs: []artifact.ArtifactRef{desig}}.Encode()
	if err != nil {
		panic(err)
	}
	space, err := s.Put(physmath.FamilyCollection, spaceRaw)
	if err != nil {
		panic(err)
	}
	coordRef := func(e physmath.Expr) artifact.ArtifactRef {
		r, err := physmath.PutExpression(s, e)
		if err != nil {
			panic(err)
		}
		return r
	}
	chart := physmath.CoordinateSystem{Dimension: 3,
		CoordinateRefs: []artifact.ArtifactRef{coordRef(x), coordRef(y), coordRef(z)},
		SpaceRef:       space}
	if err := chart.Validate(); err != nil {
		panic(err)
	}
	f.FRWChartRef, err = s.Put(physmath.FamilyCoordinateSystem, chart.Encode())
	if err != nil {
		panic(err)
	}
	f.FRWMapping = physmath.CoordParamMap{Entries: []physmath.BoundParameterRef{
		{DeclaringArtifactRef: declFn, ParameterID: px},
		{DeclaringArtifactRef: declFn, ParameterID: py},
		{DeclaringArtifactRef: declFn, ParameterID: pz},
	}}
	// Covariant-fragment metric over the disjoint scale identity.
	a2cov := physmath.Pow(ascale, physmath.MustRational(2, 1))
	f.FRWCovMetricRef = diagTensor(s, []physmath.Expr{a2cov, a2cov, a2cov})
	// General inverse of the fragment metric (never a diagonal shortcut).
	detRef, _, err := physmath.GeneralDeterminant(s, f.FRWCovMetricRef)
	if err != nil {
		panic(err)
	}
	detExpr, err := resolveExprRef(s, f, detRef)
	if err != nil {
		panic(err)
	}
	cond := condNEQExpr(s, detExpr)
	connRef, connDeriv, err := physmath.LeviCivita(s, f.FRWCovMetricRef, f.FRWChartRef,
		f.FRWMapping, declFn, []artifact.ArtifactRef{cond})
	if err != nil {
		panic(err)
	}
	f.FRWConnRef = connRef
	f.FRWConnDerivRef = mustMintDerivation(s, connDeriv)
	f.Derivs = append(f.Derivs, f.FRWConnDerivRef)
	// Test covector omega = [x, y, z] over the chart space.
	mkIdx := func(label string) artifact.ArtifactRef {
		p, err := physmath.Index{Label: label, Variance: physmath.VarianceCovariant, Space: &space}.Encode()
		if err != nil {
			panic(err)
		}
		r, err := s.Put(physmath.FamilyIndex, p)
		if err != nil {
			panic(err)
		}
		return r
	}
	crefs := []artifact.ArtifactRef{
		mustStoreExprOf(s, x), mustStoreExprOf(s, y), mustStoreExprOf(s, z),
	}
	tn := physmath.Tensor{Rank: 1, Dimension: 3,
		IndexRefs: []artifact.ArtifactRef{mkIdx("nu")}, ComponentRefs: crefs}
	if err := tn.Validate(); err != nil {
		panic(err)
	}
	omega, err := s.Put(physmath.FamilyTensor, tn.Encode())
	if err != nil {
		panic(err)
	}
	omegaField, err := s.Put(physmath.FamilyTensorField,
		physmath.TensorField{TensorRef: omega, CoordinateSystemRef: f.FRWChartRef}.Encode())
	if err != nil {
		panic(err)
	}
	f.FRWCovFieldRef = omegaField
	wres, wderiv, err := physmath.EvaluateCovariantDerivative(s, omegaField, connRef,
		f.FRWMapping, declFn, nil)
	if err != nil {
		panic(err)
	}
	f.FRWCovResultRef = wres
	f.FRWCovDerivRef = mustMintDerivation(s, wderiv)
	f.Derivs = append(f.Derivs, f.FRWCovDerivRef)
	rec := physmath.CovariantEvaluation{FieldRef: omegaField, ConnectionRef: connRef,
		ResultRef: wres, DerivationRef: f.FRWCovDerivRef}
	if err := rec.Validate(); err != nil {
		panic(err)
	}
	if err := physmath.CheckCovariantEvaluation(s, rec, f.FRWMapping, declFn); err != nil {
		panic(err)
	}
	f.FRWCovEvalRef, err = s.Put(physmath.FamilyCovariantEvaluation, rec.Encode())
	if err != nil {
		panic(err)
	}
}

// condNEQExpr mints an exact NEQ(e,0) condition/1 artifact.
func condNEQExpr(s physmath.Store, e physmath.Expr) artifact.ArtifactRef {
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

// resolveExprRef resolves an expression/1 reference to its value.
func resolveExprRef(s physmath.Store, f *Fixture, ref artifact.ArtifactRef) (physmath.Expr, error) {
	_ = f
	fam, payload, err := s.Lookup(ref)
	if err != nil {
		return physmath.Expr{}, err
	}
	if fam != physmath.FamilyExpression {
		return physmath.Expr{}, fmt.Errorf("not expression/1")
	}
	return physmath.DecodeExpr(artifact.NewDecoder(payload))
}

// mustStoreExprOf mints an expression/1 artifact, panicking on error.
func mustStoreExprOf(s physmath.Store, e physmath.Expr) artifact.ArtifactRef {
	ref, err := physmath.PutExpression(s, e)
	if err != nil {
		panic(err)
	}
	return ref
}

// Verify replays every stored evidence derivation and checks the explicit
// Q/density relation and the FRW covariant record. It establishes that the
// recorded computations reproduce from their premises; it does not solve
// equations, evolve time, or verify the full Wheeler-DeWitt theory.
// Representations without derivation evidence (symbols, assumptions) are
// checked as constructions, never described as replayed.
func (f *Fixture) Verify() error {
	for i, dref := range f.Derivs {
		if err := replayDerivation(f.Store, i, dref); err != nil {
			return err
		}
	}
	// The FRW determinant derivation is stored outside Derivs; it is part
	// of the declared evidence and must replay as well.
	if err := replayDerivation(f.Store, -1, f.FRWDerivRef); err != nil {
		return fmt.Errorf("FRW determinant evidence: %w", err)
	}
	// Q/density relation: decode the STORED equation and verify its
	// content, with `a` resolved through the amplitude declaring function.
	if err := checkQRelation(f.Store, f.QRelRef, f.QRef, f.RhoQRef, f.AmpFunRef); err != nil {
		return err
	}
	// FRW covariant fragment: the stored evaluation record must resolve,
	// replay through its derivation, and match recomputation; the result
	// tensor must carry the chart-bound structure and Kronecker values.
	if err := f.verifyCovariant(); err != nil {
		return err
	}
	return nil
}

// verifyCovariant validates the fixture-integrated covariant fragment:
// record resolution, store-aware check (replay + recomputation), result
// rank/component count, derivative-index-first ordering with exact index
// identities, both slots covariant over the chart space, and the
// hand-derived Kronecker component values.
func (f *Fixture) verifyCovariant() error {
	fam, payload, err := f.Store.Lookup(f.FRWCovEvalRef)
	if err != nil {
		return fmt.Errorf("covariant record unresolvable: %w", err)
	}
	if fam != physmath.FamilyCovariantEvaluation {
		return fmt.Errorf("covariant record is not covariant_evaluation/1")
	}
	rec, err := physmath.DecodeCovariantEvaluation(payload)
	if err != nil {
		return fmt.Errorf("covariant record undecodable: %w", err)
	}
	if rec.FieldRef != f.FRWCovFieldRef {
		return fmt.Errorf("covariant record field is not the fixture covector")
	}
	if rec.ConnectionRef != f.FRWConnRef {
		return fmt.Errorf("covariant record connection is not the fixture connection")
	}
	if rec.ResultRef != f.FRWCovResultRef {
		return fmt.Errorf("covariant record result is not the fixture result")
	}
	if rec.DerivationRef != f.FRWCovDerivRef {
		return fmt.Errorf("covariant record derivation is not the fixture derivation")
	}
	if err := physmath.CheckCovariantEvaluation(f.Store, rec, f.FRWMapping, f.FRWDeclFnRef); err != nil {
		return fmt.Errorf("covariant record invalid: %w", err)
	}
	rfam, rpayload, err := f.Store.Lookup(f.FRWCovResultRef)
	if err != nil {
		return fmt.Errorf("covariant result unresolvable: %w", err)
	}
	if rfam != physmath.FamilyTensorField {
		return fmt.Errorf("covariant result is not tensor_field/1")
	}
	rf, err := physmath.DecodeTensorField(rpayload)
	if err != nil {
		return fmt.Errorf("covariant result undecodable: %w", err)
	}
	if rf.CoordinateSystemRef != f.FRWChartRef {
		return fmt.Errorf("covariant result chart not preserved")
	}
	tfam, tpayload, err := f.Store.Lookup(rf.TensorRef)
	if err != nil {
		return fmt.Errorf("covariant result tensor unresolvable: %w", err)
	}
	if tfam != physmath.FamilyTensor {
		return fmt.Errorf("covariant result tensor is not tensor/1")
	}
	tn, err := physmath.DecodeTensor(tpayload)
	if err != nil {
		return fmt.Errorf("covariant result tensor undecodable: %w", err)
	}
	if tn.Rank != 2 || len(tn.ComponentRefs) != 9 {
		return fmt.Errorf("covariant result shape wrong")
	}
	if len(tn.IndexRefs) != 2 {
		return fmt.Errorf("covariant result index count wrong")
	}
	for i, r := range tn.IndexRefs {
		ifam, ipayload, err := f.Store.Lookup(r)
		if err != nil {
			return fmt.Errorf("covariant result index %d unresolvable: %w", i, err)
		}
		if ifam != physmath.FamilyIndex {
			return fmt.Errorf("covariant result index %d is not index/1", i)
		}
		idx, err := physmath.DecodeIndex(ipayload)
		if err != nil {
			return fmt.Errorf("covariant result index %d undecodable: %w", i, err)
		}
		if idx.Variance != physmath.VarianceCovariant {
			return fmt.Errorf("covariant result index %d is not covariant", i)
		}
		if idx.Space == nil {
			return fmt.Errorf("covariant result index %d lacks space", i)
		}
		cfam, cpayload, err := f.Store.Lookup(f.FRWChartRef)
		if err != nil {
			return fmt.Errorf("covariant chart unresolvable: %w", err)
		}
		if cfam != physmath.FamilyCoordinateSystem {
			return fmt.Errorf("covariant chart is not coordinate_system/1")
		}
		chart, err := physmath.DecodeCoordinateSystem(cpayload)
		if err != nil {
			return fmt.Errorf("covariant chart undecodable: %w", err)
		}
		if *idx.Space != chart.SpaceRef {
			return fmt.Errorf("covariant result index %d space differs from chart space", i)
		}
	}
	// Derivative index first, original covector index second: exact ordered
	// identity against the input field's slot.
	wfield, wfpayload, err := f.Store.Lookup(f.FRWCovFieldRef)
	if err != nil {
		return fmt.Errorf("covariant input field unresolvable: %w", err)
	}
	if wfield != physmath.FamilyTensorField {
		return fmt.Errorf("covariant input field is not tensor_field/1")
	}
	wf, err := physmath.DecodeTensorField(wfpayload)
	if err != nil {
		return fmt.Errorf("covariant input field undecodable: %w", err)
	}
	wtfam, wtpayload, err := f.Store.Lookup(wf.TensorRef)
	if err != nil {
		return fmt.Errorf("covariant input tensor unresolvable: %w", err)
	}
	if wtfam != physmath.FamilyTensor {
		return fmt.Errorf("covariant input tensor is not tensor/1")
	}
	wt, err := physmath.DecodeTensor(wtpayload)
	if err != nil {
		return fmt.Errorf("covariant input tensor undecodable: %w", err)
	}
	if len(wt.IndexRefs) != 1 || tn.IndexRefs[1] != wt.IndexRefs[0] {
		return fmt.Errorf("covariant free index not preserved in order")
	}
	one, err := physmath.MustRational(1, 1).Encode()
	if err != nil {
		return err
	}
	zero, err := physmath.MustRational(0, 1).Encode()
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			want := zero
			if i == j {
				want = one
			}
			_, cp, err := f.Store.Lookup(tn.ComponentRefs[i*3+j])
			if err != nil {
				return fmt.Errorf("covariant result component (%d,%d) unresolvable: %w", i, j, err)
			}
			if string(cp) != string(want) {
				return fmt.Errorf("covariant result component (%d,%d) is not Kronecker", i, j)
			}
		}
	}
	return nil
}

func replayDerivation(s physmath.Store, i int, dref artifact.ArtifactRef) error {
	_, payload, err := s.Lookup(dref)
	if err != nil {
		return fmt.Errorf("evidence derivation %d unresolvable: %w", i, err)
	}
	d, err := physmath.DecodeDerivation(payload)
	if err != nil {
		return fmt.Errorf("evidence derivation %d undecodable: %w", i, err)
	}
	if err := physmath.Replay(s, d, nil); err != nil {
		return fmt.Errorf("evidence derivation %d does not replay: %w", i, err)
	}
	return nil
}

// checkQRelation verifies the actual equation referenced by relRef: it must
// decode as an EQ relation whose left side is byte-identical to the stored
// Q expression and whose right side canonically equals Mul(a³, ρQ), with `a`
// resolved as parameter 0 of the stored amplitude declaring function (never
// a magic ordinal or display-name match).
func checkQRelation(s physmath.Store, relRef, qref, rhoref, ampFnRef artifact.ArtifactRef) error {
	relfam, relpayload, err := s.Lookup(relRef)
	if err != nil {
		return fmt.Errorf("Q relation unresolvable: %w", err)
	}
	if relfam != physmath.FamilyRelation {
		return fmt.Errorf("Q relation is not relation/1")
	}
	rd := artifact.NewDecoder(relpayload)
	opStr, err := rd.String()
	if err != nil {
		return fmt.Errorf("Q relation missing operator: %w", err)
	}
	if opStr != string(physmath.OpEQ) {
		return fmt.Errorf("Q relation is not an equality")
	}
	relLeft, err := physmath.DecodeExpr(rd)
	if err != nil {
		return fmt.Errorf("Q relation left undecodable: %w", err)
	}
	relRight, err := physmath.DecodeExpr(rd)
	if err != nil {
		return fmt.Errorf("Q relation right undecodable: %w", err)
	}
	if !rd.Exhausted() {
		return fmt.Errorf("trailing bytes in Q relation")
	}
	qf, qpayload, err := s.Lookup(qref)
	if err != nil {
		return fmt.Errorf("Q unresolvable: %w", err)
	}
	if qf != physmath.FamilyExpression {
		return fmt.Errorf("Q is not expression/1")
	}
	lb, err := relLeft.Encode()
	if err != nil {
		return fmt.Errorf("Q relation left unencodable: %w", err)
	}
	if string(lb) != string(qpayload) {
		return fmt.Errorf("Q relation left side is not the stored Q")
	}
	rf, rpayload, err := s.Lookup(rhoref)
	if err != nil {
		return fmt.Errorf("density unresolvable: %w", err)
	}
	if rf != physmath.FamilyExpression {
		return fmt.Errorf("density is not expression/1")
	}
	rho, err := physmath.DecodeExpr(artifact.NewDecoder(rpayload))
	if err != nil {
		return fmt.Errorf("density undecodable: %w", err)
	}
	a, err := amplitudeScaleFactor(s, ampFnRef)
	if err != nil {
		return err
	}
	want, err := physmath.Canonicalize(physmath.Mul(physmath.Pow(a, physmath.MustRational(3, 1)), rho))
	if err != nil {
		return fmt.Errorf("Q relation expectation unbuildable: %w", err)
	}
	rb, err := physmath.Canonicalize(relRight)
	if err != nil {
		return fmt.Errorf("Q relation right uncanonicalizable: %w", err)
	}
	if !rb.Equal(want) {
		return fmt.Errorf("Q relation right side is not Mul(a³, density)")
	}
	return nil
}

// amplitudeScaleFactor resolves the scale-factor symbol through the stored
// amplitude declaring function: parameter 0 of that function is the `a`
// identity used across the fixture. No ordinal literal, no label match.
func amplitudeScaleFactor(s physmath.Store, ampFnRef artifact.ArtifactRef) (physmath.Expr, error) {
	ffam, fpayload, err := s.Lookup(ampFnRef)
	if err != nil {
		return physmath.Expr{}, fmt.Errorf("amplitude function unresolvable: %w", err)
	}
	if ffam != physmath.FamilyFunction {
		return physmath.Expr{}, fmt.Errorf("amplitude target is not function/1: %w", physmath.ErrBadInput)
	}
	fn, err := physmath.DecodeFunction(artifact.NewDecoder(fpayload))
	if err != nil {
		return physmath.Expr{}, fmt.Errorf("amplitude function undecodable: %w", err)
	}
	if len(fn.ParameterIDs) < 1 {
		return physmath.Expr{}, fmt.Errorf("amplitude function declares no parameter: %w", physmath.ErrBadInput)
	}
	return physmath.NewBoundSymbol("a", fn.ParameterIDs[0]), nil
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

// simpReplay runs SIMPLIFY-001 with replay evidence, returning the output
// ref plus a minted derivation/1 ref for protocol verification.
func simpReplay(s physmath.Store, in artifact.ArtifactRef, conds []artifact.ArtifactRef) (artifact.ArtifactRef, artifact.ArtifactRef) {
	res, err := physmath.Apply(s, physmath.Invocation{
		ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprSimplify,
		InputRefs: []artifact.ArtifactRef{in}, ConditionRefs: conds,
	})
	if err != nil {
		panic(err)
	}
	d := physmath.Derivation{
		PremiseRefs: []artifact.ArtifactRef{in},
		Steps: []physmath.DerivationStep{{
			Ordinal: 0, ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprSimplify,
			InputRefs: []artifact.ArtifactRef{in}, OutputRef: res.OutputRef,
			ConditionRefs: res.ConditionRefs,
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
