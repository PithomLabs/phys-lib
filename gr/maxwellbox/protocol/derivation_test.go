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
	if len(unsup) != 4 {
		t.Fatalf("want 4 unsupported sketch steps, got %v", unsup)
	}
	for _, s := range f.Sketch.Steps {
		if s.Status == "SUPPORTED" && (s.StepID == "s5" || s.StepID == "s6") {
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
