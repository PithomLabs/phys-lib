package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
)

// B3: derivation sketch records. A ProtocolSketch is an explicitly
// NON-executed ordered record of a derivation unit's steps with per-step
// statuses. It exists so unsupported reasoning (K1 s1-s4-class steps) can
// be preserved with references without claiming replay. There is
// deliberately NO Verify or Replay method on this type: adding one would
// turn a status record into a false proof. Executable fragments live in
// replayed physmath Derivations referenced from the sketch.

// Sketch statuses: closed vocabulary mirroring representation statuses.
const (
	SketchSupported   = "SUPPORTED"
	SketchUnsupported = "UNSUPPORTED"
	SketchNotApplies  = "NOT_APPLICABLE"
)

// ValidSketchStatus reports membership in the closed status vocabulary.
func ValidSketchStatus(s string) bool {
	switch s {
	case SketchSupported, SketchUnsupported, SketchNotApplies:
		return true
	}
	return false
}

// SketchStep is one ordered derivation step with its execution status and
// artifact references. A SUPPORTED step must cite the replayed derivation
// or equation artifact that executes it; an UNSUPPORTED step cites the
// representational artifacts (metric/integral/contraction) that preserve
// it without executing it.
type SketchStep struct {
	StepID      string
	Description string
	Status      string
	Refs        []artifact.ArtifactRef
}

// ProtocolSketch records one derivation unit (K1/D1 or K2/D2 shape).
type ProtocolSketch struct {
	SketchID string
	UnitID   string
	Steps    []SketchStep // order is semantic
}

// Validate enforces structure: non-empty steps, unique IDs, closed
// statuses, non-empty descriptions, reference grammar. It never executes.
func (s ProtocolSketch) Validate() error {
	if err := validateIdentifier("ProtocolSketch.SketchID", s.SketchID); err != nil {
		return err
	}
	if err := validateIdentifier("ProtocolSketch.UnitID", s.UnitID); err != nil {
		return err
	}
	if len(s.Steps) == 0 {
		return &Error{Op: "ProtocolSketch.Steps", Err: fmt.Errorf("at least one step required")}
	}
	seen := map[string]bool{}
	for i, st := range s.Steps {
		if err := validateIdentifier(fmt.Sprintf("ProtocolSketch.Steps[%d].StepID", i), st.StepID); err != nil {
			return err
		}
		if seen[st.StepID] {
			return &Error{Op: "ProtocolSketch.Steps", Err: fmt.Errorf("duplicate step %q", st.StepID)}
		}
		seen[st.StepID] = true
		if st.Description == "" {
			return &Error{Op: "ProtocolSketch.Steps", Err: fmt.Errorf("step %q needs a description", st.StepID)}
		}
		if !ValidSketchStatus(st.Status) {
			return &Error{Op: "ProtocolSketch.Steps", Err: fmt.Errorf("step %q has invalid status %q", st.StepID, st.Status)}
		}
		for j, r := range st.Refs {
			if err := artifact.ValidateNamespace(r.Namespace); err != nil {
				return &Error{Op: "ProtocolSketch.Steps", Err: fmt.Errorf("step %q ref %d: %w", st.StepID, j, err)}
			}
		}
	}
	return nil
}

// UnsupportedSteps returns the IDs of steps not executed. A sketch with
// an empty return claims full execution ONLY through cited replayed
// artifacts — callers must still verify those citations independently.
func (s ProtocolSketch) UnsupportedSteps() []string {
	var out []string
	for _, st := range s.Steps {
		if st.Status == SketchUnsupported {
			out = append(out, st.StepID)
		}
	}
	return out
}
