package maxwellbox

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// buildContractionEvidence mints the K1 s2 current-definition evidence:
// honest schematic operands (distinct display-symbol components shared over
// one resolved index space), the evaluated w_a = T_ab xi^b tensor with its
// contraction_evaluation/1 certificate, and the fixture-side negated
// j_a = -w_a tensor with per-component Neg+CANON derivations. It panics on
// programmer error, matching Build's convention. The energy integral stays
// unexecuted: s2 remains UNSUPPORTED.
func (f *Fixture) buildContractionEvidence(s physmath.Store) {
	desig, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol("lab_space"))
	if err != nil {
		panic(err)
	}
	spacePayload, err := physmath.Collection{ElementRefs: []artifact.ArtifactRef{desig}}.Encode()
	if err != nil {
		panic(err)
	}
	f.LabSpace, err = s.Put(physmath.FamilyCollection, spacePayload)
	if err != nil {
		panic(err)
	}
	mkIdx := func(label string, v physmath.IndexVariance) artifact.ArtifactRef {
		p, err := physmath.Index{Label: label, Variance: v, Space: &f.LabSpace}.Encode()
		if err != nil {
			panic(err)
		}
		r, err := s.Put(physmath.FamilyIndex, p)
		if err != nil {
			panic(err)
		}
		return r
	}
	ia, ib, ix := mkIdx("a", physmath.VarianceCovariant),
		mkIdx("b", physmath.VarianceCovariant),
		mkIdx("b", physmath.VarianceContravariant)
	mkTensor := func(rank int, dim int, idx []artifact.ArtifactRef, comps []physmath.Expr) artifact.ArtifactRef {
		crefs := make([]artifact.ArtifactRef, len(comps))
		for i, e := range comps {
			r, err := physmath.PutExpression(s, e)
			if err != nil {
				panic(err)
			}
			crefs[i] = r
		}
		tn := physmath.Tensor{Rank: uint16(rank), Dimension: uint16(dim), IndexRefs: idx, ComponentRefs: crefs}
		if err := tn.Validate(); err != nil {
			panic(err)
		}
		r, err := s.Put(physmath.FamilyTensor, tn.Encode())
		if err != nil {
			panic(err)
		}
		return r
	}
	f.HonestTRef = mkTensor(2, 2, []artifact.ArtifactRef{ia, ib}, []physmath.Expr{
		physmath.NewEphemeralSymbol("T00"), physmath.NewEphemeralSymbol("T01"),
		physmath.NewEphemeralSymbol("T10"), physmath.NewEphemeralSymbol("T11"),
	})
	f.XiRef = mkTensor(1, 2, []artifact.ArtifactRef{ix}, []physmath.Expr{
		physmath.NewEphemeralSymbol("xi0"), physmath.NewEphemeralSymbol("xi1"),
	})
	wDeriv, err := func() (physmath.Derivation, error) {
		w, d, err := physmath.EvaluateContraction(s, f.HonestTRef, f.XiRef, 1)
		if err != nil {
			return physmath.Derivation{}, err
		}
		f.WRef = w
		return d, nil
	}()
	if err != nil {
		panic(err)
	}
	f.WDerivRef = mustMintDerivation(s, wDeriv)
	weval := physmath.ContractionEvaluation{TensorRef: f.HonestTRef, VectorRef: f.XiRef,
		ContractedTensorSlot: 1, ResultTensorRef: f.WRef, DerivationRef: f.WDerivRef}
	if err := weval.Validate(); err != nil {
		panic(err)
	}
	if err := physmath.CheckContractionEvaluation(s, weval); err != nil {
		panic(err)
	}
	f.WEvalRef = mustMintEvaluation(s, weval)
	// Fixture-side negation j_a = -w_a: per-component Neg construction +
	// CANON replay, linked explicitly. Not part of the evaluation record.
	wfam, wpayload, err := s.Lookup(f.WRef)
	if err != nil {
		panic(err)
	}
	if wfam != physmath.FamilyTensor {
		panic("w reference is not tensor/1")
	}
	wt, err := physmath.DecodeTensor(wpayload)
	if err != nil {
		panic(err)
	}
	f.JDerivRefs = make([]artifact.ArtifactRef, 0, len(wt.ComponentRefs))
	jcomps := make([]artifact.ArtifactRef, 0, len(wt.ComponentRefs))
	for _, wref := range wt.ComponentRefs {
		wfam, wpay, err := s.Lookup(wref)
		if err != nil {
			panic(err)
		}
		if wfam != physmath.FamilyExpression {
			panic("w component is not expression/1")
		}
		we, err := physmath.DecodeExpr(artifact.NewDecoder(wpay))
		if err != nil {
			panic(err)
		}
		negRef, err := physmath.PutExpression(s, physmath.Neg(we))
		if err != nil {
			panic(err)
		}
		canon, err := physmath.Apply(s, physmath.Invocation{
			ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{negRef},
		})
		if err != nil {
			panic(err)
		}
		jd := physmath.Derivation{
			PremiseRefs: []artifact.ArtifactRef{negRef},
			Steps: []physmath.DerivationStep{{
				Ordinal: 0, ProducerNamespace: physmath.NamespacePhysMath, RuleID: physmath.RuleExprCanon,
				InputRefs: []artifact.ArtifactRef{negRef}, OutputRef: canon.OutputRef,
			}},
			ResultRefs: []artifact.ArtifactRef{canon.OutputRef},
		}
		if err := physmath.Replay(s, jd, nil); err != nil {
			panic(err)
		}
		f.JDerivRefs = append(f.JDerivRefs, mustMintDerivation(s, jd))
		jcomps = append(jcomps, canon.OutputRef)
	}
	jt := physmath.Tensor{Rank: 1, Dimension: wt.Dimension, IndexRefs: wt.IndexRefs, ComponentRefs: jcomps}
	if err := jt.Validate(); err != nil {
		panic(err)
	}
	f.JRef, err = s.Put(physmath.FamilyTensor, jt.Encode())
	if err != nil {
		panic(err)
	}
	if err := RequireNegatedCurrent(s, f.WRef, f.JRef, f.JDerivRefs); err != nil {
		panic(err)
	}
}

