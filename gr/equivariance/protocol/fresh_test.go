package protocol_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-lib/gr/equivariance"
	"github.com/PithomLabs/phys-math"
)

// buildPersisted returns ONLY the durable entry points a real verifier
// starts from: the content-addressed store, the derivation ref, and the
// theorem ref. No Fixture semantic fields (LHS0/RHS0/InnerRho/InnerFx)
// escape this function; the caller must discard the Fixture.
func buildPersisted(t *testing.T) (physmath.Store, artifact.ArtifactRef, artifact.ArtifactRef) {
	t.Helper()
	fx, err := equivariance.Build()
	if err != nil {
		t.Fatal(err)
	}
	store, drvRef, thRef := fx.Store, fx.DerivationRef, fx.TheoremRef
	fx = nil
	return store, drvRef, thRef
}

func resolveExpr(t *testing.T, store physmath.Store, ref artifact.ArtifactRef) physmath.Expr {
	t.Helper()
	fam, payload, err := store.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyExpression {
		t.Fatalf("expected expression/1, got %s", fam.ArtifactType)
	}
	e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func resolvePayload(t *testing.T, store physmath.Store, ref artifact.ArtifactRef) []byte {
	t.Helper()
	_, payload, err := store.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// freshVerify reconstructs the equivariance proposition and its trace to the
// conclusion using ONLY persisted artifacts: the derivation payload, theorem
// payload, and transitively resolved expression/function/relation payloads.
// Manifest data is never consulted here (metadata firewall); Fixture-memory
// fields are inaccessible by construction (only refs cross the boundary).
func freshVerify(t *testing.T, store physmath.Store, drvRef, thRef artifact.ArtifactRef) error {
	t.Helper()
	drvPayload := resolvePayload(t, store, drvRef)
	drv, err := equivariance.DecodeDerivationPayload(drvPayload)
	if err != nil {
		return err
	}
	if len(drv.Steps) != 6 {
		return errf("want 6 steps, got %d", len(drv.Steps))
	}
	wantRules := []string{
		physmath.RuleExprApply, physmath.RuleExprApply, physmath.RuleExprCanon,
		physmath.RuleExprApply, physmath.RuleExprApply, physmath.RuleExprCanon,
	}
	for i, s := range drv.Steps {
		if s.RuleID != wantRules[i] {
			return fmt.Errorf("step %d rule %s", i, s.RuleID)
		}
	}
	// Availability context discovered from persisted data only: premises
	// plus every step input (prior outputs are available by construction;
	// over-supplying fixtures cannot forge replay equality).
	fixtures := map[artifact.ArtifactRef]bool{}
	for _, r := range drv.PremiseRefs {
		fixtures[r] = false // premises need no fixture entry
	}
	for _, s := range drv.Steps {
		for _, r := range s.InputRefs {
			if _, ok := fixtures[r]; !ok {
				fixtures[r] = true
			}
		}
	}
	onlyFixtures := map[artifact.ArtifactRef]bool{}
	for r, isFixture := range fixtures {
		if isFixture {
			onlyFixtures[r] = true
		}
	}
	if err := physmath.Replay(store, drv, onlyFixtures); err != nil {
		return err
	}

	// Recover the left proposition side from persisted step data.
	s0in := resolveExpr(t, store, drv.Steps[0].InputRefs[0])
	rhoXDecl := drv.Steps[0].InputRefs[1]
	gExpr, xExpr, err := callArgs2(t, s0in, rhoXDecl, "Step-0 input")
	if err != nil {
		return err
	}
	p0 := physmath.DeriveParameterID(0)
	p1 := physmath.DeriveParameterID(1)
	requireSchematic(t, gExpr, p0, "g", "Step-0 g")
	requireSchematic(t, xExpr, p1, "x", "Step-0 x")
	fDecl := drv.Steps[1].InputRefs[1]
	lhs1 := resolveExpr(t, store, drv.Steps[1].InputRefs[0])
	s0out := resolveExpr(t, store, drv.Steps[0].OutputRef)
	requireCallShape(t, lhs1, fDecl, 1, "staged LHS1")
	if !lhs1.Args[0].Equal(s0out) {
		return errf("staged LHS1 argument != Step-0 output")
	}
	// Reconstructed original left proposition: f(rho_X(g,x)).
	lhs0 := physmath.Call(fDecl, mustCall(t, rhoXDecl, gExpr, xExpr))
	_ = lhs0

	// Recover the right proposition side.
	s3in := resolveExpr(t, store, drv.Steps[3].InputRefs[0])
	fDecl3 := drv.Steps[3].InputRefs[1]
	if fDecl3 != fDecl {
		return errf("map declaration differs between branches")
	}
	xInner, err := callArgs1(t, s3in, fDecl3, "Step-3 input")
	if err != nil {
		return err
	}
	if !xInner.Equal(xExpr) {
		return errf("right-branch variable != left-branch variable")
	}
	rhs1 := resolveExpr(t, store, drv.Steps[4].InputRefs[0])
	rhoDecl := drv.Steps[4].InputRefs[1]
	s3out := resolveExpr(t, store, drv.Steps[3].OutputRef)
	requireCallShape(t, rhs1, rhoDecl, 2, "staged RHS1")
	if !rhs1.Args[0].Equal(gExpr) {
		return errf("staged RHS1 first argument != g")
	}
	if !rhs1.Args[1].Equal(s3out) {
		return errf("staged RHS1 argument != Step-3 output")
	}
	// Reconstructed original right proposition: rho_Y(g,f(x)).
	rhs0 := physmath.Call(rhoDecl, gExpr, mustCall(t, fDecl, xExpr))
	_ = rhs0

	// Conclusion binding, decoded structurally (no encoder trust): the
	// durable relation must be EQ(final-left, final-right) with sides
	// byte-identical to the designated final derivation outputs.
	thPayload := resolvePayload(t, store, thRef)
	th, err := equivariance.DecodeTheoremPayload(thPayload)
	if err != nil {
		return err
	}
	relPayload := resolvePayload(t, store, th.ConclusionRef)
	d := artifact.NewDecoder(relPayload)
	op, err := d.String()
	if err != nil {
		return err
	}
	if op != "EQ" {
		return errf("conclusion operator %q", op)
	}
	s2Payload := resolvePayload(t, store, drv.Steps[2].OutputRef)
	s5Payload := resolvePayload(t, store, drv.Steps[5].OutputRef)
	leftBytes, err := readEmbeddedExpr(d)
	if err != nil {
		return err
	}
	rightBytes, err := readEmbeddedExpr(d)
	if err != nil {
		return err
	}
	if !d.Exhausted() {
		return errf("trailing bytes in conclusion relation")
	}
	if !bytes.Equal(leftBytes, s2Payload) {
		return errf("conclusion.LHS != Step-2 bytes")
	}
	if !bytes.Equal(rightBytes, s5Payload) {
		return errf("conclusion.RHS != Step-5 bytes")
	}
	if !bytes.Equal(s2Payload, s5Payload) {
		return errf("final sides differ")
	}
	return nil
}

func readEmbeddedExpr(d *artifact.Decoder) ([]byte, error) {
	e, err := physmath.DecodeExpr(d)
	if err != nil {
		return nil, err
	}
	return e.Encode()
}

func callArgs2(t *testing.T, e physmath.Expr, fn artifact.ArtifactRef, what string) (physmath.Expr, physmath.Expr, error) {
	t.Helper()
	requireCallShape(t, e, fn, 2, what)
	return e.Args[0], e.Args[1], nil
}

func callArgs1(t *testing.T, e physmath.Expr, fn artifact.ArtifactRef, what string) (physmath.Expr, error) {
	t.Helper()
	requireCallShape(t, e, fn, 1, what)
	return e.Args[0], nil
}

func mustCall(t *testing.T, fn artifact.ArtifactRef, args ...physmath.Expr) physmath.Expr {
	t.Helper()
	return physmath.Call(fn, args...)
}

func requireCallShape(t *testing.T, e physmath.Expr, fn artifact.ArtifactRef, arity int, what string) {
	t.Helper()
	if e.Kind != physmath.TagCall {
		t.Fatalf("fresh verifier: %s is not a Call", what)
	}
	if e.FuncRef != fn {
		t.Fatalf("fresh verifier: %s references the wrong declaration", what)
	}
	if len(e.Args) != arity {
		t.Fatalf("fresh verifier: %s arity %d", what, len(e.Args))
	}
}

func requireSchematic(t *testing.T, e physmath.Expr, pid physmath.ParameterID, display, what string) {
	t.Helper()
	if e.Kind != physmath.TagSymbol || e.DisplayName != display {
		t.Fatalf("fresh verifier: %s is not schematic %q", what, display)
	}
	if e.ParamID == nil || *e.ParamID != pid {
		t.Fatalf("fresh verifier: %s is not bound to the declared slot", what)
	}
	if e.ConstantID != nil {
		t.Fatalf("fresh verifier: %s carries a ConstantID", what)
	}
}

type freshError string

func (e freshError) Error() string { return "fresh verifier: " + string(e) }
func errf(format string, args ...any) error {
	return freshError(fmt.Sprintf(format, args...))
}
