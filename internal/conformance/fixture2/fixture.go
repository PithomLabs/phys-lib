package fixture2

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// Fixture is the independently rebuilt equivariance mathematics wrapped in
// fixture2 framework artifacts. The mathematical construction duplicates the
// canonical six-step shape using phys-math APIs ONLY — no import of
// phys-lib/gr anywhere in this package. Identical canonical bytes therefore
// prove PhysMath identity is framework-independent.
type Fixture struct {
	Store physmath.Store

	// Mathematical artifacts (same canonical content as the GR fixture's math).
	FinalRef      artifact.ArtifactRef // 2*g*x (both sides)
	RelationRef   artifact.ArtifactRef
	EquationRef   artifact.ArtifactRef
	TheoremRef    artifact.ArtifactRef
	DerivationRef artifact.ArtifactRef

	// Fixture2 framework artifacts (different meaning, different identity).
	FrameworkRef      artifact.ArtifactRef
	AssumptionRef     artifact.ArtifactRef
	DefinitionRef     artifact.ArtifactRef
	EquationFWRef     artifact.ArtifactRef
	InterpretationRef artifact.ArtifactRef
	TheoremFWRef      artifact.ArtifactRef
}

func declareFunction(store physmath.Store, n int, body physmath.Expr) (artifact.ArtifactRef, error) {
	ids := make([]physmath.ParameterID, n)
	for i := range ids {
		ids[i] = physmath.DeriveParameterID(uint32(i))
	}
	bodyBytes, err := body.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	bodyRef, err := store.Put(physmath.FamilyExpression, bodyBytes)
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	fn := physmath.Function{ParameterIDs: ids, BodyRef: bodyRef}
	if err := physmath.ValidateFunctionBody(fn, body); err != nil {
		return artifact.ArtifactRef{}, err
	}
	fnBytes, err := fn.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(physmath.FamilyFunction, fnBytes)
}

