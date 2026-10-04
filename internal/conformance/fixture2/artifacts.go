package fixture2

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// Minimal family structs under the fixture2 namespace. Shapes mirror the
// specification's six families; code is intentionally not shared with
// phys-lib/gr (isolation is the point of this package).

// Assumption is a minimal fixture2 physical assumption. Category is an OPEN
// framework-owned string here — deliberately different from GR's closed
// vocabulary, proving Category is framework-owned (spec §14.2).
type Assumption struct {
	AssumptionID  string
	ConditionRefs []artifact.ArtifactRef
	Category      string
	Description   string
}

// Validate checks fields; any non-empty Category is accepted (framework-owned).
func (a Assumption) Validate() error {
	if err := validateID("AssumptionID", a.AssumptionID); err != nil {
		return err
	}
	if a.Category == "" {
		return fmt.Errorf("fixture2: Category required")
	}
	if a.Description == "" {
		return fmt.Errorf("fixture2: Description required")
	}
	seen := map[artifact.ArtifactRef]bool{}
	for _, r := range a.ConditionRefs {
		if seen[r] {
			return fmt.Errorf("fixture2: duplicate ConditionRef")
		}
		seen[r] = true
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (a Assumption) Encode() ([]byte, error) {
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

// Mint stores the canonical payload and returns its reference.
func (a Assumption) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := a.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(FamilyAssumption, payload)
}

// Definition is a minimal fixture2 physical definition.
type Definition struct {
	DefinitionID     string
	FrameworkRef     artifact.ArtifactRef
	PhysicalType     string
	MathematicalRefs []artifact.ArtifactRef
	AssumptionRefs   []artifact.ArtifactRef
}

// Validate checks fields.
func (d Definition) Validate() error {
	if err := validateID("DefinitionID", d.DefinitionID); err != nil {
		return err
	}
	if d.PhysicalType == "" {
		return fmt.Errorf("fixture2: PhysicalType required")
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (d Definition) Encode() ([]byte, error) {
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
	return append(out, asmSet...), nil
}

// Mint stores the canonical payload and returns its reference.
func (d Definition) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := d.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(FamilyDefinition, payload)
}

// Equation is a minimal fixture2 physical equation.
type Equation struct {
	EquationID     string
	FrameworkRef   artifact.ArtifactRef
	EquationRef    artifact.ArtifactRef
	AssumptionRefs []artifact.ArtifactRef
}

// Validate checks fields.
func (e Equation) Validate() error {
	if err := validateID("EquationID", e.EquationID); err != nil {
		return err
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (e Equation) Encode() ([]byte, error) {
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
	return append(out, asmSet...), nil
}

// Mint stores the canonical payload and returns its reference.
func (e Equation) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := e.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(FamilyEquation, payload)
}

// Theorem is a minimal fixture2 physical theorem.
type Theorem struct {
	TheoremID              string
	FrameworkRef           artifact.ArtifactRef
	MathematicalTheoremRef artifact.ArtifactRef
	AssumptionRefs         []artifact.ArtifactRef
	InterpretationRefs     []artifact.ArtifactRef
}

// Validate checks fields.
func (t Theorem) Validate() error {
	if err := validateID("TheoremID", t.TheoremID); err != nil {
		return err
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (t Theorem) Encode() ([]byte, error) {
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

// Mint stores the canonical payload and returns its reference.
func (t Theorem) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := t.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(FamilyTheorem, payload)
}

// Interpretation is a minimal fixture2 physical interpretation with a
// deliberately different PhysicalType over the same mathematics.
type Interpretation struct {
	InterpretationID string
	FrameworkRef     artifact.ArtifactRef
	PhysicalType     string
	MathematicalRefs []artifact.ArtifactRef
	AssumptionRefs   []artifact.ArtifactRef
}

// Validate checks fields.
func (in Interpretation) Validate() error {
	if err := validateID("InterpretationID", in.InterpretationID); err != nil {
		return err
	}
	if in.PhysicalType == "" {
		return fmt.Errorf("fixture2: PhysicalType required")
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (in Interpretation) Encode() ([]byte, error) {
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

// Mint stores the canonical payload and returns its reference.
func (in Interpretation) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := in.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(FamilyInterpretation, payload)
}
