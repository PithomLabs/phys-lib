package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// PhysicalDefinition states what mathematical structure is taken to mean as
// a physical object in the GR framework (spec §14.3). It is distinct from
// PhysicalInterpretation (interpretive context), and the two are not merged.
// Family phys-gr/definition/1.
type PhysicalDefinition struct {
	DefinitionID       string
	FrameworkRef       artifact.ArtifactRef
	PhysicalType       string
	MathematicalRefs   []artifact.ArtifactRef // SET: any of the 16 physmath families
	AssumptionRefs     []artifact.ArtifactRef // SET: owning/closure assumption/1
	DefinitionEquation *artifact.ArtifactRef  // optional → physmath/equation/1
}

// Family returns FamilyDefinition.
func (PhysicalDefinition) Family() physmath.FamilyID { return FamilyDefinition }

// Validate checks fields (reference targets via TargetChecks + resolution).
func (d PhysicalDefinition) Validate() error {
	if err := validateIdentifier("DefinitionID", d.DefinitionID); err != nil {
		return err
	}
	if d.PhysicalType == "" {
		return &Error{Op: "Definition.PhysicalType",
			Err: fmt.Errorf("physical type is required framework-owned content")}
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (d PhysicalDefinition) Encode() ([]byte, error) {
	id, err := artifact.EncodeString(d.DefinitionID)
	if err != nil {
		return nil, err
	}
	out := append(id, d.FrameworkRef.Encode()...)
	pt, err := artifact.EncodeString(d.PhysicalType)
	if err != nil {
		return nil, err
	}
	out = append(out, pt...)
	mathElems := make([][]byte, len(d.MathematicalRefs))
	for i, r := range d.MathematicalRefs {
		mathElems[i] = r.Encode()
	}
	mathSet, err := artifact.EncodeSet(mathElems)
	if err != nil {
		return nil, err
	}
	out = append(out, mathSet...)
	asmElems := make([][]byte, len(d.AssumptionRefs))
	for i, r := range d.AssumptionRefs {
		asmElems[i] = r.Encode()
	}
	asmSet, err := artifact.EncodeSet(asmElems)
	if err != nil {
		return nil, err
	}
	out = append(out, asmSet...)
	if d.DefinitionEquation == nil {
		return append(out, artifact.EncodeOptional(false, nil)...), nil
	}
	return append(out, artifact.EncodeOptional(true, d.DefinitionEquation.Encode())...), nil
}

// TargetChecks validates FrameworkRef (own framework/1), MathematicalRefs
// (any of the 16 physmath families), AssumptionRefs (closure-checked by the
// caller via FrameworkArtifact.InClosure), and DefinitionEquation.
func (d PhysicalDefinition) TargetChecks() []physmath.TargetCheck {
	math := physmath.AllPhysMathFamilies()
	checks := []physmath.TargetCheck{
		physmath.TargetCheck{Field: "PhysicalDefinition.FrameworkRef", Ref: d.FrameworkRef, Allowed: []physmath.FamilyID{FamilyFramework}},
	}
	for i, r := range d.MathematicalRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalDefinition.MathematicalRefs[%d]", i),
			Ref:     r,
			Allowed: math,
		})
	}
	for i, r := range d.AssumptionRefs {
		checks = append(checks, physmath.TargetCheck{
			Field:   fmt.Sprintf("PhysicalDefinition.AssumptionRefs[%d]", i),
			Ref:     r,
			Allowed: []physmath.FamilyID{FamilyAssumption},
		})
	}
	if d.DefinitionEquation != nil {
		checks = append(checks, physmath.TargetCheck{
			Field:   "PhysicalDefinition.DefinitionEquation",
			Ref:     *d.DefinitionEquation,
			Allowed: []physmath.FamilyID{physmath.FamilyEquation},
		})
	}
	return checks
}

// Mint stores the canonical payload and returns its reference.
func (d PhysicalDefinition) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := d.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(d.Family(), payload)
}
