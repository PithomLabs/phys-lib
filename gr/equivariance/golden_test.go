package equivariance

import (
	"encoding/hex"
	"testing"

	"github.com/PithomLabs/phys-artifact"
)

// Golden ArtifactRef digests (hex). Any change in canonical bytes, display
// names, ParameterID framing, or fixture structure alters these values and
// requires a fixture-version change — never a silent bytes drift.
var goldenDigests = map[string]string{
	"F":               "f3c36f6387f0a90c011c934b4f0b103fa373660587aca476fcc94094401cfc75",
	"RhoX":            "eab53ca00cbf854d018fe866a9bff55d2a42d81e1326af16d597e46fbe1ee1dd",
	"RhoY":            "eab53ca00cbf854d018fe866a9bff55d2a42d81e1326af16d597e46fbe1ee1dd",
	"LHS0":            "a89b3685657f1aa03c7303637240a32737d463373f9421610d9f7b2d673d891b",
	"LHS1":            "f93e289770834592ad01fbd06d137370fa6d0e0de6d3aede672ab165439b24a9",
	"RHS0":            "1ef7a69401eff27377ce7cc143b908f9ec2a812a3b12e22d9e7a7aaed94a2f3b",
	"RHS1":            "38c0c07b22b434b01f0612511c6bcc6132dc6c68e2f5fb3592b63646e5d7d6d0",
	"Step0":           "190c417d4df580b9caf412d94aa37d1476f5c1b5d91bb5f6442a71e8c4a23fb6",
	"Step1":           "4b0e614d4fc04dc7d0153b776ffd082deaa8e60e3deb0e82cb0d674579ec85b9",
	"Step2":           "6b7526f1c90be7958d86bb346abb378d698c2c7d97f9ce147d4c1b2b87658be9",
	"Step3":           "8f56b60b746efd32aebd7044ac7dd3372069b4d9bd4db5bbee36fc8861a76b95",
	"Step4":           "e585aab3a9c570e52251aa6adcb818b78fd1d69ea675f5cfd8d102db9dda94da",
	"Step5":           "6b7526f1c90be7958d86bb346abb378d698c2c7d97f9ce147d4c1b2b87658be9",
	"Relation":        "84319104fa5c2e3ddd6f052e80b57ec8b04344cef4eaed5f22aee7295e3c4da9",
	"Equation":        "ba18cb4d76adcdb362e5346c0c388f2d73926a47063b8e6a5952eeeae056db83",
	"Theorem":         "4239a9353df2965138f4b8ed8c576a17641b98dc1713dd8b5bff732fc6a6d95f",
	"Derivation":      "5e4c735cd45cbb9401a323224f2269cc156d913d505960fec8a570d751d77258",
	"Framework":       "ca34a772616c727feca62dddfb6c6bd8a1e0fd99259a724f6bd9f6a107e66f92",
	"Assumption":      "47624043855c66b6902bf61f340e8e7e85ece13c6d8c6ed99376f4c519d2d6c4",
	"Interpretation":  "5ff22f22ac1d2080b1fdb88a40a6e21e63dfc6041bc871d89e6a5ab03dd4a41d",
	"PhysicalTheorem": "1e335a9ae2495c96a480ce873a5e4ae9c1b7493bb6473ddca0fed910cc9a06ce",
}

func digestHex(ref artifact.ArtifactRef) string { return hex.EncodeToString(ref.Digest[:]) }

func TestGoldenRefs(t *testing.T) {
	fx, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{
		"F": digestHex(fx.F), "RhoX": digestHex(fx.RhoX), "RhoY": digestHex(fx.RhoY),
		"LHS0": digestHex(fx.LHS0), "LHS1": digestHex(fx.LHS1),
		"RHS0": digestHex(fx.RHS0), "RHS1": digestHex(fx.RHS1),
		"Step0": digestHex(fx.Step[0]), "Step1": digestHex(fx.Step[1]),
		"Step2": digestHex(fx.Step[2]), "Step3": digestHex(fx.Step[3]),
		"Step4": digestHex(fx.Step[4]), "Step5": digestHex(fx.Step[5]),
		"Relation": digestHex(fx.RelationRef), "Equation": digestHex(fx.EquationRef),
		"Theorem": digestHex(fx.TheoremRef), "Derivation": digestHex(fx.DerivationRef),
		"Framework": digestHex(fx.FrameworkRef), "Assumption": digestHex(fx.AssumptionRef),
		"Interpretation":  digestHex(fx.InterpretationRef),
		"PhysicalTheorem": digestHex(fx.PhysicalTheoremRef),
	}
	if len(got) != len(goldenDigests) {
		t.Fatalf("golden set size %d != %d", len(got), len(goldenDigests))
	}
	for k, want := range goldenDigests {
		g, ok := got[k]
		if !ok {
			t.Fatalf("missing golden %q", k)
		}
		if g != want {
			t.Fatalf("golden %s changed:\n got %s\nwant %s", k, g, want)
		}
	}
	// rho_X/rho_Y coincidence and final equality are structural, not accidental.
	if fx.RhoX != fx.RhoY {
		t.Fatal("rho_X/rho_Y coincidence lost")
	}
	if fx.Step[2] != fx.Step[5] {
		t.Fatal("final byte equality lost")
	}
	// Namespaces pinned (D12).
	for _, r := range []artifact.ArtifactRef{fx.F, fx.Step[0]} {
		if r.Namespace != "physmath" {
			t.Fatalf("math namespace drifted: %q", r.Namespace)
		}
	}
	if fx.FrameworkRef.Namespace != "phys-gr" {
		t.Fatalf("framework namespace drifted: %q", fx.FrameworkRef.Namespace)
	}
}

func TestBuildDeterminism(t *testing.T) {
	a, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	refs := func(fx *Fixture) []artifact.ArtifactRef {
		out := []artifact.ArtifactRef{fx.F, fx.RhoX, fx.LHS0, fx.LHS1, fx.RHS0, fx.RHS1,
			fx.RelationRef, fx.EquationRef, fx.TheoremRef, fx.DerivationRef,
			fx.FrameworkRef, fx.AssumptionRef, fx.InterpretationRef, fx.PhysicalTheoremRef}
		out = append(out, fx.Step[:]...)
		return out
	}
	ra, rb := refs(a), refs(b)
	for i := range ra {
		if ra[i] != rb[i] {
			t.Fatalf("ref %d nondeterministic across builds", i)
		}
	}
}
