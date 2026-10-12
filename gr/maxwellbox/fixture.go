// Package maxwellbox is the K1 (Maxwell mass-energy) derivation-trace
// attempt for rev2 (digest 000f6ead…5280). It exercises the real
// PhysMath/PhysLib interfaces on feature branches and reports honestly:
//
//   - s1/s4-class steps (geometry, currents, integrals, flux) are
//     REPRESENTED as validated artifacts (metric, domain, integral,
//     contraction, sketch) and NEVER executed.
//   - s5/s6-class endpoint algebra is EXECUTED through replayed
//     derivations with theorem-conclusion gates.
//   - The ProtocolSketch records per-step statuses; execution coverage
//     is PARTIAL by construction. A fully green suite here proves the
//     fragment replays and the unsupported remainder is preserved —
//     never that K1 is verified end to end.
package maxwellbox

import (
	"crypto/sha256"
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-lib/gr"
	"github.com/PithomLabs/phys-math"
)

// Fixture holds every minted artifact of the K1 trace attempt.
type Fixture struct {
	Store             physmath.Store
	FrameworkRef      artifact.ArtifactRef
	Conventions       map[string]artifact.ArtifactRef
	Assumptions       map[string]artifact.ArtifactRef
	MinkowskiRef      artifact.ArtifactRef   // symmetric_bilinear/1 diag(-1,1,1,1)
	DomainRef         artifact.ArtifactRef   // domain/1 Sigma_tau
	IntegralRef       artifact.ArtifactRef   // integral/1 box energy
	TensorRef         artifact.ArtifactRef   // tensor/1 schematic T^a_b
	ContractionRef    artifact.ArtifactRef   // contraction/1 certificate
	TransformDeriv    physmath.Derivation    // replayed s5-class fragment (30)
	TransformThm      artifact.ArtifactRef   // math theorem/1 for (30)
	TransformPhysThm  artifact.ArtifactRef   // phys-gr theorem/1 for (30)
	MassDeriv         physmath.Derivation    // replayed s6-class fragment (32)
	MassThm           artifact.ArtifactRef   // math theorem/1 for (32)
	Equation30        artifact.ArtifactRef   // phys-gr equation/1
	Equation32        artifact.ArtifactRef   // phys-gr equation/1
	Sketch            gr.ProtocolSketch      // s1-s6 statuses (s4 SUPPORTED via flux inference)
	FluxCert          gr.FluxCertificate     // s4 FLUX-CONSERVE-001 result
	RindlerRef        artifact.ArtifactRef   // symmetric_bilinear/1 diag(-xbar^2,1,1,1), s1 schematic
	MinkowskiDet      artifact.ArtifactRef   // expression/1 exactly -1, replayed
	MinkowskiDetDeriv artifact.ArtifactRef   // derivation/1 exact CANON+SIMPLIFY evidence for MinkowskiDet
	RindlerDet        artifact.ArtifactRef   // expression/1 canonical product, replayed
	RindlerDetDeriv   artifact.ArtifactRef   // derivation/1 exact CANON+SIMPLIFY evidence for RindlerDet
	MinkowskiVol      artifact.ArtifactRef   // volume_density/1 factor 1, no reduction
	RindlerVol        artifact.ArtifactRef   // volume_density/1 factor |xbar|, unreduced
	RindlerVolRed     artifact.ArtifactRef   // volume_density/1 reduced to xbar under xbar>0
	LabSpace          artifact.ArtifactRef   // collection/1 shared index space for T/xi slots
	HonestTRef        artifact.ArtifactRef   // tensor/1 rank-2 T_ab, distinct display-symbol components
	XiRef             artifact.ArtifactRef   // tensor/1 rank-1 xi^b lab vector, display-symbol components
	WRef              artifact.ArtifactRef   // tensor/1 rank-1 w_a = T_ab xi^b, evaluated
	WEvalRef          artifact.ArtifactRef   // contraction_evaluation/1 certificate for w_a
	WDerivRef         artifact.ArtifactRef   // derivation/1 flat CANON+SIMPLIFY evidence for w_a
	JRef              artifact.ArtifactRef   // tensor/1 rank-1 j_a = -w_a, negated fixture-side
	JDerivRefs        []artifact.ArtifactRef // derivation/1 per-component Neg+CANON evidence for j_a
	Symbols           map[string]artifact.ArtifactRef
}

