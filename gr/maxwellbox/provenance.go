package maxwellbox

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// RequireVolumeProvenance is the consumer-side gate for volume-density
// records (MINOR-04). A structurally valid record is never treated as
// provenance-valid on Validate alone: the referenced artifact must decode
// as volume_density/1 and pass physmath.CheckVolumeDensity, which resolves
// the determinant derivation, replays the exact CANON+SIMPLIFY path, and
// matches the diagonal-expression premise and determinant output.
//
// Guarantee limit: the check binds determinant-relevant diagonal-expression
// content, not exact metric artifact identity.
func RequireVolumeProvenance(s physmath.Store, volRef artifact.ArtifactRef) error {
	fam, payload, err := s.Lookup(volRef)
	if err != nil {
		return fmt.Errorf("maxwellbox: volume unresolvable: %w", err)
	}
	if fam != physmath.FamilyVolumeDensity {
		return fmt.Errorf("maxwellbox: volume reference is not volume_density/1")
	}
	v, err := physmath.DecodeVolumeDensity(payload)
	if err != nil {
		return fmt.Errorf("maxwellbox: volume does not decode: %w", err)
	}
	if err := physmath.CheckVolumeDensity(s, v); err != nil {
		return fmt.Errorf("maxwellbox: volume provenance invalid: %w", err)
	}
	return nil
}
