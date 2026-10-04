// Package equivariance is the canonical Phase-1 equivariance-theorem
// reference fixture owned by phys-lib/gr (plan9.3 §§10-11).
//
// Mathematical realization (closed under the seven Phase-1 rules):
//
//	f(v)       = 2 * v
//	rho_X(g,v) = g * v
//	rho_Y(g,v) = g * v
//
// Equivariance: f(rho_X(g,x)) = rho_Y(g,f(x)), both sides closing to 2*g*x
// by application + canonicalization only. No distributivity, no Map-for-f,
// no durable FunctionApplication staged sides, no GroupAction sort, no trig.
//
// Schematic g/x use the exact ordinal-derived ParameterIDs of the action
// Function declarations (§16.5). Display names are pinned golden data.
package equivariance

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-lib/gr"
	"github.com/PithomLabs/phys-math"
)

// Display names (pinned; identity-bearing per §7.4).
const (
	DisplayV = "v"
	DisplayG = "g"
	DisplayX = "x"
)

// Fixture is the fully constructed equivariance fixture: framework + GR
// physical artifacts + the six-step mathematical derivation with every
// intermediate reference pinned.
type Fixture struct {
	Store physmath.Store

	FrameworkRef artifact.ArtifactRef

	// Declaring Functions.
	F    artifact.ArtifactRef
	RhoX artifact.ArtifactRef
	RhoY artifact.ArtifactRef

	// Setup expressions.
	LHS0 artifact.ArtifactRef // Call(f, [Call(rho_X, [g, x])])
	LHS1 artifact.ArtifactRef // staged Call(f, [g*x])
	RHS0 artifact.ArtifactRef // Call(rho_Y, [g, Call(f, [x])])
	RHS1 artifact.ArtifactRef // staged Call(rho_Y, [g, 2*x])

	// Premise Call artifacts consumed by Step 0 and Step 3. Exported solely
	// so the external protocol_test can prove byte-identity with the calls
	// embedded in LHS0/RHS0 (D8). In-memory linkage only, not durable schema,
	// not general production API.
	InnerRho artifact.ArtifactRef // Call(rho_X, [g, x])
	InnerFx  artifact.ArtifactRef // Call(f, [x])

	// Step outputs (six transformations).
	Step [6]artifact.ArtifactRef

	// Theorem artifacts.
	RelationRef   artifact.ArtifactRef
	EquationRef   artifact.ArtifactRef
	TheoremRef    artifact.ArtifactRef
	DerivationRef artifact.ArtifactRef

	// GR physical artifacts.
	AssumptionRef      artifact.ArtifactRef
	InterpretationRef  artifact.ArtifactRef
	PhysicalTheoremRef artifact.ArtifactRef
}

// declareFunction mints a Function with ordinal-derived ParameterIDs after
// validating the body against the declaration (§8.6).
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
	if err := fn.Validate(); err != nil {
		return artifact.ArtifactRef{}, err
	}
	if err := physmath.ValidateFunctionBody(fn, body); err != nil {
		return artifact.ArtifactRef{}, err
	}
	fnBytes, err := fn.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(physmath.FamilyFunction, fnBytes)
}

