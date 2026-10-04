package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// PhysicalInterpretation is the explicit framework-specific interpretive
// context in which PhysicalDefinition meanings are assigned (spec §14.6).
// There is no global interpretation registry. Family phys-gr/interpretation/1.
type PhysicalInterpretation struct {
	InterpretationID string
	FrameworkRef     artifact.ArtifactRef
	PhysicalType     string
	MathematicalRefs []artifact.ArtifactRef // SET: any of the 16 physmath families
	AssumptionRefs   []artifact.ArtifactRef // SET: owning/closure assumption/1
}

// Family returns FamilyInterpretation.
func (PhysicalInterpretation) Family() physmath.FamilyID { return FamilyInterpretation }

// Validate checks fields (reference targets via TargetChecks + resolution).
func (in PhysicalInterpretation) Validate() error {
	if err := validateIdentifier("InterpretationID", in.InterpretationID); err != nil {
		return err
	}
	if in.PhysicalType == "" {
		return &Error{Op: "Interpretation.PhysicalType",
			Err: fmt.Errorf("physical type is required framework-owned content")}
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (in PhysicalInterpretation) Encode() ([]byte, error) {
	id, err := artifact.EncodeString(in.InterpretationID)
	if err != nil {
		return nil, err
	}
	out := append(id, in.FrameworkRef.Encode()...)
	pt, err := artifact.EncodeString(in.PhysicalType)
	if err != nil {
		return nil, err
	}
	out = append(out, pt...)
	mathElems := make([][]byte, len(in.MathematicalRefs))
	for i, r := range in.MathematicalRefs {
		mathElems[i] = r.Encode()
	}
	mathSet, err := artifact.EncodeSet(mathElems)
	if err != nil {
		return nil, err
	}
	out = append(out, mathSet...)
	asmElems := make([][]byte, len(in.AssumptionRefs))
	for i, r := range in.AssumptionRefs {
		asmElems[i] = r.Encode()
	}
	asmSet, err := artifact.EncodeSet(asmElems)
	if err != nil {
		return nil, err
	}
	return append(out, asmSet...), nil
}

// TargetChecks: FrameworkRef → own framework/1; MathematicalRefs → any of
// the 16 physmath families; AssumptionRefs → owning/closure assumption/1.
func (in PhysicalInterpretation) TargetChecks() []physmath.TargetCheck {
	math := physmath.AllPhysMathFamilies()
	checks := []physmath.TargetCheck{
		physmath.TargetCheck{Field: "PhysicalInterpretation.FrameworkRef", Ref: in.FrameworkRef, Allowed: []physmath.FamilyID{FamilyFramework}},
	}
	for i, r := range in.MathematicalRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalInterpretation.MathematicalRefs[%d]", i),
			Ref:     r,
			Allowed: math,
		})
	}
	for i, r := range in.AssumptionRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalInterpretation.AssumptionRefs[%d]", i),
			Ref:     r,
			Allowed: []physmath.FamilyID{FamilyAssumption},
		})
	}
	return checks
}

// Mint stores the canonical payload and returns its reference.
func (in PhysicalInterpretation) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := in.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(in.Family(), payload)
}