// sym mints an ephemeral display symbol expression.
func sym(s physmath.Store, name string) artifact.ArtifactRef {
	ref, err := physmath.PutExpression(s, physmath.NewEphemeralSymbol(name))
	if err != nil {
		panic(err)
	}
	return ref
}

// rat mints an exact rational expression.
func rat(s physmath.Store, n, d int64) artifact.ArtifactRef {
	ref, err := physmath.PutExpression(s, physmath.MustRational(n, d))
	if err != nil {
		panic(err)
	}
	return ref
}

// Build constructs the full fixture. It panics on programmer error;
// validation failures of crafted artifacts are returned as errors by the
// protocol test through Verify.
func Build() *Fixture {
	s := physmath.NewMemStore()
	f := &Fixture{Store: s, Symbols: map[string]artifact.ArtifactRef{}}

	fw, err := gr.NewFramework("general_relativity")
	if err != nil {
		panic(err)
	}
	f.FrameworkRef, err = fw.Mint(s)
	if err != nil {
		panic(err)
	}
	f.Conventions, err = gr.SeedConventions(s)
	if err != nil {
		panic(err)
	}
	f.Assumptions, err = gr.SeedMaxwellBoxAssumptions(s)
	if err != nil {
		panic(err)
	}
	for _, n := range []string{"gamma", "Ebox0", "EboxV", "m", "c", "T", "tau", "xbar",
		"div_j", "bulk_t", "mom", "jN_walls", "bulk", "caps"} {
		f.Symbols[n] = sym(s, n)
	}

	// s1-class: Minkowski metric diag(-1,1,1,1) as symmetric bilinear form
	// (pure mathematics; signature convention lives in assumptions).
	zero, one, negOne := rat(s, 0, 1), rat(s, 1, 1), rat(s, -1, 1)
	comps := make([]artifact.ArtifactRef, 16)
	for i := range comps {
		comps[i] = zero
	}
	for i, v := range []artifact.ArtifactRef{negOne, one, one, one} {
		comps[i*4+i] = v
	}
	// diag(-1,1,1,1): nil InverseRef makes no inverse claim (field
	// semantics); self-inversion holds by component inspection, reviewed
	// at construction, not asserted by the artifact.
	mink := physmath.SymmetricBilinear{Dimension: 4, ComponentRefs: comps, InverseRef: nil}
	if err := mink.Validate(); err != nil {
		panic(err)
	}
	if err := physmath.CheckSymmetricBilinear(s, mink); err != nil {
		panic(err)
	}
	f.MinkowskiRef, err = s.Put(physmath.FamilySymmetricBilinear, mink.Encode())
	if err != nil {
		panic(err)
	}
	if err := physmath.CheckSignature(s, f.MinkowskiRef, physmath.SignatureLorentzian); err != nil {
		panic(err)
	}

	// s1 R1-R3: Rindler metric diag(-xbar^2,1,1,1) as schematic stored
	// geometry (hyperbolic transform itself stays gr-local schematic).
	xbar := physmath.NewEphemeralSymbol("xbar")
	negx2 := physmath.Neg(physmath.Pow(xbar, physmath.MustRational(2, 1)))
	rone := rat(s, 1, 1)
	rzero := rat(s, 0, 1)
	full := []artifact.ArtifactRef{
		mustMint(s, negx2), rone, rone, rone,
		rzero, rone, rzero, rzero,
		rzero, rzero, rone, rzero,
		rzero, rzero, rzero, rone,
	}
	full[1], full[2], full[3] = rzero, rzero, rzero
	rin := physmath.SymmetricBilinear{Dimension: 4, ComponentRefs: full, InverseRef: nil}
	if err := rin.Validate(); err != nil {
		panic(err)
	}
	f.RindlerRef, err = s.Put(physmath.FamilySymmetricBilinear, rin.Encode())
	if err != nil {
		panic(err)
	}

	// R2: determinants through the existing executable path. Derivations
	// are retained as provenance evidence (MINOR-04): volume densities
	// must reference the exact derivation, not just the determinant value.
	var mDeriv, rDeriv physmath.Derivation
	f.MinkowskiDet, mDeriv, err = physmath.DiagonalDeterminant(s, f.MinkowskiRef)
	if err != nil {
		panic(err)
	}
	f.RindlerDet, rDeriv, err = physmath.DiagonalDeterminant(s, f.RindlerRef)
	if err != nil {
		panic(err)
	}
	f.MinkowskiDetDeriv = mustMintDerivation(s, mDeriv)
	f.RindlerDetDeriv = mustMintDerivation(s, rDeriv)

	// R3: volume densities. Minkowski factor 1 (no reduction). Rindler
	// factor |xbar| unreduced, plus reduced-to-xbar gated on xbar>0.
	absx := sym(s, "abs_xbar")
	mvol := physmath.VolumeDensity{MetricRef: f.MinkowskiRef, DeterminantRef: f.MinkowskiDet,
		DeterminantDerivationRef: f.MinkowskiDetDeriv,
		SignatureClass:           physmath.SignatureLorentzian, FactorRef: rat(s, 1, 1)}
	f.MinkowskiVol = mustMintVolumeProven(s, mvol)
	rvol := physmath.VolumeDensity{MetricRef: f.RindlerRef, DeterminantRef: f.RindlerDet,
		DeterminantDerivationRef: f.RindlerDetDeriv,
		SignatureClass:           physmath.SignatureLorentzian, FactorRef: absx}
	f.RindlerVol = mustMintVolumeProven(s, rvol)
	xbarRef := mustMint(s, physmath.NewEphemeralSymbol("xbar"))
	gteBody, err := physmath.EncodeRelationBody(physmath.OpGTE, physmath.NewEphemeralSymbol("xbar"), physmath.MustRational(0, 1))
	if err != nil {
		panic(err)
	}
	gteRel, err := s.Put(physmath.FamilyRelation, gteBody)
	if err != nil {
		panic(err)
	}
	posCond, err := s.Put(physmath.FamilyCondition, gteRel.Encode())
	if err != nil {
		panic(err)
	}
	rred := rvol
	rred.ReductionRef = &xbarRef
	rred.ReductionConditionRefs = []artifact.ArtifactRef{posCond}
	rred.DomainAssumptionRefs = []artifact.ArtifactRef{posCond}
	f.RindlerVolRed = mustMintVolumeProven(s, rred)

	// s2/s4-class: hypersurface domain + box-energy integral (represented).
	dm := physmath.Domain{Kind: physmath.DomainHypersurface,
		DefiningRefs: []artifact.ArtifactRef{f.Symbols["tau"]}}
	if err := dm.Validate(); err != nil {
		panic(err)
	}
	dmPayload, err := dm.Encode()
	if err != nil {
		panic(err)
	}
	f.DomainRef, err = s.Put(physmath.FamilyDomain, dmPayload)
	if err != nil {
		panic(err)
	}
	in := physmath.Integral{DomainRef: f.DomainRef,
		IntegrandRefs: []artifact.ArtifactRef{f.Symbols["T"]}, WeightRef: nil}
	if err := in.Validate(); err != nil {
		panic(err)
	}
	inPayload, err := in.Encode()
	if err != nil {
		panic(err)
	}
	f.IntegralRef, err = s.Put(physmath.FamilyIntegral, inPayload)
	if err != nil {
		panic(err)
	}

	// s3-class: schematic T^a_b (contra/co slots) + contraction certificate.
	mkIdx := func(label string, v physmath.IndexVariance) artifact.ArtifactRef {
		p, err := physmath.Index{Label: label, Variance: v}.Encode()
		if err != nil {
			panic(err)
		}
		r, err := s.Put(physmath.FamilyIndex, p)
		if err != nil {
			panic(err)
		}
		return r
	}
	ia := mkIdx("a", physmath.VarianceContravariant)
	ib := mkIdx("b", physmath.VarianceCovariant)
	tcomps := make([]artifact.ArtifactRef, 4)
	for i := range tcomps {
		tcomps[i] = f.Symbols["T"]
	}
	tn := physmath.Tensor{Rank: 2, Dimension: 2,
		IndexRefs: []artifact.ArtifactRef{ia, ib}, ComponentRefs: tcomps}
	if err := tn.Validate(); err != nil {
		panic(err)
	}
	f.TensorRef, err = s.Put(physmath.FamilyTensor, tn.Encode())
	if err != nil {
		panic(err)
	}
	cert := physmath.Contraction{TensorRef: f.TensorRef, First: 0, Second: 1, ResultRef: f.Symbols["T"]}
	if err := cert.Validate(); err != nil {
		panic(err)
	}
	if err := physmath.CheckContraction(s, cert); err != nil {
		panic(err)
	}
	f.ContractionRef, err = s.Put(physmath.FamilyContraction, cert.Encode())
	if err != nil {
		panic(err)
	}
	f.buildContractionEvidence(s)

	// s5-class executable fragment: EboxV = gamma * Ebox0 (paper (30)).
	mulRef, err := physmath.PutExpression(s, physmath.Mul(
		physmath.NewEphemeralSymbol("gamma"), physmath.NewEphemeralSymbol("Ebox0")))
	if err != nil {
		panic(err)
	}
	canonRes, err := physmath.Apply(s, physmath.Invocation{
		ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
		InputRefs: []artifact.ArtifactRef{mulRef},
	})
	if err != nil {
		panic(err)
	}
	f.TransformDeriv = physmath.Derivation{
		PremiseRefs: []artifact.ArtifactRef{mulRef},
		Steps: []physmath.DerivationStep{{
			Ordinal: 0, ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{mulRef}, OutputRef: canonRes.OutputRef,
		}},
		ResultRefs: []artifact.ArtifactRef{canonRes.OutputRef},
	}
	relPayload, err := physmath.EncodeRelationBody(physmath.OpEQ,
		mustResolveCanon(s, canonRes.OutputRef), physmath.NewEphemeralSymbol("EboxV"))
	if err != nil {
		panic(err)
	}
	relRef, err := s.Put(physmath.FamilyRelation, relPayload)
	if err != nil {
		panic(err)
	}
	th := physmath.Theorem{PremiseRefs: []artifact.ArtifactRef{mulRef},
		ConclusionRef: relRef, DerivationRef: mustMintDerivation(s, f.TransformDeriv)}
	if err := th.Validate(); err != nil {
		panic(err)
	}
	thPayload, err := th.Encode()
	if err != nil {
		panic(err)
	}
	f.TransformThm, err = s.Put(physmath.FamilyTheorem, thPayload)
	if err != nil {
		panic(err)
	}

	// s6-class executable fragment: Ebox0 = m * c^2 (paper (32)).
	mul2, err := physmath.PutExpression(s, physmath.Mul(
		physmath.NewEphemeralSymbol("m"),
		physmath.Pow(physmath.NewEphemeralSymbol("c"), physmath.MustRational(2, 1))))
	if err != nil {
		panic(err)
	}
	canon2, err := physmath.Apply(s, physmath.Invocation{
		ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
		InputRefs: []artifact.ArtifactRef{mul2},
	})
	if err != nil {
		panic(err)
	}
	f.MassDeriv = physmath.Derivation{
		PremiseRefs: []artifact.ArtifactRef{mul2},
		Steps: []physmath.DerivationStep{{
			Ordinal: 0, ProducerNamespace: "physmath", RuleID: physmath.RuleExprCanon,
			InputRefs: []artifact.ArtifactRef{mul2}, OutputRef: canon2.OutputRef,
		}},
		ResultRefs: []artifact.ArtifactRef{canon2.OutputRef},
	}
	rel2, err := physmath.EncodeRelationBody(physmath.OpEQ,
		mustResolveCanon(s, canon2.OutputRef), physmath.NewEphemeralSymbol("Ebox0"))
	if err != nil {
		panic(err)
	}
	relRef2, err := s.Put(physmath.FamilyRelation, rel2)
	if err != nil {
		panic(err)
	}
	th2 := physmath.Theorem{PremiseRefs: []artifact.ArtifactRef{mul2},
		ConclusionRef: relRef2, DerivationRef: mustMintDerivation(s, f.MassDeriv)}
	if err := th2.Validate(); err != nil {
		panic(err)
	}
	thPayload2, err := th2.Encode()
	if err != nil {
		panic(err)
	}
	f.MassThm, err = s.Put(physmath.FamilyTheorem, thPayload2)
	if err != nil {
		panic(err)
	}

	// Framework equations (30) and (32) with K1 assumptions.
	eqWrap := func(id string, rel artifact.ArtifactRef) artifact.ArtifactRef {
		eqPayload := physmath.Equation{RelationRef: rel}.Encode()
		eqRef, err := s.Put(physmath.FamilyEquation, eqPayload)
		if err != nil {
			panic(err)
		}
		pe := gr.PhysicalEquation{EquationID: id, FrameworkRef: f.FrameworkRef,
			EquationRef: eqRef,
			AssumptionRefs: []artifact.ArtifactRef{
				f.Assumptions["stationarity_zero_momentum"],
				f.Assumptions["perfectly_conducting_walls"],
				f.Assumptions["rindler_chart"],
			}}
		if err := pe.Validate(); err != nil {
			panic(err)
		}
		ref, err := pe.Mint(s)
		if err != nil {
			panic(err)
		}
		return ref
	}
	f.Equation30 = eqWrap("maxwellbox_energy_transformation", relRef)
	f.Equation32 = eqWrap("maxwellbox_mass_identification", relRef2)

	// PhysicalTheorem for the replayed (30) fragment (assumptions cited,
	// no interpretations in slice 1).
	pt := gr.PhysicalTheorem{TheoremID: "maxwellbox_transformation_fragment",
		FrameworkRef: f.FrameworkRef, MathematicalTheoremRef: f.TransformThm,
		AssumptionRefs: []artifact.ArtifactRef{
			f.Assumptions["stationarity_zero_momentum"],
			f.Assumptions["perfectly_conducting_walls"],
			f.Assumptions["rindler_chart"],
		}}
	if err := pt.Validate(); err != nil {
		panic(err)
	}
	f.TransformPhysThm, err = pt.Mint(s)
	if err != nil {
		panic(err)
	}

	// s4-class FLUX-CONSERVE-001 inference over schematic s4 evidence.
	// Divergence relation (25-shape), stationarity condition (23-shape),
	// and wall-flux lemma (27-shape) are schematic symbols exactly like
	// the T symbols elsewhere in this fixture; the inference verifies
	// the EVIDENCE CHAIN (families, resolution, domain digest equality,
	// cited stationarity, covered boundary, assembled conclusion), never
	// the analysis itself. See gr/fluxconserve.go for the documented
	// residual gaps (term denotation, lemma soundness).
	mintRel := func(l, r string) artifact.ArtifactRef {
		body, err := physmath.EncodeRelationBody(physmath.OpEQ,
			physmath.NewEphemeralSymbol(l), physmath.NewEphemeralSymbol(r))
		if err != nil {
			panic(err)
		}
		ref, err := s.Put(physmath.FamilyRelation, body)
		if err != nil {
			panic(err)
		}
		return ref
	}
	divergenceRef := mintRel("div_j", "bulk_t")
	statPred, err := physmath.EncodeRelationBody(physmath.OpEQ,
		physmath.NewEphemeralSymbol("mom"), physmath.MustRational(0, 1))
	if err != nil {
		panic(err)
	}
	statPredRef, err := s.Put(physmath.FamilyRelation, statPred)
	if err != nil {
		panic(err)
	}
	statRef, err := s.Put(physmath.FamilyCondition, statPredRef.Encode())
	if err != nil {
		panic(err)
	}
	wallBody, err := physmath.EncodeRelationBody(physmath.OpEQ,
		physmath.NewEphemeralSymbol("jN_walls"), physmath.MustRational(0, 1))
	if err != nil {
		panic(err)
	}
	wallRel, err := s.Put(physmath.FamilyRelation, wallBody)
	if err != nil {
		panic(err)
	}
	wallLemma, err := s.Put(physmath.FamilyEquation,
		physmath.Equation{RelationRef: wallRel}.Encode())
	if err != nil {
		panic(err)
	}
	f.FluxCert, err = gr.CheckFluxConservation(s, divergenceRef, f.IntegralRef, f.DomainRef, gr.FluxEvidence{
		BoundaryLemmaRefs: []artifact.ArtifactRef{wallLemma},
		StationarityRef:   statRef,
		BulkTermRef:       f.Symbols["bulk"],
		CapsTermRef:       f.Symbols["caps"],
		AssumptionRefs: []artifact.ArtifactRef{
			f.Assumptions["stationarity_zero_momentum"],
			f.Assumptions["perfectly_conducting_walls"],
			f.Assumptions["rindler_chart"],
		},
		FrameworkRef: f.FrameworkRef,
	})
	if err != nil {
		panic(err)
	}

	// Sketch: s1-s3 UNSUPPORTED (represented), s4 SUPPORTED via the flux
	// inference conclusion, s5-s6 SUPPORTED (replayed).
	// UnitID is lowercase per gr identifier grammar; it denotes rev2 D1.
	f.Sketch = gr.ProtocolSketch{SketchID: "maxwellbox_k1", UnitID: "d1", Steps: []gr.SketchStep{
		{StepID: "s1", Description: "Rindler setup (4,6,7); metrics/dets/volumes validated, transform schematic",
			Status: gr.SketchUnsupported, Refs: []artifact.ArtifactRef{f.MinkowskiRef, f.RindlerRef, f.MinkowskiDet, f.RindlerDet, f.MinkowskiVol, f.RindlerVol, f.RindlerVolRed}},
		{StepID: "s2", Description: "Lab current + energy integral (18-21); w_a evaluated + j_a negation evidenced, integral unexecutable",
			Status: gr.SketchUnsupported, Refs: []artifact.ArtifactRef{f.IntegralRef, f.DomainRef, f.WRef, f.WEvalRef, f.JRef}},
		{StepID: "s3", Description: "Stationarity (23); contracted tensor integral",
			Status: gr.SketchUnsupported, Refs: []artifact.ArtifactRef{f.ContractionRef, f.Assumptions["stationarity_zero_momentum"]}},
		{StepID: "s4", Description: "Conserved integral + boundary flux (24-29, A28/A29); FLUX-CONSERVE-001 schema-checked",
			Status: gr.SketchSupported, Refs: []artifact.ArtifactRef{f.FluxCert.ConclusionRef}},
		{StepID: "s5", Description: "Energy transformation (30); replayed fragment",
			Status: gr.SketchSupported, Refs: []artifact.ArtifactRef{f.TransformThm}},
		{StepID: "s6", Description: "Mass identification (32); replayed fragment",
			Status: gr.SketchSupported, Refs: []artifact.ArtifactRef{f.MassThm}},
	}}
	if err := f.Sketch.Validate(); err != nil {
		panic(err)
	}
	return f
}