// Build constructs the complete fixture in a fresh store, executing the
// exact six-step sequence with staged-call linkage validation (§16.9).
func Build() (*Fixture, error) {
	store := physmath.NewMemStore()
	fx := &Fixture{Store: store}
	p0 := physmath.DeriveParameterID(0)
	p1 := physmath.DeriveParameterID(1)
	g := physmath.NewBoundSymbol(DisplayG, p0)
	x := physmath.NewBoundSymbol(DisplayX, p1)
	v := physmath.NewBoundSymbol(DisplayV, p0)

	f, err := declareFunction(store, 1,
		physmath.Mul(physmath.MustRational(2, 1), v))
	if err != nil {
		return nil, err
	}
	rhoX, err := declareFunction(store, 2,
		physmath.Mul(g, physmath.NewBoundSymbol(DisplayV, p1)))
	if err != nil {
		return nil, err
	}
	rhoY, err := declareFunction(store, 2,
		physmath.Mul(g, physmath.NewBoundSymbol(DisplayV, p1)))
	if err != nil {
		return nil, err
	}
	fx.F, fx.RhoX, fx.RhoY = f, rhoX, rhoY

	put := func(e physmath.Expr) (artifact.ArtifactRef, error) {
		return physmath.PutExpression(store, e)
	}
	if fx.LHS0, err = put(physmath.Call(f, physmath.Call(rhoX, g, x))); err != nil {
		return nil, err
	}
	innerF, err := put(physmath.Call(f, x))
	if err != nil {
		return nil, err
	}
	fx.InnerFx = innerF
	if fx.RHS0, err = put(physmath.Call(rhoY, g, mustResolve(innerF, store))); err != nil {
		return nil, err
	}

	apply := func(inputs []artifact.ArtifactRef) (artifact.ArtifactRef, error) {
		res, err := physmath.Apply(store, physmath.Invocation{
			ProducerNamespace: "physmath",
			RuleID:            physmath.RuleExprApply,
			InputRefs:         inputs,
		})
		if err != nil {
			return artifact.ArtifactRef{}, err
		}
		return res.OutputRef, nil
	}
	canon := func(input artifact.ArtifactRef) (artifact.ArtifactRef, error) {
		res, err := physmath.Apply(store, physmath.Invocation{
			ProducerNamespace: "physmath",
			RuleID:            physmath.RuleExprCanon,
			InputRefs:         []artifact.ArtifactRef{input},
		})
		if err != nil {
			return artifact.ArtifactRef{}, err
		}
		return res.OutputRef, nil
	}

	// Step 0: APPLY inner rho_X(g,x) → g·x.
	innerRho, err := put(physmath.Call(rhoX, g, x))
	if err != nil {
		return nil, err
	}
	fx.InnerRho = innerRho
	if fx.Step[0], err = apply([]artifact.ArtifactRef{innerRho, rhoX}); err != nil {
		return nil, fmt.Errorf("step 0: %w", err)
	}
	// Staged LHS_1 = f(g·x) with linkage to Step 0.
	if fx.LHS1, err = put(physmath.Call(f, mustResolve(fx.Step[0], store))); err != nil {
		return nil, err
	}
	// Step 1: APPLY staged LHS_1 → 2·(g·x).
	if fx.Step[1], err = apply([]artifact.ArtifactRef{fx.LHS1, f}); err != nil {
		return nil, fmt.Errorf("step 1: %w", err)
	}
	// Step 2: CANON → 2·g·x.
	if fx.Step[2], err = canon(fx.Step[1]); err != nil {
		return nil, fmt.Errorf("step 2: %w", err)
	}
	// Step 3: APPLY inner f(x) → 2·x.
	innerFx, err := put(physmath.Call(f, x))
	if err != nil {
		return nil, err
	}
	if fx.Step[3], err = apply([]artifact.ArtifactRef{innerFx, f}); err != nil {
		return nil, fmt.Errorf("step 3: %w", err)
	}
	// Staged RHS_1 = rho_Y(g, 2·x) with linkage to Step 3.
	if fx.RHS1, err = put(physmath.Call(rhoY, g, mustResolve(fx.Step[3], store))); err != nil {
		return nil, err
	}
	// Step 4: APPLY staged RHS_1 → g·(2·x).
	if fx.Step[4], err = apply([]artifact.ArtifactRef{fx.RHS1, rhoY}); err != nil {
		return nil, fmt.Errorf("step 4: %w", err)
	}
	// Step 5: CANON → 2·g·x.
	if fx.Step[5], err = canon(fx.Step[4]); err != nil {
		return nil, fmt.Errorf("step 5: %w", err)
	}

	// Final byte-equality gate (§16.11): no hidden equality theorem.
	lhs := mustResolve(fx.Step[2], store)
	rhs := mustResolve(fx.Step[5], store)
	if !lhs.Equal(rhs) {
		return nil, fmt.Errorf("final LHS/RHS bytes differ")
	}

	// Relation → Equation → Theorem (deterministic packaging, §16.10).
	relBytes, err := physmath.EncodeRelationBody(physmath.OpEQ, lhs, rhs)
	if err != nil {
		return nil, err
	}
	if fx.RelationRef, err = store.Put(physmath.FamilyRelation, relBytes); err != nil {
		return nil, err
	}
	eq := physmath.Equation{RelationRef: fx.RelationRef}
	if err := eq.Validate(); err != nil {
		return nil, err
	}
	eqBytes := fx.RelationRef.Encode()
	if fx.EquationRef, err = store.Put(physmath.FamilyEquation, eqBytes); err != nil {
		return nil, err
	}
	steps := []physmath.DerivationStep{
		{Ordinal: 0, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{innerRho, rhoX}, OutputRef: fx.Step[0]},
		{Ordinal: 1, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{fx.LHS1, f}, OutputRef: fx.Step[1]},
		{Ordinal: 2, ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{fx.Step[1]}, OutputRef: fx.Step[2]},
		{Ordinal: 3, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{innerFx, f}, OutputRef: fx.Step[3]},
		{Ordinal: 4, ProducerNamespace: "physmath", RuleID: physmath.RuleExprApply,
			InputRefs: []artifact.ArtifactRef{fx.RHS1, rhoY}, OutputRef: fx.Step[4]},
		{Ordinal: 5, ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{fx.Step[4]}, OutputRef: fx.Step[5]},
	}
	// rho_X and rho_Y are canonically identical in the minimal fixture by
	// design (§16.5: same positional ParameterIDs, same displays, same body);
	// their physical distinction lives in framework interpretation, not in
	// duplicate mathematical identity. The SET therefore carries rhoX once.
	premises := []artifact.ArtifactRef{innerRho, innerFx, f, rhoX}
	if rhoY != rhoX {
		premises = append(premises, rhoY)
	}
	drv := physmath.Derivation{
		PremiseRefs: premises,
		Steps:       steps,
		ResultRefs:  []artifact.ArtifactRef{fx.Step[2]},
	}
	if err := drv.Validate(); err != nil {
		return nil, err
	}
	drvBytes, err := drv.Encode()
	if err != nil {
		return nil, err
	}
	if fx.DerivationRef, err = store.Put(physmath.FamilyDerivation, drvBytes); err != nil {
		return nil, err
	}
	th := physmath.Theorem{ConclusionRef: fx.RelationRef, DerivationRef: fx.DerivationRef}
	if err := th.Validate(); err != nil {
		return nil, err
	}
	thBytes, err := th.Encode()
	if err != nil {
		return nil, err
	}
	if fx.TheoremRef, err = store.Put(physmath.FamilyTheorem, thBytes); err != nil {
		return nil, err
	}

	// GR physical artifacts: framework + regularity assumption + interpretation
	// + physical theorem over the mathematical theorem.
	fw, err := gr.NewFramework("general_relativity")
	if err != nil {
		return nil, err
	}
	fwBytes, err := fw.Encode()
	if err != nil {
		return nil, err
	}
	if fx.FrameworkRef, err = store.Put(gr.FamilyFramework, fwBytes); err != nil {
		return nil, err
	}
	asm := gr.PhysicalAssumption{
		AssumptionID: "equivariance_domain",
		Category:     gr.CategoryRegularity,
		Description:  "Equivariance fixture domain: schematic action parameters with exact application.",
	}
	if err := asm.Validate(); err != nil {
		return nil, err
	}
	asmBytes, err := asm.Encode()
	if err != nil {
		return nil, err
	}
	var asmRef artifact.ArtifactRef
	if asmRef, err = store.Put(gr.FamilyAssumption, asmBytes); err != nil {
		return nil, err
	}
	fx.AssumptionRef = asmRef
	interp := gr.PhysicalInterpretation{
		InterpretationID: "equivariance_reading",
		FrameworkRef:     fx.FrameworkRef,
		PhysicalType:     "equivariant_map",
		MathematicalRefs: []artifact.ArtifactRef{fx.TheoremRef},
		AssumptionRefs:   []artifact.ArtifactRef{asmRef},
	}
	if err := interp.Validate(); err != nil {
		return nil, err
	}
	interpBytes, err := interp.Encode()
	if err != nil {
		return nil, err
	}
	if fx.InterpretationRef, err = store.Put(gr.FamilyInterpretation, interpBytes); err != nil {
		return nil, err
	}
	pt := gr.PhysicalTheorem{
		TheoremID:              "equivariance",
		FrameworkRef:           fx.FrameworkRef,
		MathematicalTheoremRef: fx.TheoremRef,
		AssumptionRefs:         []artifact.ArtifactRef{asmRef},
		InterpretationRefs:     []artifact.ArtifactRef{fx.InterpretationRef},
	}
	if err := pt.Validate(); err != nil {
		return nil, err
	}
	ptBytes, err := pt.Encode()
	if err != nil {
		return nil, err
	}
	if fx.PhysicalTheoremRef, err = store.Put(gr.FamilyTheorem, ptBytes); err != nil {
		return nil, err
	}
	return fx, nil
}

func mustResolve(ref artifact.ArtifactRef, store physmath.Store) physmath.Expr {
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

// Verify replays the fixture derivation and checks the theorem conclusion.
// It is the secondary verification entry point: independent replay plus the
// conclusion gate, over the artifacts Build produced.
func Verify(fx *Fixture) error {
	_, drvPayload, err := fx.Store.Lookup(fx.DerivationRef)
	if err != nil {
		return err
	}
	drv, err := DecodeDerivationPayload(drvPayload)
	if err != nil {
		return err
	}
	fixtures := map[artifact.ArtifactRef]bool{fx.LHS1: true, fx.RHS1: true}
	if err := physmath.Replay(fx.Store, drv, fixtures); err != nil {
		return err
	}
	if err := VerifyClaim(fx); err != nil {
		return err
	}
	_, thPayload, err := fx.Store.Lookup(fx.TheoremRef)
	if err != nil {
		return err
	}
	th, err := DecodeTheoremPayload(thPayload)
	if err != nil {
		return err
	}
	return physmath.CheckTheoremConclusion(fx.Store, th, fx.Step[2], fx.Step[5])
}
