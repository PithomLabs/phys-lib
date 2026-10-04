package equivariance

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// DecodeDerivationPayload decodes a physmath/derivation/1 canonical payload:
// SET(premises) || SEQUENCE(steps) || SET(results), steps in declaration order.
func DecodeDerivationPayload(payload []byte) (physmath.Derivation, error) {
	d := artifact.NewDecoder(payload)
	n, err := d.SequenceCount()
	if err != nil {
		return physmath.Derivation{}, err
	}
	pre := make([]artifact.ArtifactRef, 0, n)
	for i := 0; i < n; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return physmath.Derivation{}, err
		}
		pre = append(pre, r)
	}
	m, err := d.SequenceCount()
	if err != nil {
		return physmath.Derivation{}, err
	}
	steps := make([]physmath.DerivationStep, 0, m)
	for i := 0; i < m; i++ {
		s, err := decodeStep(d)
		if err != nil {
			return physmath.Derivation{}, err
		}
		steps = append(steps, s)
	}
	k, err := d.SequenceCount()
	if err != nil {
		return physmath.Derivation{}, err
	}
	res := make([]artifact.ArtifactRef, 0, k)
	for i := 0; i < k; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return physmath.Derivation{}, err
		}
		res = append(res, r)
	}
	if !d.Exhausted() {
		return physmath.Derivation{}, fmt.Errorf("equivariance: trailing bytes in derivation")
	}
	return physmath.Derivation{PremiseRefs: pre, Steps: steps, ResultRefs: res}, nil
}

func decodeStep(d *artifact.Decoder) (physmath.DerivationStep, error) {
	var s physmath.DerivationStep
	ord, err := d.UInt32()
	if err != nil {
		return s, err
	}
	s.Ordinal = ord
	if s.ProducerNamespace, err = d.String(); err != nil {
		return s, err
	}
	if s.RuleID, err = d.String(); err != nil {
		return s, err
	}
	n, err := d.SequenceCount()
	if err != nil {
		return s, err
	}
	for i := 0; i < n; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return s, err
		}
		s.InputRefs = append(s.InputRefs, r)
	}
	if s.OutputRef, err = artifact.DecodeArtifactRef(d); err != nil {
		return s, err
	}
	m, err := d.SequenceCount()
	if err != nil {
		return s, err
	}
	for i := 0; i < m; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return s, err
		}
		s.ConditionRefs = append(s.ConditionRefs, r)
	}
	params, err := d.Bytes()
	if err != nil {
		return s, err
	}
	s.Parameters = params
	return s, nil
}

// DecodeTheoremPayload decodes a physmath/theorem/1 canonical payload:
// SET(premises) || ConclusionRef || SET(assumptions) || DerivationRef.
func DecodeTheoremPayload(payload []byte) (physmath.Theorem, error) {
	d := artifact.NewDecoder(payload)
	var th physmath.Theorem
	n, err := d.SequenceCount()
	if err != nil {
		return th, err
	}
	for i := 0; i < n; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return th, err
		}
		th.PremiseRefs = append(th.PremiseRefs, r)
	}
	if th.ConclusionRef, err = artifact.DecodeArtifactRef(d); err != nil {
		return th, err
	}
	m, err := d.SequenceCount()
	if err != nil {
		return th, err
	}
	for i := 0; i < m; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return th, err
		}
		th.AssumptionRefs = append(th.AssumptionRefs, r)
	}
	if th.DerivationRef, err = artifact.DecodeArtifactRef(d); err != nil {
		return th, err
	}
	if !d.Exhausted() {
		return th, fmt.Errorf("equivariance: trailing bytes in theorem")
	}
	return th, nil
}
