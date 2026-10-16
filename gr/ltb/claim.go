package ltb

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
)

// Scientific statuses used by the manifest metadata. These are strings in
// durable metadata, not a state machine: no artifact schema enforces
// transitions, and Verify does not prove or enforce them. The function
// below is a metadata/evidence consistency check, nothing more.
const (
	StatusLiteratureClaim = "literature-claim-requiring-verification"
	StatusCandidate       = "candidate-construction-incomplete"
	StatusOpenConjecture  = "open-conjecture"
	StatusDeferred        = "deferred"
	StatusProved          = "proved"
)

// ClaimStatus pins one research claim to its manifest status plus the
// evidence refs that status cites (derivation/1 refs for executed
// evidence; empty where the status cites none).
type ClaimStatus struct {
	ID       string
	Status   string
	Evidence []artifact.ArtifactRef
}

// CheckClaimPromotion rejects a status that cites no valid evidence
// path. In particular, marking ZEH-Q-03 proved without a derivation
// artifact in the fixture evidence inventory fails. A nil return means
// the metadata is consistent, not that the science is settled.
func CheckClaimPromotion(claims []ClaimStatus, evidence []artifact.ArtifactRef) error {
	known := map[artifact.ArtifactRef]bool{}
	for _, r := range evidence {
		known[r] = true
	}
	for _, c := range claims {
		switch c.Status {
		case StatusLiteratureClaim, StatusCandidate, StatusOpenConjecture, StatusDeferred:
			continue
		case StatusProved:
			if len(c.Evidence) == 0 {
				return fmt.Errorf("ltb: claim %s marked proved with no evidence path", c.ID)
			}
			for _, r := range c.Evidence {
				if !known[r] {
					return fmt.Errorf("ltb: claim %s cites evidence outside the fixture inventory", c.ID)
				}
			}
		default:
			return fmt.Errorf("ltb: claim %s has unknown status %q", c.ID, c.Status)
		}
	}
	return nil
}
