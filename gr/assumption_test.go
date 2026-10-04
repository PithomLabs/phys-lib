package gr

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// Category is a closed GR-package-contract vocabulary (D15), not global.
func TestCategoryClosed(t *testing.T) {
	for _, c := range []string{CategoryConvention, CategoryRegularity, CategoryModel} {
		if !ValidCategory(c) {
			t.Fatalf("%q rejected", c)
		}
	}
	for _, c := range []string{"", "quantum-postulate", "CONVENTION", "physical_law"} {
		if ValidCategory(c) {
			t.Fatalf("%q accepted into GR vocabulary", c)
		}
		a := PhysicalAssumption{AssumptionID: "x", Category: c, Description: "d"}
		if err := a.Validate(); err == nil {
			t.Fatalf("assumption with Category %q validated", c)
		}
	}
}

func TestAssumptionVectors(t *testing.T) {
	s := testStore()
	a := PhysicalAssumption{AssumptionID: "r_positive", Category: CategoryRegularity,
		Description: "Coordinate regularity: r != 0 on the domain."}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	r1, err := a.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := a.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	if r1 != r2 {
		t.Fatal("assumption identity not deterministic")
	}
	fam, _, _ := s.Lookup(r1)
	if fam != FamilyAssumption {
		t.Fatal("minted under wrong family")
	}
	// Duplicate ConditionRefs rejected (SET semantics).
	dup := PhysicalAssumption{AssumptionID: "x", Category: CategoryRegularity,
		Description: "d", ConditionRefs: []artifact.ArtifactRef{r1, r1}}
	if err := dup.Validate(); err == nil {
		t.Fatal("duplicate ConditionRefs accepted")
	}
	// ConditionRefs must target condition/1: an assumption ref is rejected.
	fams := map[artifact.ArtifactRef]physmath.FamilyID{r1: FamilyAssumption}
	wrong := PhysicalAssumption{AssumptionID: "x", Category: CategoryRegularity,
		Description: "d", ConditionRefs: []artifact.ArtifactRef{r1}}
	if err := physmath.ValidateAgainst(wrong.TargetChecks(), fams); err == nil {
		t.Fatal("ConditionRefs → assumption/1 accepted")
	}
}
