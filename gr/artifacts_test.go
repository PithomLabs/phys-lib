package gr

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// fixtureFramework mints the GR framework and returns its ref plus a families
// map seeded with the framework entry.
func fixtureFramework(t *testing.T, s physmath.Store) (artifact.ArtifactRef, map[artifact.ArtifactRef]physmath.FamilyID) {
	t.Helper()
	fw, err := NewFramework("general_relativity")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := fw.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	return ref, map[artifact.ArtifactRef]physmath.FamilyID{ref: FamilyFramework}
}

func putExpression(t *testing.T, s physmath.Store, e physmath.Expr) artifact.ArtifactRef {
	t.Helper()
	ref, err := physmath.PutExpression(s, e)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func TestDefinitionFamily(t *testing.T) {
	s := testStore()
	fwRef, fams := fixtureFramework(t, s)
	expr := putExpression(t, s, physmath.NewEphemeralSymbol("g"))
	fams[expr] = physmath.FamilyExpression
	asm := PhysicalAssumption{AssumptionID: "a", Category: CategoryModel, Description: "d"}
	asmRef, err := asm.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	fams[asmRef] = FamilyAssumption
	def := PhysicalDefinition{
		DefinitionID:     "metric_tensor",
		FrameworkRef:     fwRef,
		PhysicalType:     "lorentzian_metric",
		MathematicalRefs: []artifact.ArtifactRef{expr},
		AssumptionRefs:   []artifact.ArtifactRef{asmRef},
	}
	if err := def.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := physmath.ValidateAgainst(def.TargetChecks(), fams); err != nil {
		t.Fatal(err)
	}
	r1, err := def.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := def.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	if r1 != r2 {
		t.Fatal("definition identity not deterministic")
	}
	// Wrong FrameworkRef target rejected.
	other := def
	other.FrameworkRef = expr
	if err := physmath.ValidateAgainst(other.TargetChecks(), fams); err == nil {
		t.Fatal("FrameworkRef → expression/1 accepted")
	}
	// DefinitionEquation must target equation/1 when present.
	eqPayload := expr // reuse bytes; family map decides
	fams[eqPayload] = physmath.FamilyEquation
	withEq := def
	withEq.DefinitionEquation = &eqPayload
	if err := physmath.ValidateAgainst(withEq.TargetChecks(), fams); err != nil {
		t.Fatal(err)
	}
}

func TestEquationFamily(t *testing.T) {
	s := testStore()
	fwRef, fams := fixtureFramework(t, s)
	eqRef, _ := artifact.ComputeRef("physmath", "1", "equation/1", []byte("eq"))
	fams[eqRef] = physmath.FamilyEquation
	asm := PhysicalAssumption{AssumptionID: "a", Category: CategoryConvention, Description: "d"}
	asmRef, _ := asm.Mint(s)
	fams[asmRef] = FamilyAssumption
	pe := PhysicalEquation{EquationID: "vacuum", FrameworkRef: fwRef, EquationRef: eqRef,
		AssumptionRefs: []artifact.ArtifactRef{asmRef}, ScopeRefs: []artifact.ArtifactRef{asmRef}}
	if err := pe.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := physmath.ValidateAgainst(pe.TargetChecks(), fams); err != nil {
		t.Fatal(err)
	}
	if _, err := pe.Mint(s); err != nil {
		t.Fatal(err)
	}
	// ScopeRefs limited to assumption artifacts: equation ref rejected.
	bad := pe
	bad.ScopeRefs = []artifact.ArtifactRef{eqRef}
	if err := physmath.ValidateAgainst(bad.TargetChecks(), fams); err == nil {
		t.Fatal("ScopeRefs → equation/1 accepted")
	}
}

func TestTheoremAndInterpretationFamilies(t *testing.T) {
	s := testStore()
	fwRef, fams := fixtureFramework(t, s)
	thmRef, _ := artifact.ComputeRef("physmath", "1", "theorem/1", []byte("th"))
	fams[thmRef] = physmath.FamilyTheorem
	asm := PhysicalAssumption{AssumptionID: "a", Category: CategoryRegularity, Description: "d"}
	asmRef, _ := asm.Mint(s)
	fams[asmRef] = FamilyAssumption
	interp := PhysicalInterpretation{InterpretationID: "i", FrameworkRef: fwRef,
		PhysicalType:     "vacuum_spacetime",
		MathematicalRefs: []artifact.ArtifactRef{thmRef},
		AssumptionRefs:   []artifact.ArtifactRef{asmRef}}
	// MathematicalRefs may target any of the 16 physmath families.
	mathFams := append([]physmath.FamilyID{}, physmath.AllPhysMathFamilies()...)
	_ = mathFams
	fams[thmRef] = physmath.FamilyTheorem
	if err := interp.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := physmath.ValidateAgainst(interp.TargetChecks(), fams); err != nil {
		t.Fatal(err)
	}
	intRef, err := interp.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	fams[intRef] = FamilyInterpretation
	pt := PhysicalTheorem{TheoremID: "t", FrameworkRef: fwRef,
		MathematicalTheoremRef: thmRef,
		AssumptionRefs:         []artifact.ArtifactRef{asmRef},
		InterpretationRefs:     []artifact.ArtifactRef{intRef}}
	if err := pt.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := physmath.ValidateAgainst(pt.TargetChecks(), fams); err != nil {
		t.Fatal(err)
	}
	if _, err := pt.Mint(s); err != nil {
		t.Fatal(err)
	}
	// MathematicalTheoremRef must be theorem/1.
	bad := pt
	bad.MathematicalTheoremRef = asmRef
	if err := physmath.ValidateAgainst(bad.TargetChecks(), fams); err == nil {
		t.Fatal("MathematicalTheoremRef → assumption/1 accepted")
	}
}
