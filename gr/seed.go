package gr

import (
	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// SeedConventions mints the Phase-1 GR convention assumptions: metric
// signature, curvature convention, and cosmological-constant choice. These
// are descriptive framework-owned assumptions with EMPTY ConditionRefs —
// conventions are not PhysMath predicates (plan9.3 D15). They exist so the
// equivariance fixture and framework validation have explicit assumption
// artifacts to reference instead of code comments.
func SeedConventions(store physmath.Store) (map[string]artifact.ArtifactRef, error) {
	seeds := []PhysicalAssumption{
		{
			AssumptionID: "metric_signature",
			Category:     CategoryConvention,
			Description:  "GR metric signature convention -+++.",
		},
		{
			AssumptionID: "curvature_convention",
			Category:     CategoryConvention,
			Description:  "GR curvature convention MTW (R^rho_sigmamu_nu sign).",
		},
		{
			AssumptionID: "cosmological_constant",
			Category:     CategoryConvention,
			Description:  "GR cosmological constant choice Lambda = 0 for the vacuum regime.",
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
