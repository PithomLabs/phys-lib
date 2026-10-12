package gr

import (
	"crypto/sha256"
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// FLUX-CONSERVE-001: GR-side checked conservation inference (Option N).
//
// This is NOT a PhysMath executable rule and NOT general divergence or
// Stokes execution. It is a narrowly specified schema-conformance check
// for box-region flux-conservation claims of the K1.s4 shape: given a
// divergence relation, an integral over a domain, matching domain
// identity, cited boundary-flux lemmas, a cited stationarity condition,
// and caller-declared bulk/cap term expressions, it verifies the evidence
// chain and mints the conservation conclusion relation assembled from the
// validated terms.
//
// What it establishes: every required premise is present, resolved,
// correctly typed, mutually consistent (integral-domain digest equality),
// and covered by framework assumptions. What it does NOT establish: the
// mathematical truth of the analysis (term denotation — that the bulk
// expression denotes the integrated divergence — and lemma soundness are
// caller-asserted and recorded, never proven here). The seven PhysMath
// rules are untouched; replay contracts are unchanged.

// ErrFluxPrecondition: required flux-inference evidence missing, invalid,
// unresolved, or inconsistent. Fail closed, no artifact minted.
var ErrFluxPrecondition = fmt.Errorf("gr: flux-conservation precondition failed")

// FluxEvidence is the typed evidence bundle for FLUX-CONSERVE-001. Every
// reference is resolved and validated; nothing travels as opaque metadata.
type FluxEvidence struct {
	// BoundaryLemmaRefs: one equation/1 artifact per boundary piece,
	// each stating the flux condition for that piece (e.g. vanishing
	// wall flux). Non-empty: a claim with zero boundary evidence is
	// vacuous and rejected.
	BoundaryLemmaRefs []artifact.ArtifactRef
	// StationarityRef: the condition/1 premise eliminating the bulk term.
	StationarityRef artifact.ArtifactRef
	// BulkTermRef / CapsTermRef: caller-declared expression/1 terms for
	// the integrated-divergence side and the cap-flux-difference side.
	// Their denotation is caller-asserted (documented residual gap);
	// the conclusion is assembled from them, never from caller bytes.
	BulkTermRef artifact.ArtifactRef
	CapsTermRef artifact.ArtifactRef
	// AssumptionRefs: framework assumption/1 artifacts covering the
	// lemmas and stationarity (non-empty); FrameworkRef for membership.
	AssumptionRefs []artifact.ArtifactRef
	FrameworkRef   artifact.ArtifactRef
	// ExpectedRef, when non-nil, is a POSTCONDITION: the minted
	// conclusion must equal it, else tampering is reported. It never
	// constructs the result.
	ExpectedRef *artifact.ArtifactRef
}

// FluxCertificate is the in-memory result of a successful inference. It is
// deliberately NOT a new durable family: the minted conclusion relation
// is the durable artifact; the certificate records what was checked.
type FluxCertificate struct {
	DivergenceRef artifact.ArtifactRef
	IntegralRef   artifact.ArtifactRef
	DomainRef     artifact.ArtifactRef
	Evidence      FluxEvidence
	ConclusionRef artifact.ArtifactRef // minted relation/1 EQ(bulk, caps)
}

// resolveFamily resolves ref and requires the exact family.
func resolveFamily(store physmath.Store, what string, ref artifact.ArtifactRef, want physmath.FamilyID) ([]byte, error) {
	fam, payload, err := store.Lookup(ref)
	if err != nil {
		return nil, &Error{Op: "FluxConserve." + what, Err: fmt.Errorf("%w: unresolvable: %v", ErrFluxPrecondition, err)}
	}
	if fam != want {
		return nil, &Error{Op: "FluxConserve." + what, Err: fmt.Errorf("%w: want %s, got %s/%s",
			ErrFluxPrecondition, want.ArtifactType, fam.Namespace, fam.ArtifactType)}
	}
	return payload, nil
}

// CheckFluxConservation executes the FLUX-CONSERVE-001 schema check.
// P1 divergence relation resolves to relation/1.
// P2 integral resolves to integral/1 AND its DomainRef digest-equals domainRef.
// P3 stationarity resolves to condition/1 with a resolving relation/1 predicate.
// P4 every boundary lemma resolves to equation/1; assumptions resolve to
// assumption/1 with a resolving framework/1; at least one of each.
// P5 bulk/caps terms resolve to expression/1; conclusion EQ(bulk,caps) is
// assembled and minted; ExpectedRef, if present, must match exactly.
func CheckFluxConservation(store physmath.Store,
	divergenceRef, integralRef, domainRef artifact.ArtifactRef,
	ev FluxEvidence) (FluxCertificate, error) {
	fail := func(what string, err error) (FluxCertificate, error) {
		return FluxCertificate{}, &Error{Op: "FluxConserve." + what, Err: err}
	}
	if _, err := resolveFamily(store, "divergence", divergenceRef, physmath.FamilyRelation); err != nil {
		return fail("P1", err)
	}
	intPayload, err := resolveFamily(store, "integral", integralRef, physmath.FamilyIntegral)
	if err != nil {
		return fail("P2", err)
	}
	if _, err := resolveFamily(store, "domain", domainRef, physmath.FamilyDomain); err != nil {
		return fail("P2", err)
	}
	integral, err := physmath.DecodeIntegral(intPayload)
	if err != nil {
		return fail("P2", err)
	}
	if integral.DomainRef != domainRef {
		return fail("P2", fmt.Errorf("%w: integral domain digest mismatch", ErrFluxPrecondition))
	}
	condPayload, err := resolveFamily(store, "stationarity", ev.StationarityRef, physmath.FamilyCondition)
	if err != nil {
		return fail("P3", err)
	}
	predRef, err := artifact.DecodeArtifactRef(artifact.NewDecoder(condPayload))
	if err != nil {
		return fail("P3", err)
	}
	if _, err := resolveFamily(store, "stationarity-predicate", predRef, physmath.FamilyRelation); err != nil {
		return fail("P3", err)
	}
	if len(ev.BoundaryLemmaRefs) == 0 {
		return fail("P4", fmt.Errorf("%w: zero boundary lemmas (vacuous claim)", ErrFluxPrecondition))
	}
	for i, r := range ev.BoundaryLemmaRefs {
		if _, err := resolveFamily(store, fmt.Sprintf("boundary[%d]", i), r, physmath.FamilyEquation); err != nil {
			return fail("P4", err)
		}
	}
	if len(ev.AssumptionRefs) == 0 {
		return fail("P4", fmt.Errorf("%w: zero covering assumptions", ErrFluxPrecondition))
	}
	for i, r := range ev.AssumptionRefs {
		if _, err := resolveFamily(store, fmt.Sprintf("assumption[%d]", i), r, FamilyAssumption); err != nil {
			return fail("P4", err)
		}
	}
	if _, err := resolveFamily(store, "framework", ev.FrameworkRef, FamilyFramework); err != nil {
		return fail("P4", err)
	}
	bulkPayload, err := resolveFamily(store, "bulk-term", ev.BulkTermRef, physmath.FamilyExpression)
	if err != nil {
		return fail("P5", err)
	}
	capsPayload, err := resolveFamily(store, "caps-term", ev.CapsTermRef, physmath.FamilyExpression)
	if err != nil {
		return fail("P5", err)
	}
	bulk, err := physmath.DecodeExpr(artifact.NewDecoder(bulkPayload))
	if err != nil {
		return fail("P5", err)
	}
	caps, err := physmath.DecodeExpr(artifact.NewDecoder(capsPayload))
	if err != nil {
		return fail("P5", err)
	}
	body, err := physmath.EncodeRelationBody(physmath.OpEQ, bulk, caps)
	if err != nil {
		return fail("P5", err)
	}
	concRef, err := store.Put(physmath.FamilyRelation, body)
	if err != nil {
		return fail("P5", err)
	}
	if ev.ExpectedRef != nil && *ev.ExpectedRef != concRef {
		return fail("P5", fmt.Errorf("%w: conclusion mismatch (tampered expectation)", ErrFluxPrecondition))
	}
	return FluxCertificate{
		DivergenceRef: divergenceRef, IntegralRef: integralRef, DomainRef: domainRef,
		Evidence: ev, ConclusionRef: concRef,
	}, nil
}

// CertDigest digests the certificate's reference set for sketch citation
// and determinism comparison.
func (c FluxCertificate) CertDigest() [32]byte {
	h := sha256.New()
	refs := []artifact.ArtifactRef{c.DivergenceRef, c.IntegralRef, c.DomainRef,
		c.Evidence.StationarityRef, c.Evidence.BulkTermRef, c.Evidence.CapsTermRef,
		c.Evidence.FrameworkRef, c.ConclusionRef}
	refs = append(refs, c.Evidence.BoundaryLemmaRefs...)
	refs = append(refs, c.Evidence.AssumptionRefs...)
	for _, r := range refs {
		h.Write(r.Encode())
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}
