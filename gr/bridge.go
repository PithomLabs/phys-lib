package gr

import (
	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// CaptureCondition implements the D2 mechanical bridge: a MathCondition that
// a PhysMath rule surfaced (via TransformResult threading — rules generate
// nothing) is captured by reference into a framework-owned PhysicalAssumption
// with the given GR-local Category. The relevant PhysicalTheorem or
// PhysicalEquation then references the assumption. No condition is silently
// discharged: the assumption records exactly the condition the rule required.
func CaptureCondition(store physmath.Store, condRef artifact.ArtifactRef, assumptionID, category, description string) (PhysicalAssumption, error) {
	fam, _, err := store.Lookup(condRef)
	if err != nil {
		return PhysicalAssumption{}, &Error{Op: "Bridge.CaptureCondition", Err: err}
	}
	if fam != physmath.FamilyCondition {
		return PhysicalAssumption{}, &Error{Op: "Bridge.CaptureCondition",
			Err: &categoryError{"condition reference is not physmath/condition/1"}}
	}
	a := PhysicalAssumption{
		AssumptionID:  assumptionID,
		ConditionRefs: []artifact.ArtifactRef{condRef},
		Category:      category,
		Description:   description,
	}
	if err := a.Validate(); err != nil {
		return PhysicalAssumption{}, err
	}
	return a, nil
}

type categoryError struct{ msg string }

func (e *categoryError) Error() string { return "gr: " + e.msg }