func mustMintEvaluation(s physmath.Store, c physmath.ContractionEvaluation) artifact.ArtifactRef {
	r, err := s.Put(physmath.FamilyContractionEvaluation, c.Encode())
	if err != nil {
		panic(err)
	}
	return r
}

// RequireContractionEvaluation is the consumer-side gate for evaluated
// contractions: the reference must decode as contraction_evaluation/1 and
// pass physmath.CheckContractionEvaluation (operand gates, flat-derivation
// shape and indexed linkage, replay, ordered component membership).
func RequireContractionEvaluation(s physmath.Store, evalRef artifact.ArtifactRef) error {
	fam, payload, err := s.Lookup(evalRef)
	if err != nil {
		return fmt.Errorf("maxwellbox: evaluation unresolvable: %w", err)
	}
	if fam != physmath.FamilyContractionEvaluation {
		return fmt.Errorf("maxwellbox: evaluation reference is not contraction_evaluation/1")
	}
	c, err := physmath.DecodeContractionEvaluation(payload)
	if err != nil {
		return fmt.Errorf("maxwellbox: evaluation does not decode: %w", err)
	}
	if err := physmath.CheckContractionEvaluation(s, c); err != nil {
		return fmt.Errorf("maxwellbox: evaluation invalid: %w", err)
	}
	return nil
}

