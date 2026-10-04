package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// PhysicalTheorem attaches framework meaning to a mathematical theorem: the
// mathematical theorem and its derivation remain owned by PhysMath
// (spec §14.5). Family phys-gr/theorem/1.
type PhysicalTheorem struct {
	TheoremID              string
	FrameworkRef           artifact.ArtifactRef
	MathematicalTheoremRef artifact.ArtifactRef
	AssumptionRefs         []artifact.ArtifactRef // SET: owning/closure assumption/1
	InterpretationRefs     []artifact.ArtifactRef // SET: owning/closure interpretation/1
}

// Family returns FamilyTheorem.
func (PhysicalTheorem) Family() physmath.FamilyID { return FamilyTheorem }

// Validate checks fields (reference targets via TargetChecks + resolution).
func (t PhysicalTheorem) Validate() error {
	if err := validateIdentifier("TheoremID", t.TheoremID); err != nil {
		return err
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (t PhysicalTheorem) Encode() ([]byte, error) {
	id, err := artifact.EncodeString(t.TheoremID)
	if err != nil {
		return nil, err
	}
	out := append(id, t.FrameworkRef.Encode()...)
	out = append(out, t.MathematicalTheoremRef.Encode()...)
	asmElems := make([][]byte, len(t.AssumptionRefs))
	for i, r := range t.AssumptionRefs {
		asmElems[i] = r.Encode()
	}
	asmSet, err := artifact.EncodeSet(asmElems)
	if err != nil {
		return nil, err
	}
	out = append(out, asmSet...)
	intElems := make([][]byte, len(t.InterpretationRefs))
	for i, r := range t.InterpretationRefs {
		intElems[i] = r.Encode()
	}
	intSet, err := artifact.EncodeSet(intElems)
	if err != nil {
		return nil, err
	}
	return append(out, intSet...), nil
}

// TargetChecks: FrameworkRef → own framework/1; MathematicalTheoremRef →
// theorem/1; AssumptionRefs → owning/closure assumption/1;
// InterpretationRefs → owning/closure interpretation/1.
func (t PhysicalTheorem) TargetChecks() []physmath.TargetCheck {
	checks := []physmath.TargetCheck{
		physmath.TargetCheck{Field: "PhysicalTheorem.FrameworkRef", Ref: t.FrameworkRef, Allowed: []physmath.FamilyID{FamilyFramework}},
		physmath.TargetCheck{Field: "PhysicalTheorem.MathematicalTheoremRef", Ref: t.MathematicalTheoremRef,
			Allowed: []physmath.FamilyID{physmath.FamilyTheorem}},
	}
	for i, r := range t.AssumptionRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalTheorem.AssumptionRefs[%d]", i),
			Ref:     r,
			Allowed: []physmath.FamilyID{FamilyAssumption},
		})
	}
	for i, r := range t.InterpretationRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalTheorem.InterpretationRefs[%d]", i),
			Ref:     r,
			Allowed: []physmath.FamilyID{FamilyInterpretation},
		})
	}
	return checks
}

// Mint stores the canonical payload and returns its reference.
func (t PhysicalTheorem) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := t.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(t.Family(), payload)
}