func mustMint(s physmath.Store, e physmath.Expr) artifact.ArtifactRef {
	p, err := e.Encode()
	if err != nil {
		panic(err)
	}
	r, err := s.Put(physmath.FamilyExpression, p)
	if err != nil {
		panic(err)
	}
	return r
}

func mustMintVolume(s physmath.Store, v physmath.VolumeDensity) artifact.ArtifactRef {
	p, err := v.Encode()
	if err != nil {
		panic(err)
	}
	r, err := s.Put(physmath.FamilyVolumeDensity, p)
	if err != nil {
		panic(err)
	}
	return r
}

// mustMintVolumeProven mints only provenance-valid volume densities: bare
// structural validity is never enough to enter the fixture. A structurally
// valid but provenance-invalid record panics here instead of becoming a
// fixture artifact.
func mustMintVolumeProven(s physmath.Store, v physmath.VolumeDensity) artifact.ArtifactRef {
	if err := v.Validate(); err != nil {
		panic(err)
	}
	if err := physmath.CheckVolumeDensity(s, v); err != nil {
		panic(err)
	}
	return mustMintVolume(s, v)
}

func mustResolveCanon(s physmath.Store, ref artifact.ArtifactRef) physmath.Expr {
	fam, payload, err := s.Lookup(ref)
	if err != nil {
		panic(err)
	}
	if fam != physmath.FamilyExpression {
		panic(fmt.Sprintf("expected expression, got %v", fam))
	}
	e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
	if err != nil {
		panic(err)
	}
	return e
}

