// Package protocol verifies the maxwellbox K1 trace attempt: replayed
// endpoint fragments with theorem-conclusion gates, honest PARTIAL
// limitation record, and mutation resistance. A green suite here means
// the fragment replays and the unsupported remainder is preserved —
// never that K1 is verified end to end (see manifest limitations).
package protocol

import (
	"testing"

	"github.com/PithomLabs/phys-lib/gr/maxwellbox"
	"github.com/PithomLabs/phys-math"
)

func TestFragmentsReplay(t *testing.T) {
	f := maxwellbox.Build()
	if err := f.Verify(); err != nil {
		t.Fatalf("fixture verify: %v", err)
	}
}

func TestOnlyCanonStepsExecuted(t *testing.T) {
	f := maxwellbox.Build()
	for name, d := range map[string]physmath.Derivation{
		"transform": f.TransformDeriv, "mass": f.MassDeriv,
	} {
		if len(d.Steps) != 1 {
			t.Fatalf("%s: want exactly 1 step, got %d", name, len(d.Steps))
		}
		st := d.Steps[0]
		if st.ProducerNamespace != "physmath" || st.RuleID != physmath.RuleExprCanon {
			t.Fatalf("%s: step uses non-whitelisted rule (%q,%q)",
				name, st.ProducerNamespace, st.RuleID)
		}
	}
}

func TestCorruptedOutputFailsReplay(t *testing.T) {
	f := maxwellbox.Build()
	bad := f.TransformDeriv
	bad.Steps = append([]physmath.DerivationStep(nil), f.TransformDeriv.Steps...)
	bad.Steps[0].OutputRef = f.Symbols["gamma"]
	if err := physmath.Replay(f.Store, bad, nil); err == nil {
		t.Fatal("replay accepted a corrupted fragment output")
	}
}

func TestSketchRecordsPartialExecution(t *testing.T) {
	f := maxwellbox.Build()
	unsup := f.Sketch.UnsupportedSteps()
	if len(unsup) != 3 {
		t.Fatalf("want 3 unsupported sketch steps, got %v", unsup)
	}
	for _, want := range []string{"s1", "s2", "s3"} {
		found := false
		for _, got := range unsup {
			if got == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("step %s not in unsupported list %v", want, unsup)
		}
	}
	for _, s := range f.Sketch.Steps {
		// s4 SUPPORTED is covered by TestS4SupportedCitesInference
		// (must cite the flux-inference conclusion); s5/s6 by replay.
		if s.Status == "SUPPORTED" && (s.StepID == "s4" || s.StepID == "s5" || s.StepID == "s6") {
			continue
		}
		if s.Status != "UNSUPPORTED" {
			t.Fatalf("step %s has unexpected status %q", s.StepID, s.Status)
		}
	}
}

func TestBuildDeterminism(t *testing.T) {
	a := maxwellbox.Build()
	b := maxwellbox.Build()
	ra, err := a.SketchDigestForTest()
	if err != nil {
		t.Fatal(err)
	}
	rb, err := b.SketchDigestForTest()
	if err != nil {
		t.Fatal(err)
	}
	if ra != rb {
		t.Fatal("fixture build is nondeterministic")
	}
}

func TestS4SupportedCitesInference(t *testing.T) {
	f := maxwellbox.Build()
	for _, st := range f.Sketch.Steps {
		if st.StepID != "s4" {
			continue
		}
		if st.Status != "SUPPORTED" {
			t.Fatalf("s4 status = %q, want SUPPORTED via flux inference", st.Status)
		}
		if len(st.Refs) != 1 || st.Refs[0] != f.FluxCert.ConclusionRef {
			t.Fatal("s4 does not cite exactly the flux-inference conclusion")
		}
		// Conclusion resolves to the assembled EQ(bulk, caps) relation.
		fam, _, err := f.Store.Lookup(f.FluxCert.ConclusionRef)
		if err != nil {
			t.Fatal(err)
		}
		if fam != physmath.FamilyRelation {
			t.Fatal("s4 conclusion is not relation/1")
		}
	}
	// Determinism: identical evidence re-mints the identical conclusion.
	g := maxwellbox.Build()
	if f.FluxCert.ConclusionRef != g.FluxCert.ConclusionRef {
		t.Fatal("flux inference conclusion nondeterministic across builds")
	}
	if f.FluxCert.CertDigest() != g.FluxCert.CertDigest() {
		t.Fatal("flux certificate digest nondeterministic across builds")
	}
}