// Build independently rebuilds the six-step equivariance derivation and wraps
// it in fixture2 framework artifacts with deliberately different physical
// meaning (test-double interpretation, distinct assumption vocabulary).
func Build() (*Fixture, error) {
	store := physmath.NewMemStore()
	fx := &Fixture{Store: store}
	p0 := physmath.DeriveParameterID(0)
	p1 := physmath.DeriveParameterID(1)
	g := physmath.NewBoundSymbol("g", p0)
	x := physmath.NewBoundSymbol("x", p1)
	v := physmath.NewBoundSymbol("v", p0)

	f, err := declareFunction(store, 1, physmath.Mul(physmath.MustRational(2, 1), v))
	if err != nil {
		return nil, err
	}
	rho, err := declareFunction(store, 2,
		physmath.Mul(g, physmath.NewBoundSymbol("v", p1)))
	if err != nil {
		return nil, err
	}
	put := func(e physmath.Expr) (artifact.ArtifactRef, error) {
		return physmath.PutExpression(store, e)
	}
	apply := func(inputs []artifact.ArtifactRef) (artifact.ArtifactRef, error) {
		res, err := physmath.Apply(store, physmath.Invocation{
			ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply, InputRefs: inputs})
		if err != nil {
			return artifact.ArtifactRef{}, err
		}
		return res.OutputRef, nil
	}
	canon := func(input artifact.ArtifactRef) (artifact.ArtifactRef, error) {
		res, err := physmath.Apply(store, physmath.Invocation{
			ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{input}})
		if err != nil {
			return artifact.ArtifactRef{}, err
		}
		return res.OutputRef, nil
	}
	resolve := func(ref artifact.ArtifactRef) physmath.Expr {
		_, payload, err := store.Lookup(ref)
		if err != nil {
			panic(err)
		}
		e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
		if err != nil {
			panic(err)
		}
		return e
	}

	innerRho, err := put(physmath.Call(rho, g, x))
	if err != nil {
		return nil, err
	}
	innerF, err := put(physmath.Call(f, x))
	if err != nil {
		return nil, err
	}
	s0, err := apply([]artifact.ArtifactRef{innerRho, rho})
	if err != nil {
		return nil, err
	}
	lhs1, err := put(physmath.Call(f, resolve(s0)))
	if err != nil {
		return nil, err
	}
	s1, err := apply([]artifact.ArtifactRef{lhs1, f})
	if err != nil {
		return nil, err
	}
	s2, err := canon(s1)
	if err != nil {
		return nil, err
	}
	s3, err := apply([]artifact.ArtifactRef{innerF, f})
	if err != nil {
		return nil, err
	}
	rhs1, err := put(physmath.Call(rho, g, resolve(s3)))
	if err != nil {
		return nil, err
	}
	s4, err := apply([]artifact.ArtifactRef{rhs1, rho})
	if err != nil {
		return nil, err
	}
	s5, err := canon(s4)
	if err != nil {
		return nil, err
	}
	lhs, rhs := resolve(s2), resolve(s5)
	if !lhs.Equal(rhs) {
		return nil, fmt.Errorf("fixture2: final sides differ")
	}
	fx.FinalRef = s2

	relBytes, err := physmath.EncodeRelationBody(physmath.OpEQ, lhs, rhs)
	if err != nil {
		return nil, err
	}
	if fx.RelationRef, err = store.Put(physmath.FamilyRelation, relBytes); err != nil {
		return nil, err
	}
	if fx.EquationRef, err = store.Put(physmath.FamilyEquation, fx.RelationRef.Encode()); err != nil {
		return nil, err
	}
	steps := []physmath.DerivationStep{
		{Ordinal: 0, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{innerRho, rho}, OutputRef: s0},
		{Ordinal: 1, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{lhs1, f}, OutputRef: s1},
		{Ordinal: 2, ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{s1}, OutputRef: s2},
		{Ordinal: 3, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{innerF, f}, OutputRef: s3},
		{Ordinal: 4, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{rhs1, rho}, OutputRef: s4},
		{Ordinal: 5, ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{s4}, OutputRef: s5},
	}
	drv := physmath.Derivation{
		PremiseRefs: []artifact.ArtifactRef{innerRho, innerF, f, rho},
		Steps:       steps,
		ResultRefs:  []artifact.ArtifactRef{s2},
	}
	drvBytes, err := drv.Encode()
	if err != nil {
		return nil, err
	}
	if fx.DerivationRef, err = store.Put(physmath.FamilyDerivation, drvBytes); err != nil {
		return nil, err
	}
	th := physmath.Theorem{ConclusionRef: fx.RelationRef, DerivationRef: fx.DerivationRef}
	thBytes, err := th.Encode()
	if err != nil {
		return nil, err
	}
	if fx.TheoremRef, err = store.Put(physmath.FamilyTheorem, thBytes); err != nil {
		return nil, err
	}
	if err := physmath.Replay(store, drv, map[artifact.ArtifactRef]bool{lhs1: true, rhs1: true}); err != nil {
		return nil, err
	}
	if err := physmath.CheckTheoremConclusion(store, th, s2, s5); err != nil {
		return nil, err
	}

	// Fixture2 framework layer: same math, different meaning.
	fw, err := NewFramework()
	if err != nil {
		return nil, err
	}
	if fx.FrameworkRef, err = fw.Mint(store); err != nil {
		return nil, err
	}
	asm := Assumption{
		AssumptionID: "test_domain",
		Category:     "test-postulate",
		Description:  "Test-framework postulate: schematic doubling actions compose.",
	}
	if err := asm.Validate(); err != nil {
		return nil, err
	}
	if fx.AssumptionRef, err = asm.Mint(store); err != nil {
		return nil, err
	}
	def := Definition{
		DefinitionID:     "doubling_map",
		FrameworkRef:     fx.FrameworkRef,
		PhysicalType:     "test_double",
		MathematicalRefs: []artifact.ArtifactRef{fx.TheoremRef},
		AssumptionRefs:   []artifact.ArtifactRef{fx.AssumptionRef},
	}
	if err := def.Validate(); err != nil {
		return nil, err
	}
	if fx.DefinitionRef, err = def.Mint(store); err != nil {
		return nil, err
	}
	eq := Equation{
		EquationID:     "doubling_equation",
		FrameworkRef:   fx.FrameworkRef,
		EquationRef:    fx.EquationRef,
		AssumptionRefs: []artifact.ArtifactRef{fx.AssumptionRef},
	}
	if err := eq.Validate(); err != nil {
		return nil, err
	}
	if fx.EquationFWRef, err = eq.Mint(store); err != nil {
		return nil, err
	}
	interp := Interpretation{
		InterpretationID: "doubling_reading",
		FrameworkRef:     fx.FrameworkRef,
		PhysicalType:     "test_rescaling",
		MathematicalRefs: []artifact.ArtifactRef{fx.TheoremRef},
		AssumptionRefs:   []artifact.ArtifactRef{fx.AssumptionRef},
	}
	if err := interp.Validate(); err != nil {
		return nil, err
	}
	if fx.InterpretationRef, err = interp.Mint(store); err != nil {
		return nil, err
	}
	pt := Theorem{
		TheoremID:              "doubling",
		FrameworkRef:           fx.FrameworkRef,
		MathematicalTheoremRef: fx.TheoremRef,
		AssumptionRefs:         []artifact.ArtifactRef{fx.AssumptionRef},
		InterpretationRefs:     []artifact.ArtifactRef{fx.InterpretationRef},
	}
	if err := pt.Validate(); err != nil {
		return nil, err
	}
	if fx.TheoremFWRef, err = pt.Mint(store); err != nil {
		return nil, err
	}
	return fx, nil
}