func mustMintDerivation(s physmath.Store, d physmath.Derivation) artifact.ArtifactRef {
	if err := d.ValidateStructure(); err != nil {
		panic(err)
	}
	payload, err := d.Encode()
	if err != nil {
		panic(err)
	}
	ref, err := s.Put(physmath.FamilyDerivation, payload)
	if err != nil {
		panic(err)
	}
	return ref
}

// Verify replays both executable fragments with theorem-conclusion gates
// and checks the sketch records exactly s1-s4 as unsupported. It proves
// fragment replay + honest limitation record — never K1 end-to-end.
func (f *Fixture) Verify() error {
	if err := physmath.Replay(f.Store, f.TransformDeriv, nil); err != nil {
		return fmt.Errorf("maxwellbox: transform fragment replay: %w", err)
	}
	if err := physmath.Replay(f.Store, f.MassDeriv, nil); err != nil {
		return fmt.Errorf("maxwellbox: mass fragment replay: %w", err)
	}
	thFam, _, err := f.Store.Lookup(f.TransformThm)
	if err != nil {
		return err
	}
	if thFam != physmath.FamilyTheorem {
		return fmt.Errorf("maxwellbox: transform theorem family wrong")
	}
	// Conclusion gates over designated finals.
	th := decodeTheorem(f.Store, f.TransformThm)
	if err := physmath.CheckTheoremConclusion(f.Store, th,
		f.TransformDeriv.ResultRefs[0], f.Symbols["EboxV"]); err != nil {
		return fmt.Errorf("maxwellbox: transform conclusion gate: %w", err)
	}
	th2 := decodeTheorem(f.Store, f.MassThm)
	if err := physmath.CheckTheoremConclusion(f.Store, th2,
		f.MassDeriv.ResultRefs[0], f.Symbols["Ebox0"]); err != nil {
		return fmt.Errorf("maxwellbox: mass conclusion gate: %w", err)
	}
	unsup := f.Sketch.UnsupportedSteps()
	want := []string{"s1", "s2", "s3"}
	if len(unsup) != len(want) {
		return fmt.Errorf("maxwellbox: unsupported steps %v, want %v", unsup, want)
	}
	for i := range want {
		if unsup[i] != want[i] {
			return fmt.Errorf("maxwellbox: unsupported steps %v, want %v", unsup, want)
		}
	}
	return nil
}

