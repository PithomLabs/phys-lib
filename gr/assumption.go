package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// Assumption categories: a CLOSED GR-package-contract vocabulary
// (plan9.3 D15). This constrains phys-lib/gr only — it is NOT a global
// restriction on future phys-* frameworks (spec §14.2 preserved).
const (
	CategoryConvention = "convention" // e.g. metric signature, curvature convention, Λ choice
	CategoryRegularity = "regularity" // e.g. domain/non-vanishing requirements
	CategoryModel      = "model"      // e.g. model/interpretation assumptions
)

// ValidCategory reports membership in the closed GR vocabulary.
func ValidCategory(c string) bool {
	switch c {
	case CategoryConvention, CategoryRegularity, CategoryModel:
		return true
	}
	return false
}

// PhysicalAssumption is a package-owned GR assumption (spec §14.2).
// Family phys-gr/assumption/1. Description is identity-bearing explanatory
// content. Category uses the closed GR vocabulary above.
type PhysicalAssumption struct {
	AssumptionID  string
	ConditionRefs []artifact.ArtifactRef // SET → physmath/condition/1
	Category      string
	Description   string
}

// Family returns FamilyAssumption.
func (PhysicalAssumption) Family() physmath.FamilyID { return FamilyAssumption }

// Validate checks fields (reference targets via TargetChecks + resolution).
func (a PhysicalAssumption) Validate() error {
	if err := validateIdentifier("AssumptionID", a.AssumptionID); err != nil {
		return err
	}
	if !ValidCategory(a.Category) {
		return &Error{Op: "Assumption.Category", Err: ErrUnknownCategory}
	}
	if a.Description == "" {
		return &Error{Op: "Assumption.Description",
			Err: fmt.Errorf("description is required package-owned content")}
	}
	seen := map[artifact.ArtifactRef]bool{}
	for _, r := range a.ConditionRefs {
		if seen[r] {
			return &Error{Op: "Assumption.ConditionRefs",
				Err: fmt.Errorf("duplicate reference")}
		}
		seen[r] = true
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (a PhysicalAssumption) Encode() ([]byte, error) {
	id, err := artifact.EncodeString(a.AssumptionID)
	if err != nil {
		return nil, err
	}
	elems := make([][]byte, len(a.ConditionRefs))
	for i, r := range a.ConditionRefs {
		elems[i] = r.Encode()
	}
	set, err := artifact.EncodeSet(elems)
	if err != nil {
		return nil, err
	}
	out := append(id, set...)
	cat, err := artifact.EncodeString(a.Category)
	if err != nil {
		return nil, err
	}
	out = append(out, cat...)
	desc, err := artifact.EncodeString(a.Description)
	if err != nil {
		return nil, err
	}
	return append(out, desc...), nil
}

// TargetChecks: ConditionRefs → physmath/condition/1.
func (a PhysicalAssumption) TargetChecks() []physmath.TargetCheck {
	checks := make([]physmath.TargetCheck, len(a.ConditionRefs))
	for i, r := range a.ConditionRefs {
		checks[i] = physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalAssumption.ConditionRefs[%d]", i),
			Ref:     r,
			Allowed: []physmath.FamilyID{physmath.FamilyCondition},
		}
	}
	return checks
}

// Mint stores the canonical payload and returns its reference.
func (a PhysicalAssumption) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := a.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(a.Family(), payload)
}