// RequireNegatedCurrent is the consumer-side gate for the fixture-side
// negation j_a = -w_a. It verifies, per component in tensor order: the
// derivation replays a single CANON step whose premise is exactly
// Neg(w_i) by construction and whose output is exactly j_i, with matching
// rank, dimension, and free index. Negation evidence stays separate from
// the contraction-evaluation record by design.
func RequireNegatedCurrent(s physmath.Store, wRef, jRef artifact.ArtifactRef, derivRefs []artifact.ArtifactRef) error {
	wfam, wpayload, err := s.Lookup(wRef)
	if err != nil {
		return fmt.Errorf("maxwellbox: w tensor unresolvable: %w", err)
	}
	if wfam != physmath.FamilyTensor {
		return fmt.Errorf("maxwellbox: w reference is not tensor/1")
	}
	jfam, jpayload, err := s.Lookup(jRef)
	if err != nil {
		return fmt.Errorf("maxwellbox: j tensor unresolvable: %w", err)
	}
	if jfam != physmath.FamilyTensor {
		return fmt.Errorf("maxwellbox: j reference is not tensor/1")
	}
	wt, err := physmath.DecodeTensor(wpayload)
	if err != nil {
		return fmt.Errorf("maxwellbox: w tensor undecodable: %w", err)
	}
	jt, err := physmath.DecodeTensor(jpayload)
	if err != nil {
		return fmt.Errorf("maxwellbox: j tensor undecodable: %w", err)
	}
	if wt.Rank != 1 || jt.Rank != 1 {
		return fmt.Errorf("maxwellbox: w/j tensors must be rank 1")
	}
	if wt.Dimension != jt.Dimension {
		return fmt.Errorf("maxwellbox: w/j dimension mismatch")
	}
	if len(wt.IndexRefs) != 1 || len(jt.IndexRefs) != 1 || wt.IndexRefs[0] != jt.IndexRefs[0] {
		return fmt.Errorf("maxwellbox: j must preserve the w free index")
	}
	if len(derivRefs) != int(wt.Dimension) {
		return fmt.Errorf("maxwellbox: negation needs one derivation per component")
	}
	for i := range wt.ComponentRefs {
		dfam, dpayload, err := s.Lookup(derivRefs[i])
		if err != nil {
			return fmt.Errorf("maxwellbox: negation derivation %d unresolvable: %w", i, err)
		}
		if dfam != physmath.FamilyDerivation {
			return fmt.Errorf("maxwellbox: negation derivation %d is not derivation/1", i)
		}
		dd, err := physmath.DecodeDerivation(dpayload)
		if err != nil {
			return fmt.Errorf("maxwellbox: negation derivation %d undecodable: %w", i, err)
		}
		if len(dd.Steps) != 1 {
			return fmt.Errorf("maxwellbox: negation derivation %d must have exactly one step", i)
		}
		st := dd.Steps[0]
		if st.Ordinal != 0 || st.ProducerNamespace != physmath.NamespacePhysMath || st.RuleID != physmath.RuleExprCanon {
			return fmt.Errorf("maxwellbox: negation derivation %d must be a single CANON step", i)
		}
		if len(st.ConditionRefs) != 0 || len(st.Parameters) != 0 {
			return fmt.Errorf("maxwellbox: negation derivation %d must carry no conditions or parameters", i)
		}
		if len(dd.PremiseRefs) != 1 || len(st.InputRefs) != 1 || st.InputRefs[0] != dd.PremiseRefs[0] {
			return fmt.Errorf("maxwellbox: negation derivation %d must consume its premise", i)
		}
		if len(dd.ResultRefs) != 1 || st.OutputRef != dd.ResultRefs[0] || st.OutputRef != jt.ComponentRefs[i] {
			return fmt.Errorf("maxwellbox: negation derivation %d must produce j component %d", i, i)
		}
		wexpr, err := resolveExprRef(s, "w", wt.ComponentRefs[i])
		if err != nil {
			return err
		}
		premExpr, err := resolveExprRef(s, "negation premise", dd.PremiseRefs[0])
		if err != nil {
			return err
		}
		wantPremise, err := physmath.PutExpression(s, physmath.Neg(wexpr))
		if err != nil {
			return fmt.Errorf("maxwellbox: negation premise %d unmintable: %w", i, err)
		}
		if wantPremise != dd.PremiseRefs[0] {
			return fmt.Errorf("maxwellbox: negation premise %d is not Neg(w_%d)", i, i)
		}
		_ = premExpr
		if err := physmath.Replay(s, dd, nil); err != nil {
			return fmt.Errorf("maxwellbox: negation derivation %d does not replay: %w", i, err)
		}
	}
	return nil
}

func resolveExprRef(s physmath.Store, what string, ref artifact.ArtifactRef) (physmath.Expr, error) {
	fam, payload, err := s.Lookup(ref)
	if err != nil {
		return physmath.Expr{}, fmt.Errorf("maxwellbox: %s unresolvable: %w", what, err)
	}
	if fam != physmath.FamilyExpression {
		return physmath.Expr{}, fmt.Errorf("maxwellbox: %s is not expression/1", what)
	}
	e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
	if err != nil {
		return physmath.Expr{}, fmt.Errorf("maxwellbox: %s undecodable: %w", what, err)
	}
	return e, nil
}
