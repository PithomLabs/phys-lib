package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
)

// R1: coordinate-map record. A TransformMap binds a named source chart to
// a named target chart with ordered component expressions and the metric
// artifacts on both sides. It validates references and shape ONLY: it
// never executes the transformation, never evaluates component functions,
// and never proves the target metric follows from the source. A fully
// validated map is still schematic where its components are opaque
// (hyperbolic or otherwise unprojectable) nodes.

// TransformMap is GR-local data (no new durable family, mirroring the
// ProtocolSketch precedent): validated structure for derivation records,
// never an execution vehicle.
type TransformMap struct {
	MapID          string
	SourceChart    string // e.g. "inertial_minkowski"
	TargetChart    string // e.g. "rindler_accelerated"
	Components     []GRExpr
	SourceMetric   artifact.ArtifactRef   // → symmetric_bilinear/1
	TargetMetric   artifact.ArtifactRef   // → symmetric_bilinear/1
	AssumptionRefs []artifact.ArtifactRef // SET: chart/domain assumptions
}

// Validate enforces identifier grammar, chart names, non-empty components,
// and reference grammar. Component semantics (what the functions
// mean) are NOT validated: opaque nodes are legal content.
func (m TransformMap) Validate() error {
	if err := validateIdentifier("TransformMap.MapID", m.MapID); err != nil {
		return err
	}
	if m.SourceChart == "" || m.TargetChart == "" {
		return &Error{Op: "TransformMap.Chart", Err: fmt.Errorf("source and target charts required")}
	}
	if len(m.Components) == 0 {
		return &Error{Op: "TransformMap.Components", Err: fmt.Errorf("at least one component required")}
	}
	fields := map[string]artifact.ArtifactRef{
		"TransformMap.SourceMetric": m.SourceMetric, "TransformMap.TargetMetric": m.TargetMetric,
	}
	for i, r := range m.AssumptionRefs {
		fields[fmt.Sprintf("TransformMap.AssumptionRefs[%d]", i)] = r
	}
	for k, r := range fields {
		if err := artifact.ValidateNamespace(r.Namespace); err != nil {
			return &Error{Op: "TransformMap", Err: fmt.Errorf("%s: %w", k, err)}
		}
	}
	seen := map[artifact.ArtifactRef]bool{}
	for _, r := range m.AssumptionRefs {
		if seen[r] {
			return &Error{Op: "TransformMap.AssumptionRefs", Err: fmt.Errorf("duplicate reference")}
		}
		seen[r] = true
	}
	return nil
}

// OpaqueComponents returns the indices of components that cannot project
// to PhysMath (hyperbolic, trigonometric, or unknown nodes anywhere in
// the component tree). A map with opaque components is schematic by
// construction: representable, never executable.
func (m TransformMap) OpaqueComponents() []int {
	var out []int
	for i, c := range m.Components {
		if containsOpaque(c) {
			out = append(out, i)
		}
	}
	return out
}

func containsOpaque(g GRExpr) bool {
	switch g.Kind {
	case GSin, GCos, GSinh, GCosh, GUnknown:
		return true
	}
	for _, c := range g.Children {
		if containsOpaque(c) {
			return true
		}
	}
	if g.Operand != nil && containsOpaque(*g.Operand) {
		return true
	}
	if g.Base != nil && containsOpaque(*g.Base) {
		return true
	}
	for _, a := range g.FuncArgs {
		if containsOpaque(a) {
			return true
		}
	}
	return false
}
