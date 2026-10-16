// Package protocol verifies the ltb ZEH/LTB reference slice: replayed
// gauge-derivative fragment with theorem gate, honest conditional EdS
// baseline, independent goldens with anti-goldens, stored-artifact
// cone checks, production-path assumption linkage, claim-state
// preservation, and verification coverage. A green suite means the
// fragment replays and the open conjecture stays open — never that
// the LTB theorem is proved (see manifest limitations).
//
// Discipline: every test inspects the actual stored artifact it
// claims to verify. Structural expectations are shape checks (node
// kinds, rational leaves, bound identities), never transcriptions
// of the fixture's own construction sequence. Status language is
// kept exact: finite-sample corroboration is labeled as such, and
// the all-r derivative argument is structural.
package protocol

import (
	"math"
	"math/big"
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-lib/gr/ltb"
	"github.com/PithomLabs/phys-math"
)

func rat(n, d int64) *big.Rat { return big.NewRat(n, d) }

func TestFixtureVerify(t *testing.T) {
	f := ltb.Build()
	if err := f.Verify(); err != nil {
		t.Fatalf("fixture verify: %v", err)
	}
}

func decodeDerivation(t *testing.T, f *ltb.Fixture, dref artifact.ArtifactRef) physmath.Derivation {
	t.Helper()
	_, payload, err := f.Store.Lookup(dref)
	if err != nil {
		t.Fatal(err)
	}
	d, err := physmath.DecodeDerivation(payload)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func decodeExpr(t *testing.T, f *ltb.Fixture, ref artifact.ArtifactRef) physmath.Expr {
	t.Helper()
	fam, payload, err := f.Store.Lookup(ref)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyExpression {
		t.Fatalf("artifact is %s/%s, not expression/1", fam.Namespace, fam.ArtifactType)
	}
	e, err := physmath.DecodeExpr(artifact.NewDecoder(payload))
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestDerivationInventory(t *testing.T) {
	f := ltb.Build()
	if len(f.Derivs) != 2 {
		t.Fatalf("Derivs holds %d entries, want 2 (diff + theta-canon)", len(f.Derivs))
	}
	seen := map[artifact.ArtifactRef]bool{}
	for i, dref := range f.Derivs {
		fam, payload, err := f.Store.Lookup(dref)
		if err != nil {
			t.Fatalf("evidence %d unresolvable: %v", i, err)
		}
		if fam != physmath.FamilyDerivation {
			t.Fatalf("evidence %d is %s/%s, not derivation/1", i, fam.Namespace, fam.ArtifactType)
		}
		if seen[dref] {
			t.Fatalf("evidence %d duplicates an earlier derivation", i)
		}
		seen[dref] = true
		d, err := physmath.DecodeDerivation(payload)
		if err != nil {
			t.Fatalf("evidence %d undecodable: %v", i, err)
		}
		if err := physmath.Replay(f.Store, d, nil); err != nil {
			t.Fatalf("evidence %d does not replay: %v", i, err)
		}
	}

	// Entry 0 is the differentiation derivation: it must consume MRef
	// and the declared radial function, target the declared parameter,
	// and produce MPrimeRef — not merely replay something.
	d0 := decodeDerivation(t, f, f.Derivs[0])
	if len(d0.Steps) != 1 {
		t.Fatalf("diff derivation has %d steps, want 1", len(d0.Steps))
	}
	s0 := d0.Steps[0]
	if s0.ProducerNamespace != physmath.NamespacePhysMath || s0.RuleID != physmath.RuleDifferentiate {
		t.Fatalf("diff step uses (%q,%q), want physmath DIFFERENTIATE-001",
			s0.ProducerNamespace, s0.RuleID)
	}
	if len(s0.InputRefs) != 2 || s0.InputRefs[0] != f.MRef || s0.InputRefs[1] != f.RFnRef {
		t.Fatal("diff step inputs are not [MRef, RFnRef]")
	}
	if len(s0.ConditionRefs) != 0 {
		t.Fatal("diff step carries unexpected conditions")
	}
	rID := declaredParam(t, f)
	wantParams := append(
		physmath.BoundParameterRef{DeclaringArtifactRef: f.RFnRef, ParameterID: rID}.Encode(),
		artifact.EncodeUInt32(1)...)
	if string(s0.Parameters) != string(wantParams) {
		t.Fatal("diff step does not target the declared radial parameter at order 1")
	}
	if s0.OutputRef != f.MPrimeRef {
		t.Fatal("diff step output is not MPrimeRef")
	}
	if len(d0.ResultRefs) != 1 || d0.ResultRefs[0] != f.MPrimeRef {
		t.Fatal("diff derivation result is not MPrimeRef")
	}

	// Entry 1 is the theta canonicalization: input ThetaRef, rule
	// CANON, output ThetaCanon.
	d1 := decodeDerivation(t, f, f.Derivs[1])
	if len(d1.Steps) != 1 {
		t.Fatalf("canon derivation has %d steps, want 1", len(d1.Steps))
	}
	s1 := d1.Steps[0]
	if s1.ProducerNamespace != physmath.NamespacePhysMath || s1.RuleID != physmath.RuleExprCanon {
		t.Fatalf("canon step uses (%q,%q), want physmath EXPR-CANON-001",
			s1.ProducerNamespace, s1.RuleID)
	}
	if len(s1.InputRefs) != 1 || s1.InputRefs[0] != f.ThetaRef {
		t.Fatal("canon step input is not ThetaRef")
	}
	if s1.OutputRef != f.ThetaCanon {
		t.Fatal("canon step output is not ThetaCanon")
	}
	if len(d1.ResultRefs) != 1 || d1.ResultRefs[0] != f.ThetaCanon {
		t.Fatal("canon derivation result is not ThetaCanon")
	}

	// The hand-golden relation and theorem conclusion must refer to
	// the intended stored derivative result.
	relFam, relPayload, err := f.Store.Lookup(f.MPrimeRelRef)
	if err != nil {
		t.Fatal(err)
	}
	if relFam != physmath.FamilyRelation {
		t.Fatalf("golden relation is %s/%s, not relation/1", relFam.Namespace, relFam.ArtifactType)
	}
	rd := artifact.NewDecoder(relPayload)
	opStr, err := rd.String()
	if err != nil || opStr != string(physmath.OpEQ) {
		t.Fatal("golden relation is not an equality")
	}
	relLeft, err := physmath.DecodeExpr(rd)
	if err != nil {
		t.Fatal(err)
	}
	lb, err := relLeft.Encode()
	if err != nil {
		t.Fatal(err)
	}
	_, outPayload, err := f.Store.Lookup(f.MPrimeRef)
	if err != nil {
		t.Fatal(err)
	}
	if string(lb) != string(outPayload) {
		t.Fatal("golden relation left side is not the stored MPrimeRef bytes")
	}

	if err := f.Verify(); err != nil {
		t.Fatalf("valid fixture rejected: %v", err)
	}
}

func TestCorruptedOutputFailsReplay(t *testing.T) {
	f := ltb.Build()
	d := decodeDerivation(t, f, f.Derivs[0])
	bad := d
	bad.Steps = append([]physmath.DerivationStep(nil), d.Steps...)
	bad.Steps[0].OutputRef = f.ThetaRef
	bad.ResultRefs = []artifact.ArtifactRef{f.ThetaRef}
	if err := physmath.Replay(f.Store, bad, nil); err == nil {
		t.Fatal("replay accepted a corrupted fragment output")
	}
}

func TestDerivationMutationsFail(t *testing.T) {
	f := ltb.Build()
	d := decodeDerivation(t, f, f.Derivs[0])
	// Wrong rule on the same inputs must not reproduce the output.
	rule := d
	rule.Steps = append([]physmath.DerivationStep(nil), d.Steps...)
	rule.Steps[0].RuleID = physmath.RuleExprCanon
	if err := physmath.Replay(f.Store, rule, nil); err == nil {
		t.Fatal("replay accepted a derivation with a substituted rule")
	}
	// Swapped inputs must not reproduce the output.
	inputs := d
	inputs.Steps = append([]physmath.DerivationStep(nil), d.Steps...)
	inputs.Steps[0].InputRefs = []artifact.ArtifactRef{f.RFnRef, f.MRef}
	if err := physmath.Replay(f.Store, inputs, nil); err == nil {
		t.Fatal("replay accepted a derivation with swapped inputs")
	}
	// Theta-canon derivation fed the wrong input must fail.
	c := decodeDerivation(t, f, f.Derivs[1])
	wrong := c
	wrong.Steps = append([]physmath.DerivationStep(nil), c.Steps...)
	wrong.Steps[0].InputRefs = []artifact.ArtifactRef{f.SigmaRef}
	if err := physmath.Replay(f.Store, wrong, nil); err == nil {
		t.Fatal("canon replay accepted a substituted input")
	}
}

// evalRat is an independent exact-rational evaluator for the closed
// arithmetic Expr subset (rational/symbol/add/mul/neg/integer-pow),
// written by hand from the Expr semantics for this test only. It never
// calls the engine under test. Symbols must carry the expected bound
// parameter identity: ephemeral symbols and same-display symbols with
// a different ParameterID are rejected, so display names alone never
// establish variable identity.
func evalRat(e physmath.Expr, wantPID physmath.ParameterID, env map[string]*big.Rat) (*big.Rat, error) {
	switch e.Kind {
	case physmath.TagRational:
		return new(big.Rat).SetFrac(e.Numerator.ToBig(), e.Denominator.ToBig()), nil
	case physmath.TagSymbol:
		if e.ParamID == nil {
			return nil, errEval("unbound symbol " + e.DisplayName)
		}
		if *e.ParamID != wantPID {
			return nil, errEval("wrong parameter identity for " + e.DisplayName)
		}
		v, ok := env[e.DisplayName]
		if !ok {
			return nil, errEval("no value for " + e.DisplayName)
		}
		return new(big.Rat).Set(v), nil
	case physmath.TagAdd:
		acc := new(big.Rat)
		for _, c := range e.Children {
			v, err := evalRat(c, wantPID, env)
			if err != nil {
				return nil, err
			}
			acc.Add(acc, v)
		}
		return acc, nil
	case physmath.TagMul:
		acc := big.NewRat(1, 1)
		for _, c := range e.Children {
			v, err := evalRat(c, wantPID, env)
			if err != nil {
				return nil, err
			}
			acc.Mul(acc, v)
		}
		return acc, nil
	case physmath.TagNeg:
		v, err := evalRat(*e.Operand, wantPID, env)
		if err != nil {
			return nil, err
		}
		return new(big.Rat).Neg(v), nil
	case physmath.TagPow:
		base, err := evalRat(*e.Base, wantPID, env)
		if err != nil {
			return nil, err
		}
		exp, err := evalRat(*e.Exp, wantPID, env)
		if err != nil {
			return nil, err
		}
		if !exp.IsInt() {
			return nil, errEval("non-integer exponent")
		}
		// Fail-closed magnitude gate (test-oracle bound, not a
		// mathematical restriction): safe big.Int comparisons BEFORE
		// any int64 conversion or negation. This closes both the
		// oversized-exponent narrowing path (Int64() is undefined
		// out of range) and the MinInt64 negation edge case
		// (-MinInt64 overflows int64). Larger exponents are rejected
		// rather than miscomputed or looped over.
		num := exp.Num()
		if !num.IsInt64() {
			return nil, errEval("exponent outside int64 range")
		}
		if new(big.Int).Abs(num).Cmp(big.NewInt(maxEvalExp)) > 0 {
			return nil, errEval("exponent outside evaluator bound")
		}
		n := num.Int64()
		res := big.NewRat(1, 1)
		b := new(big.Rat).Set(base)
		if n < 0 {
			if b.Sign() == 0 {
				return nil, errEval("zero base, negative exponent")
			}
			b.Inv(b)
			n = -n
		}
		for i := int64(0); i < n; i++ {
			res.Mul(res, b)
		}
		return res, nil
	}
	return nil, errEval("unevaluatable node")
}

type evalError string

func (e evalError) Error() string { return "eval: " + string(e) }

func errEval(s string) error { return evalError(s) }

// maxEvalExp bounds exact exponentiation in this test-side evaluator.
// It is a fail-closed implementation limit of the oracle, not a
// restriction on PhysMath's mathematical domain.
const maxEvalExp = int64(1024)

func TestEvaluatorSanity(t *testing.T) {
	// The evaluator itself is checked on hand arithmetic first; a
	// broken oracle would invalidate every pointwise claim below.
	f := ltb.Build()
	rID := declaredParam(t, f)
	r := physmath.NewBoundSymbol("r", rID)
	env := map[string]*big.Rat{"r": big.NewRat(3, 1)}
	sum := physmath.Add(physmath.MustRational(1, 2), physmath.MustRational(1, 3))
	v, err := evalRat(sum, rID, env)
	if err != nil || v.Cmp(big.NewRat(5, 6)) != 0 {
		t.Fatalf("evaluator addition = %v, %v", v, err)
	}
	golden := physmath.Mul(physmath.MustRational(3, 2),
		physmath.Pow(r, physmath.MustRational(2, 1)))
	v, err = evalRat(golden, rID, env)
	if err != nil || v.Cmp(big.NewRat(27, 2)) != 0 {
		t.Fatalf("evaluator golden at r=3 = %v, %v", v, err)
	}
	if _, err := evalRat(physmath.NewEphemeralSymbol("H"), rID, env); err == nil {
		t.Fatal("evaluator accepted an unbound symbol")
	}
	// Same display name, different parameter identity: rejected.
	other := physmath.NewBoundSymbol("r", physmath.DeriveParameterID(99))
	if _, err := evalRat(other, rID, env); err == nil {
		t.Fatal("evaluator accepted a same-display different-identity symbol")
	}
}

func TestEvaluatorExponentBounds(t *testing.T) {
	// Each case must exercise the evalRat power branch and return an
	// error: no silent narrowing, no overflow-induced value, no hang.
	f := ltb.Build()
	rID := declaredParam(t, f)
	r := physmath.NewBoundSymbol("r", rID)
	env := map[string]*big.Rat{"r": big.NewRat(2, 1)}
	// Oversized nested exponent: 10^30 overflows int64 narrowing.
	nested := physmath.Pow(r, physmath.Pow(
		physmath.MustRational(10, 1), physmath.MustRational(30, 1)))
	if _, err := evalRat(nested, rID, env); err == nil {
		t.Fatal("oversized nested exponent accepted")
	}
	// Minimum int64 exponent: negation would overflow; must fail
	// before negation or looping.
	minExp := physmath.Pow(r, physmath.MustRational(math.MinInt64, 1))
	if _, err := evalRat(minExp, rID, env); err == nil {
		t.Fatal("MinInt64 exponent accepted")
	}
	// Large but representable exponent: rejected by the oracle bound
	// without attempting a billion iterations.
	bigExp := physmath.Pow(r, physmath.MustRational(1000000000, 1))
	if _, err := evalRat(bigExp, rID, env); err == nil {
		t.Fatal("billion-scale exponent accepted")
	}
	// Positive controls: small exponents 0..3 keep working, including
	// a nonzero-base negative power.
	for _, k := range []int64{0, 1, 2, 3, -2} {
		p := physmath.Pow(r, physmath.MustRational(k, 1))
		v, err := evalRat(p, rID, env)
		if err != nil {
			t.Fatalf("exponent %d rejected: %v", k, err)
		}
		want := new(big.Rat).SetInt64(1)
		two := big.NewRat(2, 1)
		e := k
		if e < 0 {
			two.Inv(two)
			e = -e
		}
		// e is in {-2,0,1,2,3} here: safe by construction.
		for i := int64(0); i < e; i++ {
			want.Mul(want, two)
		}
		if v.Cmp(want) != 0 {
			t.Fatalf("r^%d at r=2 = %s, want %s", k, v, want)
		}
	}
	// Zero base to a negative power stays rejected.
	zero := physmath.Pow(physmath.MustRational(0, 1), physmath.MustRational(-1, 1))
	if _, err := evalRat(zero, rID, env); err == nil {
		t.Fatal("zero base with negative exponent accepted")
	}
}

func TestMPrimePointwise(t *testing.T) {
	// Finite-sample corroboration only: exact evaluation of the STORED
	// output artifact at rational sample points. Agreement here does
	// not prove the identity for all r; the all-r argument is
	// TestMPrimeStructureAllR below. Supplementary evidence per
	// Lesson 3, not a symbolic proof.
	f := ltb.Build()
	rID := declaredParam(t, f)
	out := decodeExpr(t, f, f.MPrimeRef)
	for _, qs := range []string{"0", "1/2", "3", "-2", "7/5"} {
		q, ok := new(big.Rat).SetString(qs)
		if !ok {
			t.Fatalf("bad sample %s", qs)
		}
		got, err := evalRat(out, rID, map[string]*big.Rat{"r": q})
		if err != nil {
			t.Fatalf("output unevaluatable at r=%s: %v", qs, err)
		}
		want := new(big.Rat).Mul(big.NewRat(3, 2), new(big.Rat).Mul(q, q))
		if got.Cmp(want) != 0 {
			t.Fatalf("dM/dr at r=%s = %s, want %s", qs, got, want)
		}
	}
	// Discrimination: the wrong-power form (3/2)r^3 must NOT agree.
	r := physmath.NewBoundSymbol("r", rID)
	wrong := physmath.Mul(physmath.MustRational(3, 2),
		physmath.Pow(r, physmath.MustRational(3, 1)))
	q := big.NewRat(3, 1)
	got, err := evalRat(out, rID, map[string]*big.Rat{"r": q})
	if err != nil {
		t.Fatal(err)
	}
	wv, err := evalRat(wrong, rID, map[string]*big.Rat{"r": q})
	if err != nil {
		t.Fatal(err)
	}
	if got.Cmp(wv) == 0 {
		t.Fatal("pointwise check cannot distinguish r^2 from r^3 golden")
	}
}

func TestGaugeAntiGolden(t *testing.T) {
	f := ltb.Build()
	pid := physmath.DeriveParameterID(0)
	r := physmath.NewBoundSymbol("r", pid)
	// The notorious gauge typo: M = r^(3/2) differentiates to
	// (3/2)r^(1/2), which must never equal the 3r^2/2 golden.
	wrong, err := physmath.Canonicalize(physmath.Mul(
		physmath.MustRational(3, 2),
		physmath.Pow(r, physmath.MustRational(1, 2))))
	if err != nil {
		t.Fatal(err)
	}
	gotExpr := decodeExpr(t, f, f.MPrimeRef)
	got, err := physmath.Canonicalize(gotExpr)
	if err != nil {
		t.Fatal(err)
	}
	if got.Equal(wrong) {
		t.Fatal("gauge-confusion derivative coincides with the golden")
	}
}

func TestDAStarSeries(t *testing.T) {
	// Independent mechanical recomputation from the binomial series
	// for (1+z)^-1/2, divided by (1+z). Starting point differs from
	// the README hand derivation (which works term-by-term).
	alpha := big.NewRat(-1, 2)
	c := []*big.Rat{big.NewRat(1, 1)}
	for k := 1; k <= 3; k++ {
		// c_k = c_{k-1} * (alpha-k+1)/k
		factor := new(big.Rat).Sub(alpha, big.NewRat(int64(k-1), 1))
		factor.Quo(factor, big.NewRat(int64(k), 1))
		c = append(c, new(big.Rat).Mul(c[k-1], factor))
	}
	// N = 2(1-S): n_0 = 0, n_k = -2 c_k.
	n := []*big.Rat{big.NewRat(0, 1)}
	for k := 1; k <= 3; k++ {
		n = append(n, new(big.Rat).Mul(big.NewRat(-2, 1), c[k]))
	}
	// D = N/(1+z): d_k = n_k - d_{k-1}.
	d := []*big.Rat{big.NewRat(0, 1)}
	for k := 1; k <= 3; k++ {
		d = append(d, new(big.Rat).Sub(n[k], d[k-1]))
	}
	want := []*big.Rat{rat(1, 1), rat(-7, 4), rat(19, 8)}
	f := ltb.Build()
	if len(f.DACoeffs) != 3 {
		t.Fatalf("DACoeffs holds %d entries, want 3", len(f.DACoeffs))
	}
	for k := 1; k <= 3; k++ {
		if d[k].Cmp(want[k-1]) != 0 {
			t.Fatalf("recomputed D_A* order-%d = %s, want %s", k, d[k], want[k-1])
		}
		e := decodeExpr(t, f, f.DACoeffs[k-1])
		stored, err := physmath.Canonicalize(e)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := physmath.Canonicalize(
			physmath.MustRational(want[k-1].Num().Int64(), want[k-1].Denom().Int64()))
		if err != nil {
			t.Fatal(err)
		}
		if !stored.Equal(expected) {
			t.Fatalf("stored D_A* order-%d != recomputed %s", k, want[k-1])
		}
		// Historical slip 23/8 must not coincide with the golden.
		if k == 3 {
			slip, _ := physmath.Canonicalize(physmath.MustRational(23, 8))
			if stored.Equal(slip) {
				t.Fatal("stored coefficient matches the known 23/8 slip")
			}
		}
	}
}

func TestVertexCoefficients(t *testing.T) {
	// Stored EdS vertex-series coefficients, keyed by series order,
	// checked against an independent mechanical expansion of
	// Z0(r) = (1-r/2)^-2 with z(r) = Z0(r)-1. This covers the beta=0
	// case of the candidate recursion only — not the general z3
	// formula, analytic existence, or any ZEH-Q claim.
	coeff := func(k int) *big.Rat {
		// binom(-2,k) * (-1/2)^k
		binom := big.NewRat(1, 1)
		for i := 1; i <= k; i++ {
			binom.Mul(binom, big.NewRat(int64(-2-i+1), int64(i)))
		}
		pow := big.NewRat(1, 1)
		half := big.NewRat(-1, 2)
		for i := 0; i < k; i++ {
			pow.Mul(pow, half)
		}
		return new(big.Rat).Mul(binom, pow)
	}
	f := ltb.Build()
	if len(f.VertexCoeffs) != 3 {
		t.Fatalf("VertexCoeffs holds %d entries, want 3", len(f.VertexCoeffs))
	}
	want := map[int]*big.Rat{1: rat(1, 1), 2: rat(3, 4), 3: rat(1, 2)}
	for order := 1; order <= 3; order++ {
		ref, ok := f.VertexCoeffs[order]
		if !ok {
			t.Fatalf("vertex coefficient for order %d missing (mapping, not position)", order)
		}
		if got := coeff(order); got.Cmp(want[order]) != 0 {
			t.Fatalf("recomputed z_%d = %s, want %s", order, got, want[order])
		}
		e := decodeExpr(t, f, ref)
		stored, err := physmath.Canonicalize(e)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := physmath.Canonicalize(
			physmath.MustRational(want[order].Num().Int64(), want[order].Denom().Int64()))
		if err != nil {
			t.Fatal(err)
		}
		if !stored.Equal(expected) {
			t.Fatalf("stored z_%d != recomputed %s", order, want[order])
		}
	}
	// Negative: an incorrect coefficient must not coincide.
	e := decodeExpr(t, f, f.VertexCoeffs[2])
	stored, _ := physmath.Canonicalize(e)
	wrong, _ := physmath.Canonicalize(physmath.MustRational(2, 3))
	if stored.Equal(wrong) {
		t.Fatal("stored z_2 coincides with incorrect 2/3")
	}
	// Negative: order-to-artifact mapping must be exact (z_2 != z_1's value).
	one, _ := physmath.Canonicalize(physmath.MustRational(1, 1))
	if stored.Equal(one) {
		t.Fatal("order mapping confused: z_2 artifact equals z_1 value")
	}
}

// requireBoundR asserts e is the declaration-bound radial symbol:
// display "r" with the exact ParameterID from the declaring function.
// Display names alone never establish identity.
func requireBoundR(t *testing.T, e physmath.Expr, rID physmath.ParameterID, what string) {
	t.Helper()
	if e.Kind != physmath.TagSymbol || e.DisplayName != "r" ||
		e.ParamID == nil || *e.ParamID != rID {
		t.Fatalf("%s is not the declared r", what)
	}
}

func requireRat(t *testing.T, e physmath.Expr, n, d int64, what string) {
	t.Helper()
	if e.Kind != physmath.TagRational {
		t.Fatalf("%s is not rational", what)
	}
	if new(big.Rat).SetFrac(e.Numerator.ToBig(), e.Denominator.ToBig()).Cmp(rat(n, d)) != 0 {
		t.Fatalf("%s is not %d/%d", what, n, d)
	}
}

func childrenOfKind(e physmath.Expr, kind physmath.Tag) []physmath.Expr {
	var out []physmath.Expr
	for _, c := range e.Children {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

func TestConeIdentityRational(t *testing.T) {
	// The stored cone expressions are decoded, shape-checked against
	// their documented forms, then evaluated with the strict
	// identity-checking evaluator; those evaluated values feed the
	// identity check. A wrong stored artifact fails here. Finite-sample
	// corroboration only, not a symbolic proof by the engine.
	f := ltb.Build()
	rID := declaredParam(t, f)
	for _, key := range []string{"u", "V0", "Q0", "Rc0", "P0"} {
		if _, ok := f.ConeRefs[key]; !ok {
			t.Fatalf("ConeRefs[%q] missing", key)
		}
	}
	u := decodeExpr(t, f, f.ConeRefs["u"])
	// u = 1 + Neg(Mul(1/2, r)): Add root, rational 1, one Neg factor
	// holding the 1/2-scaled declared r.
	if u.Kind != physmath.TagAdd {
		t.Fatalf("stored u is kind %d, want TagAdd", u.Kind)
	}
	foundOne, foundHalf := false, false
	for _, c := range u.Children {
		if c.Kind == physmath.TagRational {
			if new(big.Rat).SetFrac(c.Numerator.ToBig(), c.Denominator.ToBig()).Cmp(rat(1, 1)) == 0 {
				foundOne = true
			}
			continue
		}
		if c.Kind != physmath.TagNeg || c.Operand == nil {
			t.Fatalf("stored u has unexpected child kind %d", c.Kind)
		}
		m := *c.Operand
		if m.Kind != physmath.TagMul {
			t.Fatalf("stored u negated factor is kind %d, want TagMul", m.Kind)
		}
		half, hasR := false, false
		for _, g := range m.Children {
			if g.Kind == physmath.TagRational &&
				new(big.Rat).SetFrac(g.Numerator.ToBig(), g.Denominator.ToBig()).Cmp(rat(1, 2)) == 0 {
				half = true
			}
			if g.Kind == physmath.TagSymbol {
				requireBoundR(t, g, rID, "stored u radial factor")
				hasR = true
			}
		}
		if !half || !hasR {
			t.Fatal("stored u negated factor is not (1/2)r with the declared r")
		}
		foundHalf = true
	}
	if !foundOne || !foundHalf {
		t.Fatal("stored u is not 1-(1/2)r in the expected shape")
	}

	// V0 = Mul(r, Pow(u,-1)): Mul root with a bound-r factor and a
	// -1 power; Q0 = Pow(u,-1); Rc0 = Mul(r, Pow(u,2));
	// P0 = Pow(u,2). Powers must be exact integer exponents.
	checkPow := func(e physmath.Expr, wantExp *big.Rat, what string) {
		t.Helper()
		if e.Kind != physmath.TagPow {
			t.Fatalf("stored %s is kind %d, want TagPow", what, e.Kind)
		}
		x, err := evalRat(*e.Exp, rID, nil)
		if err != nil || x.Cmp(wantExp) != 0 {
			t.Fatalf("stored %s exponent is not %s", what, wantExp)
		}
	}
	v0 := decodeExpr(t, f, f.ConeRefs["V0"])
	if v0.Kind != physmath.TagMul {
		t.Fatalf("stored V0 is kind %d, want TagMul", v0.Kind)
	}
	vPows := childrenOfKind(v0, physmath.TagPow)
	if len(vPows) != 1 {
		t.Fatalf("stored V0 holds %d Pow children, want 1", len(vPows))
	}
	checkPow(vPows[0], rat(-1, 1), "V0")
	hasR := false
	for _, c := range v0.Children {
		if c.Kind == physmath.TagSymbol {
			requireBoundR(t, c, rID, "stored V0 radial factor")
			hasR = true
		}
	}
	if !hasR {
		t.Fatal("stored V0 has no declared-r factor")
	}
	q0 := decodeExpr(t, f, f.ConeRefs["Q0"])
	checkPow(q0, rat(-1, 1), "Q0")
	rc0 := decodeExpr(t, f, f.ConeRefs["Rc0"])
	if rc0.Kind != physmath.TagMul {
		t.Fatalf("stored Rc0 is kind %d, want TagMul", rc0.Kind)
	}
	rcPows := childrenOfKind(rc0, physmath.TagPow)
	if len(rcPows) != 1 {
		t.Fatalf("stored Rc0 holds %d Pow children, want 1", len(rcPows))
	}
	checkPow(rcPows[0], rat(2, 1), "Rc0")
	hasR = false
	for _, c := range rc0.Children {
		if c.Kind == physmath.TagSymbol {
			requireBoundR(t, c, rID, "stored Rc0 radial factor")
			hasR = true
		}
	}
	if !hasR {
		t.Fatal("stored Rc0 has no declared-r factor")
	}
	p0 := decodeExpr(t, f, f.ConeRefs["P0"])
	checkPow(p0, rat(2, 1), "P0")

	// Evaluate the DECODED STORED expressions at rational samples;
	// those values feed the identity check.
	for _, rs := range []string{"1/100", "7/500", "3/50", "1/2"} {
		q, ok := new(big.Rat).SetString(rs)
		if !ok {
			t.Fatalf("bad sample %s", rs)
		}
		env := map[string]*big.Rat{"r": q}
		ev := func(e physmath.Expr, what string) *big.Rat {
			t.Helper()
			v, err := evalRat(e, rID, env)
			if err != nil {
				t.Fatalf("stored %s unevaluatable at r=%s: %v", what, rs, err)
			}
			return v
		}
		vv, qq, rc, pp := ev(v0, "V0"), ev(q0, "Q0"), ev(rc0, "Rc0"), ev(p0, "P0")
		lhs := new(big.Rat).Mul(big.NewRat(2, 1), new(big.Rat).Mul(vv, qq))
		q2 := new(big.Rat).Mul(q, q)
		term1 := new(big.Rat).Quo(new(big.Rat).Mul(big.NewRat(3, 1), q2), rc)
		term2 := new(big.Rat).Quo(new(big.Rat).Mul(new(big.Rat).Mul(q2, q), pp),
			new(big.Rat).Mul(rc, rc))
		rhs := new(big.Rat).Sub(term1, term2)
		if lhs.Cmp(rhs) != 0 {
			t.Fatalf("cone identity fails at r=%s: %s != %s", rs, lhs, rhs)
		}
	}

	// Negatives on stored-shape discrimination.
	// Wrong power: an Rc0-shaped artifact with u^3 must fail the
	// exponent gate (proves the gate inspects exponents).
	if _, err := evalRat(*p0.Base, rID, map[string]*big.Rat{"r": big.NewRat(1, 2)}); err != nil {
		t.Fatalf("P0 base unexpectedly unevaluatable: %v", err)
	}
	cube := physmath.Pow(*p0.Base, physmath.MustRational(3, 1))
	x, err := evalRat(cube, rID, map[string]*big.Rat{"r": big.NewRat(1, 2)})
	if err != nil {
		t.Fatalf("cube variant unexpectedly unevaluatable: %v", err)
	}
	sq, err := evalRat(p0, rID, map[string]*big.Rat{"r": big.NewRat(1, 2)})
	if err != nil {
		t.Fatal(err)
	}
	if x.Cmp(sq) == 0 {
		t.Fatal("u^3 variant coincides with stored u^2 at the probe point")
	}
	// Same display name, different parameter identity: rejected
	// before any value is produced.
	r99 := physmath.NewBoundSymbol("r", physmath.DeriveParameterID(99))
	u99 := physmath.Add(physmath.MustRational(1, 1),
		physmath.Neg(physmath.Mul(physmath.MustRational(1, 2), r99)))
	if _, err := evalRat(u99, rID, map[string]*big.Rat{"r": big.NewRat(1, 2)}); err == nil {
		t.Fatal("evaluator accepted a same-display different-identity r")
	}
	// A corrupted coefficient (1/3 for 1/2) breaks the identity,
	// proving the evaluated values carry the stored coefficient.
	badU := physmath.Add(physmath.MustRational(1, 1),
		physmath.Neg(physmath.Mul(physmath.MustRational(1, 3), physmath.NewBoundSymbol("r", rID))))
	bad, err := evalRat(badU, rID, map[string]*big.Rat{"r": big.NewRat(1, 2)})
	if err != nil {
		t.Fatal(err)
	}
	good, err := evalRat(u, rID, map[string]*big.Rat{"r": big.NewRat(1, 2)})
	if err != nil {
		t.Fatal(err)
	}
	if bad.Cmp(good) == 0 {
		t.Fatal("corrupted 1/3 coefficient coincides with stored 1/2")
	}
}

func TestThetaStructure(t *testing.T) {
	f := ltb.Build()
	// Identity semantics (correction): the production H is EPHEMERAL.
	// Ephemeral symbols carry no ParameterID or ConstantID, so their
	// encoded identity is exactly their display string. The test
	// asserts both operands match the same intended stored symbol
	// representation (byte-identical encodings), and documents that
	// this establishes syntactic sameness — not bound-parameter
	// identity, and not physical meaning. A synthetic bound-symbol
	// control below shows the two same-display symbols with different
	// ParameterIDs are distinguished; it does not confer
	// bound-parameter identity on the production artifact.
	theta := decodeExpr(t, f, f.ThetaRef)
	if theta.Kind != physmath.TagAdd || len(theta.Children) != 2 {
		t.Fatalf("theta is not a 2-child Add")
	}
	wantH, err := physmath.NewEphemeralSymbol("H").Encode()
	if err != nil {
		t.Fatal(err)
	}
	var plain *physmath.Expr
	var scaled *physmath.Expr
	for i := range theta.Children {
		if theta.Children[i].Kind == physmath.TagMul {
			scaled = &theta.Children[i]
		} else {
			plain = &theta.Children[i]
		}
	}
	if plain == nil || scaled == nil {
		t.Fatal("theta is not H + scaled-H in the expected shape")
	}
	pb, err := plain.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(pb) != string(wantH) {
		t.Fatal("theta plain operand is not the intended stored H representation")
	}
	if len(scaled.Children) != 2 {
		t.Fatalf("theta scaled term has %d factors, want 2", len(scaled.Children))
	}
	var coef *physmath.Expr
	var hfac *physmath.Expr
	for i := range scaled.Children {
		if scaled.Children[i].Kind == physmath.TagRational {
			coef = &scaled.Children[i]
		} else {
			hfac = &scaled.Children[i]
		}
	}
	if coef == nil || hfac == nil {
		t.Fatal("theta scaled term is not coefficient-times-symbol")
	}
	requireRat(t, *coef, 2, 1, "theta coefficient")
	hb, err := hfac.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(hb) != string(wantH) {
		t.Fatal("theta scaled H differs from the plain H representation")
	}
	if string(hb) != string(pb) {
		t.Fatal("theta H occurrences are not byte-identical")
	}

	// Sigma must structurally represent (1/3)(H-H)^2 with the same H:
	// Mul root, 1/3 factor, Pow(...,2) with an Add(H, Neg(H)) base.
	sigma := decodeExpr(t, f, f.SigmaRef)
	if sigma.Kind != physmath.TagMul {
		t.Fatalf("sigma is kind %d, want TagMul", sigma.Kind)
	}
	var sfac, spow *physmath.Expr
	for i := range sigma.Children {
		if sigma.Children[i].Kind == physmath.TagRational {
			sfac = &sigma.Children[i]
		}
		if sigma.Children[i].Kind == physmath.TagPow {
			spow = &sigma.Children[i]
		}
	}
	if sfac == nil || spow == nil {
		t.Fatal("sigma is not factor-times-power in the expected shape")
	}
	requireRat(t, *sfac, 1, 3, "sigma factor")
	x, err := evalRat(*spow.Exp, physmath.DeriveParameterID(0), nil)
	if err != nil || x.Cmp(rat(2, 1)) != 0 {
		t.Fatal("sigma power is not 2")
	}
	base := *spow.Base
	if base.Kind != physmath.TagAdd || len(base.Children) != 2 {
		t.Fatal("sigma base is not a 2-child Add")
	}
	var bplain, bneg *physmath.Expr
	for i := range base.Children {
		if base.Children[i].Kind == physmath.TagNeg {
			bneg = &base.Children[i]
		} else {
			bplain = &base.Children[i]
		}
	}
	if bplain == nil || bneg == nil || bneg.Operand == nil {
		t.Fatal("sigma base is not H + Neg(H) in the expected shape")
	}
	for _, e := range []*physmath.Expr{bplain, bneg.Operand} {
		eb, err := e.Encode()
		if err != nil {
			t.Fatal(err)
		}
		if string(eb) != string(wantH) {
			t.Fatal("sigma H occurrence differs from the intended H representation")
		}
	}

	// ThetaCanon linkage: the recorded derivation must be the CANON
	// rule on ThetaRef producing ThetaCanon — replay alone would not
	// establish relevance.
	d1 := decodeDerivation(t, f, f.Derivs[1])
	if len(d1.Steps) != 1 {
		t.Fatalf("canon derivation has %d steps, want 1", len(d1.Steps))
	}
	s1 := d1.Steps[0]
	if s1.ProducerNamespace != physmath.NamespacePhysMath || s1.RuleID != physmath.RuleExprCanon {
		t.Fatalf("canon step uses (%q,%q)", s1.ProducerNamespace, s1.RuleID)
	}
	if len(s1.InputRefs) != 1 || s1.InputRefs[0] != f.ThetaRef {
		t.Fatal("canon step input is not ThetaRef")
	}
	if s1.OutputRef != f.ThetaCanon {
		t.Fatal("canon step output is not ThetaCanon")
	}
	// Canonicalization preserves variable identity: every symbol in
	// the decoded output must be the intended H representation.
	canonOut := decodeExpr(t, f, f.ThetaCanon)
	for _, s := range collectSymbols(canonOut) {
		sb, err := s.Encode()
		if err != nil {
			t.Fatal(err)
		}
		if string(sb) != string(wantH) {
			t.Fatal("canon output introduced a foreign symbol identity")
		}
	}

	// Negatives.
	// Different display name: not the intended H.
	other, err := physmath.NewEphemeralSymbol("Hperp").Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(other) == string(wantH) {
		t.Fatal("different display coincides with H representation")
	}
	// Wrong coefficient: 3H+... form must differ from stored theta.
	badTheta := physmath.Add(physmath.NewEphemeralSymbol("H"),
		physmath.Mul(physmath.MustRational(3, 1), physmath.NewEphemeralSymbol("H")))
	bb, err := badTheta.Encode()
	if err != nil {
		t.Fatal(err)
	}
	tb, err := theta.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(bb) == string(tb) {
		t.Fatal("wrong-coefficient theta coincides with stored theta")
	}
	// Same display, different parameter identity (synthetic
	// bound-symbol control): distinguished. This exercises the
	// ParameterID mechanism; it does not give the production
	// (ephemeral) H bound-parameter identity.
	hA := physmath.NewBoundSymbol("H", physmath.DeriveParameterID(7))
	hB := physmath.NewBoundSymbol("H", physmath.DeriveParameterID(8))
	if hA.Equal(hB) {
		t.Fatal("same-display different-PID symbols compare equal")
	}
	ab, _ := hA.Encode()
	bbe, _ := hB.Encode()
	if string(ab) == string(bbe) {
		t.Fatal("same-display different-PID symbols encode identically")
	}
	// Corrupted derivation output must not match ThetaCanon.
	if s1.OutputRef == f.SigmaRef {
		t.Fatal("canon derivation output is the sigma ref")
	}

	// Sigma stays represented-only: no derivation step may consume it.
	for _, dref := range f.Derivs {
		d := decodeDerivation(t, f, dref)
		for _, st := range d.Steps {
			for _, in := range st.InputRefs {
				if in == f.SigmaRef {
					t.Fatal("sigma construction is represented-only but enters a derivation")
				}
			}
		}
	}
}

func collectSymbols(e physmath.Expr) []physmath.Expr {
	var out []physmath.Expr
	var walk func(x physmath.Expr)
	walk = func(x physmath.Expr) {
		switch x.Kind {
		case physmath.TagSymbol:
			out = append(out, x)
		case physmath.TagNeg:
			if x.Operand != nil {
				walk(*x.Operand)
			}
		case physmath.TagPow:
			if x.Base != nil {
				walk(*x.Base)
			}
		}
		for _, c := range x.Children {
			walk(c)
		}
	}
	walk(e)
	return out
}

func TestVerifyRejectsMutatedAssumptions(t *testing.T) {
	// Every mutation below goes through the production f.Verify()
	// path on a freshly built fixture — not a duplicate checker.
	// Missing regime assumption.
	f := ltb.Build()
	f.PhysThm.AssumptionRefs = []artifact.ArtifactRef{f.Assumptions["ltb_gauge"]}
	if err := f.Verify(); err == nil {
		t.Fatal("Verify accepted PhysThm with a missing required assumption")
	}
	// Same-family substitution: regime replaced by target. Family
	// membership alone would admit this; the exact-set contract
	// must reject it.
	g := ltb.Build()
	g.PhysThm.AssumptionRefs = []artifact.ArtifactRef{
		g.Assumptions["ltb_gauge"], g.Assumptions["ltb_target"],
	}
	if err := g.Verify(); err == nil {
		t.Fatal("Verify accepted a same-family assumption substitution")
	}
	// Dangling reference.
	h := ltb.Build()
	h.PhysThm.AssumptionRefs = []artifact.ArtifactRef{
		h.Assumptions["ltb_gauge"], artifact.ArtifactRef{Digest: [32]byte{9}},
	}
	if err := h.Verify(); err == nil {
		t.Fatal("Verify accepted a dangling assumption reference")
	}
	// Wrong family: an expression artifact where an assumption belongs.
	k := ltb.Build()
	k.PhysThm.AssumptionRefs = []artifact.ArtifactRef{
		k.Assumptions["ltb_gauge"], k.MPrimeRef,
	}
	if err := k.Verify(); err == nil {
		t.Fatal("Verify accepted a wrong-family assumption reference")
	}
	// Same mutations on the equation wrapper.
	m := ltb.Build()
	m.EqGauge.AssumptionRefs = []artifact.ArtifactRef{m.Assumptions["ltb_regime"]}
	if err := m.Verify(); err == nil {
		t.Fatal("Verify accepted EqGauge with a missing required assumption")
	}
	n := ltb.Build()
	n.EqGauge.AssumptionRefs = []artifact.ArtifactRef{
		n.Assumptions["ltb_regime"], n.Assumptions["ltb_domain"],
	}
	if err := n.Verify(); err == nil {
		t.Fatal("Verify accepted EqGauge with a same-family substitution")
	}
	// Positive control: the unmutated fixture still verifies (also
	// covered by TestFixtureVerify).
	if err := ltb.Build().Verify(); err != nil {
		t.Fatalf("unmutated fixture rejected: %v", err)
	}
}

func TestClaimStateIntegrity(t *testing.T) {
	f := ltb.Build()
	open := []ltb.ClaimStatus{
		{ID: "ZEH-Q-01", Status: ltb.StatusLiteratureClaim},
		{ID: "ZEH-Q-02", Status: ltb.StatusCandidate},
		{ID: "ZEH-Q-03", Status: ltb.StatusOpenConjecture},
		{ID: "ZEH-Q-04", Status: ltb.StatusDeferred},
	}
	if err := ltb.CheckClaimPromotion(open, f.Derivs); err != nil {
		t.Fatalf("honest open statuses rejected: %v", err)
	}
	// Unsupported promotion of Q-03 to proved must fail.
	promoted := []ltb.ClaimStatus{
		{ID: "ZEH-Q-03", Status: ltb.StatusProved},
	}
	if err := ltb.CheckClaimPromotion(promoted, f.Derivs); err == nil {
		t.Fatal("Q-03 marked proved without evidence was accepted")
	}
	// Citing evidence outside the inventory must fail. Inventory
	// membership is a consistency check only: it does not establish
	// that the cited derivation proves the promoted claim.
	outside := []ltb.ClaimStatus{
		{ID: "ZEH-Q-03", Status: ltb.StatusProved, Evidence: []artifact.ArtifactRef{f.SigmaRef}},
	}
	if err := ltb.CheckClaimPromotion(outside, f.Derivs); err == nil {
		t.Fatal("Q-03 proved with out-of-inventory evidence was accepted")
	}
	if err := ltb.CheckClaimPromotion(
		[]ltb.ClaimStatus{{ID: "ZEH-Q-03", Status: "established"}}, f.Derivs); err == nil {
		t.Fatal("unknown status was accepted")
	}
}

// TestMPrimeStructureAllR establishes dM/dr == 3r^2/2 for ALL r by
// an independent degree-plus-coefficient argument over the decoded
// stored output, written by hand from the Expr semantics (no engine
// calls). The engine's uncollected output form is an Add of
// rational-times-integer-power monomials (including a 0*r^3
// product-rule residue it never drops). Degrees accumulate in
// arbitrary precision (math/big): no int64 narrowing anywhere, so no
// overflow can misclassify a term. The check requires: every Pow is
// a non-negative integer power of the declaration-bound r; every
// other factor is rational; no nonzero-coefficient term has degree
// other than two; and the degree-2 coefficients sum to exactly 3/2.
// Zero-coefficient residues are admitted only with non-negative
// exponents, keeping r=0 well-defined. This is the all-r argument;
// TestMPrimePointwise remains as supplementary finite-sample
// corroboration.
func TestMPrimeStructureAllR(t *testing.T) {
	f := ltb.Build()
	out := decodeExpr(t, f, f.MPrimeRef)
	if out.Kind != physmath.TagAdd {
		t.Fatalf("derivative output is kind %d, want TagAdd sum of monomials", out.Kind)
	}
	rID := declaredParam(t, f)
	two := big.NewInt(2)
	deg2 := new(big.Rat)
	for i, child := range out.Children {
		coeff, deg, err := monomialParts(child, rID)
		if err != nil {
			t.Fatalf("child %d: %v", i, err)
		}
		if coeff.Sign() != 0 && deg.Cmp(two) != 0 {
			t.Fatalf("child %d: nonzero coefficient %s at degree %s", i, coeff, deg)
		}
		if deg.Cmp(two) == 0 {
			deg2.Add(deg2, coeff)
		}
	}
	if deg2.Cmp(big.NewRat(3, 2)) != 0 {
		t.Fatalf("degree-2 coefficient sum = %s, want 3/2", deg2)
	}
}

// declaredParam resolves the radial identity through the stored
// declaring function, so structural checks bind the same variable
// the differentiation targeted (never a bare ordinal or label match).
func declaredParam(t *testing.T, f *ltb.Fixture) physmath.ParameterID {
	t.Helper()
	fam, payload, err := f.Store.Lookup(f.RFnRef)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyFunction {
		t.Fatalf("radial target is %s/%s, not function/1", fam.Namespace, fam.ArtifactType)
	}
	fn, err := physmath.DecodeFunction(artifact.NewDecoder(payload))
	if err != nil {
		t.Fatal(err)
	}
	if len(fn.ParameterIDs) < 1 || len(fn.ParamNames) < 1 || fn.ParamNames[0] != "r" {
		t.Fatalf("radial declaration = %+v", fn)
	}
	return fn.ParameterIDs[0]
}

// monomialParts splits one Mul child into (rational coefficient,
// total r-degree as *big.Int), requiring every factor to be rational
// or a non-negative integer power of the declaration-bound r.
// Exponents are validated as integers BEFORE any narrowing: a
// non-integer or negative exponent fails closed, and degrees
// accumulate in arbitrary precision.
func monomialParts(e physmath.Expr, rID physmath.ParameterID) (*big.Rat, *big.Int, error) {
	if e.Kind != physmath.TagMul {
		return nil, nil, errEval("child is not a Mul monomial")
	}
	coeff := big.NewRat(1, 1)
	deg := new(big.Int)
	for _, c := range e.Children {
		switch c.Kind {
		case physmath.TagRational:
			coeff.Mul(coeff, new(big.Rat).SetFrac(c.Numerator.ToBig(), c.Denominator.ToBig()))
		case physmath.TagPow:
			if c.Base.Kind != physmath.TagSymbol || c.Base.DisplayName != "r" ||
				c.Base.ParamID == nil || *c.Base.ParamID != rID {
				return nil, nil, errEval("power base is not the declared r")
			}
			x, err := evalRat(*c.Exp, rID, nil)
			if err != nil || !x.IsInt() || x.Sign() < 0 {
				return nil, nil, errEval("non-negative-integer exponent required")
			}
			deg.Add(deg, x.Num())
		default:
			return nil, nil, errEval("non-monomial factor")
		}
	}
	return coeff, deg, nil
}

func TestMonomialPartsAdversarial(t *testing.T) {
	// Unit cases proving monomialParts fails closed instead of
	// misclassifying. Synthetic inputs only; the production output
	// is covered by TestMPrimeStructureAllR.
	rID := physmath.DeriveParameterID(0)
	r := physmath.NewBoundSymbol("r", rID)
	mkpow := func(n, d int64) physmath.Expr {
		return physmath.Pow(r, physmath.MustRational(n, d))
	}
	mkterm := func(factors ...physmath.Expr) physmath.Expr {
		return physmath.Expr{Kind: physmath.TagMul, Children: factors}
	}
	// Valid (3/2)r^2 term: accepted with (3/2, 2).
	c, d, err := monomialParts(mkterm(physmath.MustRational(3, 2), mkpow(2, 1)), rID)
	if err != nil || c.Cmp(rat(3, 2)) != 0 || d.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("valid term rejected or misread: %v %v %v", c, d, err)
	}
	// Oversized exponent 2^62: classified exactly, no overflow.
	c, d, err = monomialParts(mkterm(mkpow(1<<62, 1)), rID)
	if err != nil {
		t.Fatalf("oversized exponent rejected instead of classified: %v", err)
	}
	if d.Cmp(new(big.Int).Lsh(big.NewInt(1), 62)) != 0 {
		t.Fatalf("oversized degree misread: %s", d)
	}
	// Degree-sum case: two 2^62 factors accumulate exactly to 2^63.
	c, d, err = monomialParts(mkterm(mkpow(1<<62, 1), mkpow(1<<62, 1)), rID)
	if err != nil {
		t.Fatalf("degree-sum case rejected: %v", err)
	}
	if d.Cmp(new(big.Int).Lsh(big.NewInt(1), 63)) != 0 {
		t.Fatalf("degree sum misread: %s", d)
	}
	// Fractional, negative exponents: rejected.
	if _, _, err := monomialParts(mkterm(mkpow(1, 2)), rID); err == nil {
		t.Fatal("fractional exponent accepted")
	}
	if _, _, err := monomialParts(mkterm(mkpow(-1, 1)), rID); err == nil {
		t.Fatal("negative exponent accepted")
	}
	// Wrong parameter identity despite the same display name.
	r99 := physmath.NewBoundSymbol("r", physmath.DeriveParameterID(99))
	badID := physmath.Expr{Kind: physmath.TagMul, Children: []physmath.Expr{
		physmath.Pow(r99, physmath.MustRational(2, 1)),
	}}
	if _, _, err := monomialParts(badID, rID); err == nil {
		t.Fatal("same-display different-identity base accepted")
	}
	// Non-monomial factor (bare symbol, no power wrapper).
	bare := physmath.Expr{Kind: physmath.TagMul, Children: []physmath.Expr{r}}
	if _, _, err := monomialParts(bare, rID); err == nil {
		t.Fatal("bare-symbol factor accepted as monomial")
	}
	// Non-Mul root.
	if _, _, err := monomialParts(r, rID); err == nil {
		t.Fatal("non-Mul root accepted")
	}
	// Nonzero degree-3 term is classified (caller rejects degree != 2).
	c, d, err = monomialParts(mkterm(physmath.MustRational(1, 1), mkpow(3, 1)), rID)
	if err != nil || d.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("degree-3 term misread: %v %v %v", c, d, err)
	}
}

func TestBuildDeterminism(t *testing.T) {
	a := ltb.Build()
	b := ltb.Build()
	if len(a.Derivs) != len(b.Derivs) {
		t.Fatal("derivation inventory differs across builds")
	}
	for i := range a.Derivs {
		if a.Derivs[i] != b.Derivs[i] {
			t.Fatal("derivation refs nondeterministic across builds")
		}
	}
	if a.MPrimeRef != b.MPrimeRef || a.MathThmRef != b.MathThmRef {
		t.Fatal("fixture refs nondeterministic across builds")
	}
}