func decodeTheorem(s physmath.Store, ref artifact.ArtifactRef) physmath.Theorem {
	fam, payload, err := s.Lookup(ref)
	if err != nil {
		panic(err)
	}
	if fam != physmath.FamilyTheorem {
		panic("not a theorem")
	}
	d := artifact.NewDecoder(payload)
	preCount, err := d.SequenceCount()
	if err != nil {
		panic(err)
	}
	var pre []artifact.ArtifactRef
	for i := 0; i < preCount; i++ {
		// SET decode = count + ordered refs (canonical order).
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			panic(err)
		}
		pre = append(pre, r)
	}
	conc, err := artifact.DecodeArtifactRef(d)
	if err != nil {
		panic(err)
	}
	asmCount, err := d.SequenceCount()
	if err != nil {
		panic(err)
	}
	var asm []artifact.ArtifactRef
	for i := 0; i < asmCount; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			panic(err)
		}
		asm = append(asm, r)
	}
	drv, err := artifact.DecodeArtifactRef(d)
	if err != nil {
		panic(err)
	}
	return physmath.Theorem{PremiseRefs: pre, ConclusionRef: conc, AssumptionRefs: asm, DerivationRef: drv}
}

// SketchDigestForTest digests sketch step IDs, statuses, and reference
// bytes for build-determinism comparison. It is test support, not a
// correctness claim.
func (f *Fixture) SketchDigestForTest() ([32]byte, error) {
	h := sha256.New()
	for _, st := range f.Sketch.Steps {
		h.Write([]byte(st.StepID))
		h.Write([]byte(st.Status))
		for _, r := range st.Refs {
			h.Write(r.Encode())
		}
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out, nil
}
