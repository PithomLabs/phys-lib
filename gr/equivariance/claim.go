package equivariance

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// VerifyClaim is exported solely to support the canonical external
// protocol_test fixture (D8 nested layout). It is not general production API.
//
// VerifyClaim establishes the H-3 semantic invariant using only existing
// Phase-1 mechanisms (no new families, nodes, rules, or schema):
//
//	LHS0 → declared left derivation/result  → conclusion.LHS
//	RHS0 → declared right derivation/result → conclusion.RHS
//
// Concretely it proves, by structural decomposition plus byte-identity
// linkage plus replayed rule execution (established by Replay/Verify):
//
//	LHS0 = Call(F, [innerRho])  with innerRho = Call(RhoX, [g, x])
//	innerRho artifact  == LHS0.Args[0]          (original-to-input linkage)
//	Step0 = APPLY(innerRho)                      (replay)
//	LHS1  == Call(F, [Step0-output])             (staged linkage)
//	Step1 = APPLY(LHS1), Step2 = CANON(Step1)    (replay)
//	conclusion.LHS == Step2                      (conclusion gate)
//	and symmetrically RHS0 → Step5 → conclusion.RHS.
//
// Every theorem-side artifact is therefore semantically live: corrupting
// LHS0, RHS0, the declarations, or any linkage breaks verification.
// Normal-form equality alone (Step2 == Step5) is necessary but not
// sufficient, and is never the whole check.
func VerifyClaim(fx *Fixture) error {
	p0 := physmath.DeriveParameterID(0)
	p1 := physmath.DeriveParameterID(1)
	lhs0 := mustResolve(fx.LHS0, fx.Store)
	rhs0 := mustResolve(fx.RHS0, fx.Store)

	// Original left side: LHS0 = Call(F, [Call(RhoX, [g(P0), x(P1)])]).
	if err := requireCallShape(lhs0, fx.F, 1, "LHS0"); err != nil {
		return err
	}
	innerRho := lhs0.Args[0]
	if err := requireCallArgs(innerRho, fx.RhoX, []physmath.Expr{
		schematicSymbol("g", p0),
		schematicSymbol("x", p1),
	}, "LHS0 inner rho_X call"); err != nil {
		return err
	}
	// Original right side: RHS0 = Call(RhoY, [g(P0), Call(F, [x(P1)])]).
	if err := requireCallShape(rhs0, fx.RhoY, 2, "RHS0"); err != nil {
		return err
	}
	if err := requireSchematic(rhs0.Args[0], p0, "g", "RHS0 g"); err != nil {
		return err
	}
	innerF := rhs0.Args[1]
	if err := requireCallArgs(innerF, fx.F, []physmath.Expr{
		schematicSymbol("x", p1),
	}, "RHS0 inner f call"); err != nil {
		return err
	}
	// Original-to-input linkage: the derivation premises carry exactly the
	// calls embedded in the original sides (byte-identity, not resemblance).
	premiseRho := mustResolve(fx.InnerRho, fx.Store)
	if !premiseRho.Equal(innerRho) {
		return fmt.Errorf("equivariance: LHS0 inner call != Step-0 input artifact")
	}
	premiseF := mustResolve(fx.InnerFx, fx.Store)
	if !premiseF.Equal(innerF) {
		return fmt.Errorf("equivariance: RHS0 inner call != Step-3 input artifact")
	}
	// Staged linkage: LHS1 == Call(F, [Step0-output]), RHS1 == Call(RhoY, [g, Step3-output]).
	lhs1 := mustResolve(fx.LHS1, fx.Store)
	s0out := mustResolve(fx.Step[0], fx.Store)
	if err := requireCallArgs(lhs1, fx.F, []physmath.Expr{s0out}, "staged LHS1"); err != nil {
		return err
	}
	rhs1 := mustResolve(fx.RHS1, fx.Store)
	s3out := mustResolve(fx.Step[3], fx.Store)
	g := schematicSymbol("g", p0)
	if err := requireCallArgs(rhs1, fx.RhoY, []physmath.Expr{g, s3out}, "staged RHS1"); err != nil {
		return err
	}
	return nil
}

func schematicSymbol(display string, pid physmath.ParameterID) physmath.Expr {
	return physmath.NewBoundSymbol(display, pid)
}

// requireCallShape asserts root-Call shape with the exact declaring Function
// reference and arity. Arguments are inspected by the caller explicitly.
func requireCallShape(e physmath.Expr, fn artifact.ArtifactRef, arity int, what string) error {
	if e.Kind != physmath.TagCall {
		return fmt.Errorf("equivariance: %s is not a Call", what)
	}
	if e.FuncRef != fn {
		return fmt.Errorf("equivariance: %s references the wrong declaration", what)
	}
	if len(e.Args) != arity {
		return fmt.Errorf("equivariance: %s arity %d, want %d", what, len(e.Args), arity)
	}
	return nil
}

// requireCallArgs asserts full Call shape including exact argument bytes.
func requireCallArgs(e physmath.Expr, fn artifact.ArtifactRef, args []physmath.Expr, what string) error {
	if e.Kind != physmath.TagCall {
		return fmt.Errorf("equivariance: %s is not a Call", what)
	}
	if e.FuncRef != fn {
		return fmt.Errorf("equivariance: %s references the wrong declaration", what)
	}
	if len(e.Args) != len(args) {
		return fmt.Errorf("equivariance: %s arity %d, want %d", what, len(e.Args), len(args))
	}
	for i := range args {
		if !e.Args[i].Equal(args[i]) {
			return fmt.Errorf("equivariance: %s argument %d mismatch", what, i)
		}
	}
	return nil
}

// requireSchematic asserts an exact schematic parameter symbol: Symbol kind,
// exact display, exact ParameterID, no ConstantID. Display spelling confers
// no meaning; exactness here is byte-identity of the declared slot.
func requireSchematic(e physmath.Expr, pid physmath.ParameterID, display, what string) error {
	if e.Kind != physmath.TagSymbol {
		return fmt.Errorf("equivariance: %s is not a Symbol", what)
	}
	if e.DisplayName != display {
		return fmt.Errorf("equivariance: %s display %q, want %q", what, e.DisplayName, display)
	}
	if e.ParamID == nil || *e.ParamID != pid {
		return fmt.Errorf("equivariance: %s is not bound to the declared slot", what)
	}
	if e.ConstantID != nil {
		return fmt.Errorf("equivariance: %s carries a ConstantID", what)
	}
	return nil
}
