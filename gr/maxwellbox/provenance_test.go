package maxwellbox

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

func TestFixtureVolumesProven(t *testing.T) {
	f := Build()
	for name, ref := range map[string]artifact.ArtifactRef{
		"MinkowskiVol": f.MinkowskiVol, "RindlerVol": f.RindlerVol, "RindlerVolRed": f.RindlerVolRed,
	} {
		if err := RequireVolumeProvenance(f.Store, ref); err != nil {
			t.Fatalf("fixture %s rejected by consumer gate: %v", name, err)
		}
	}
}

func TestConsumerRejectsCrossMetric(t *testing.T) {
	f := Build()
	// Structurally valid record pairing the Rindler metric with the
	// Minkowski determinant + Minkowski derivation.
	cross := physmath.VolumeDensity{MetricRef: f.RindlerRef, DeterminantRef: f.MinkowskiDet,
		DeterminantDerivationRef: f.MinkowskiDetDeriv,
		SignatureClass:           physmath.SignatureLorentzian, FactorRef: f.Symbols["abs_xbar"]}
	// abs_xbar symbol is not in Symbols map; mint a stand-in factor.
	factor, err := physmath.PutExpression(f.Store, physmath.NewEphemeralSymbol("abs_xbar"))
	if err != nil {
		t.Fatal(err)
	}
	cross.FactorRef = factor
	if err := cross.Validate(); err != nil {
		t.Fatalf("cross-metric record should be structurally valid: %v", err)
	}
	raw, err := cross.Encode()
	if err != nil {
		t.Fatal(err)
	}
	cref, err := f.Store.Put(physmath.FamilyVolumeDensity, raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireVolumeProvenance(f.Store, cref); err == nil {
		t.Fatal("consumer gate accepted a cross-metric volume")
	}
}

func TestConsumerRejectsSubstitutedOutput(t *testing.T) {
	f := Build()
	fam, payload, err := f.Store.Lookup(f.MinkowskiVol)
	if err != nil {
		t.Fatal(err)
	}
	if fam != physmath.FamilyVolumeDensity {
		t.Fatal("fixture vol is not volume_density/1")
	}
	v, err := physmath.DecodeVolumeDensity(payload)
	if err != nil {
		t.Fatal(err)
	}
	// Substitute the determinant output after the derivation was produced.
	one, err := physmath.PutExpression(f.Store, physmath.MustRational(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	v.DeterminantRef = one
	raw, err := v.Encode()
	if err != nil {
		t.Fatal(err)
	}
	sref, err := f.Store.Put(physmath.FamilyVolumeDensity, raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireVolumeProvenance(f.Store, sref); err == nil {
		t.Fatal("consumer gate accepted a substituted determinant reference")
	}
}

func TestConsumerRejectsWrongFamily(t *testing.T) {
	f := Build()
	// An expression/1 artifact is not a volume, however valid.
	if err := RequireVolumeProvenance(f.Store, f.MinkowskiDet); err == nil {
		t.Fatal("consumer gate accepted a non-volume reference")
	}
}
