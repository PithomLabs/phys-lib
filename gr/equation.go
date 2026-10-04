package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// PhysicalEquation adds framework meaning and assumptions to a PhysMath
// equation (spec §14.4). The mathematical equation remains a PhysMath
// artifact. Family phys-gr/equation/1.
type PhysicalEquation struct {
	EquationID     string
	FrameworkRef   artifact.ArtifactRef
	EquationRef    artifact.ArtifactRef
	AssumptionRefs []artifact.ArtifactRef // SET: owning/closure assumption/1
	ScopeRefs      []artifact.ArtifactRef // SET: owning/closure assumption/1
}

// Family returns FamilyEquation.
func (PhysicalEquation) Family() physmath.FamilyID { return FamilyEquation }

// Validate checks fields (reference targets via TargetChecks + resolution).
func (e PhysicalEquation) Validate() error {
	if err := validateIdentifier("EquationID", e.EquationID); err != nil {
		return err
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (e PhysicalEquation) Encode() ([]byte, error) {
	id, err := artifact.EncodeString(e.EquationID)
	if err != nil {
		return nil, err
	}
	out := append(id, e.FrameworkRef.Encode()...)
	out = append(out, e.EquationRef.Encode()...)
	asmElems := make([][]byte, len(e.AssumptionRefs))
	for i, r := range e.AssumptionRefs {
		asmElems[i] = r.Encode()
	}
	asmSet, err := artifact.EncodeSet(asmElems)
	if err != nil {
		return nil, err
	}
	out = append(out, asmSet...)
	scopeElems := make([][]byte, len(e.ScopeRefs))
	for i, r := range e.ScopeRefs {
		scopeElems[i] = r.Encode()
	}
	scopeSet, err := artifact.EncodeSet(scopeElems)
	if err != nil {
		return nil, err
	}
	return append(out, scopeSet...), nil
}

// TargetChecks: FrameworkRef → own framework/1; EquationRef → equation/1;
// AssumptionRefs/ScopeRefs → owning/closure assumption/1.
func (e PhysicalEquation) TargetChecks() []physmath.TargetCheck {
	checks := []physmath.TargetCheck{
		physmath.TargetCheck{Field: "PhysicalEquation.FrameworkRef", Ref: e.FrameworkRef, Allowed: []physmath.FamilyID{FamilyFramework}},
		physmath.TargetCheck{Field: "PhysicalEquation.EquationRef", Ref: e.EquationRef, Allowed: []physmath.FamilyID{physmath.FamilyEquation}},
	}
	for i, r := range e.AssumptionRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalEquation.AssumptionRefs[%d]", i),
			Ref:     r,
			Allowed: []physmath.FamilyID{FamilyAssumption},
		})
	}
	for i, r := range e.ScopeRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalEquation.ScopeRefs[%d]", i),
			Ref:     r,
			Allowed: []physmath.FamilyID{FamilyAssumption},
		})
	}
	return checks
}

// Mint stores the canonical payload and returns its reference.
func (e PhysicalEquation) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := e.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(e.Family(), payload)
}
