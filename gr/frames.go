package gr

import (
	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// B1: frame and chart artifacts. A frame is a PhysicalDefinition binding
// framework meaning (inertial/Rindler observer family) to mathematical
// references (symmetric-bilinear metric artifacts, coordinate labels as
// defining expressions). Coordinate charts and metric conventions are
// framework-owned assumptions, never PhysMath content.

// NewFrameDefinition builds a validated frame definition. mathRefs must
// resolve to PhysMath artifacts (checked via TargetChecks by the caller);
// assumptionRefs must be owning/closure assumptions.
func NewFrameDefinition(frameworkRef artifact.ArtifactRef, definitionID, physicalType string,
	mathRefs, assumptionRefs []artifact.ArtifactRef,
	definitionEquation *artifact.ArtifactRef) (PhysicalDefinition, error) {
	d := PhysicalDefinition{
		DefinitionID:       definitionID,
		FrameworkRef:       frameworkRef,
		PhysicalType:       physicalType,
		MathematicalRefs:   mathRefs,
		AssumptionRefs:     assumptionRefs,
		DefinitionEquation: definitionEquation,
	}
	if err := d.Validate(); err != nil {
		return PhysicalDefinition{}, err
	}
	return d, nil
}

// NewChartAssumption records a coordinate-chart relationship (e.g. Rindler
// coordinates (6) relating inertial and accelerated descriptions) as a
// model-category assumption. The chart mathematics lives in PhysMath
// references held by frame definitions; this assumption records that the
// framework adopts the chart, with its source location.
func NewChartAssumption(assumptionID, description string, conditionRefs []artifact.ArtifactRef) (PhysicalAssumption, error) {
	a := PhysicalAssumption{
		AssumptionID:  assumptionID,
		ConditionRefs: conditionRefs,
		Category:      CategoryModel,
		Description:   description,
	}
	if err := a.Validate(); err != nil {
		return PhysicalAssumption{}, err
	}
	return a, nil
}

// SeedMaxwellBoxAssumptions mints the K1 author-stated assumptions:
// stationarity (regularity), perfectly conducting walls (regularity), and
// the Rindler chart adoption (model). Descriptions quote the paper's
// requirements; ConditionRefs stay empty because these are framework
// regularity/model adoptions, not PhysMath predicates (seed.go pattern).
func SeedMaxwellBoxAssumptions(store physmath.Store) (map[string]artifact.ArtifactRef, error) {
	seeds := []PhysicalAssumption{
		{
			AssumptionID: "stationarity_zero_momentum",
			Category:     CategoryRegularity,
			Description:  "K1 stationarity (23): total spatial momentum of confined radiation vanishes in the box rest frame.",
		},
		{
			AssumptionID: "perfectly_conducting_walls",
			Category:     CategoryRegularity,
			Description:  "K1 boundary (A28/A29): perfectly conducting walls; field vanishes in conductor; wall flux condition holds.",
		},
		{
			AssumptionID: "rindler_chart",
			Category:     CategoryModel,
			Description:  "K1 geometric setup: Rindler coordinates (6)-(7) describe uniformly accelerated box observers; small-box regime |a|L/c^2 << 1 for a single assignable acceleration.",
		},
	}
	out := make(map[string]artifact.ArtifactRef, len(seeds))
	for _, a := range seeds {
		if err := a.Validate(); err != nil {
			return nil, err
		}
		ref, err := a.Mint(store)
		if err != nil {
			return nil, err
		}
		out[a.AssumptionID] = ref
	}
	return out, nil
}
