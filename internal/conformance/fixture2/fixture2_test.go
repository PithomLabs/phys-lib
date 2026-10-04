package fixture2

import (
	"encoding/hex"
	"os/exec"
	"strings"
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// The gr equivariance fixture's final digest, copied as a VALUE (ported
// vector discipline, D4/D18) — never via import. Identical canonical math
// must produce this identical digest under an unrelated framework.
const grFixtureFinalDigest = "6b7526f1c90be7958d86bb346abb378d698c2c7d97f9ce147d4c1b2b87658be9"

func TestFrameworkMinimal(t *testing.T) {
	fw, err := NewFramework()
	if err != nil {
		t.Fatal(err)
	}
	if fw.PackageNamespace != NamespaceFixture2 {
		t.Fatalf("namespace = %q, want phys-fixture2", fw.PackageNamespace)
	}
	if fw.FrameworkVersion != "1" {
		t.Fatal("version not pinned")
	}
	s := physmath.NewMemStore()
	r1, err := fw.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := fw.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	if r1 != r2 {
		t.Fatal("framework identity not deterministic")
	}
	if r1.Namespace != NamespaceFixture2 {
		t.Fatal("framework minted under wrong namespace")
	}
	if err := fw.CheckClosure(s); err != nil {
		t.Fatal(err)
	}
}

func TestIndependentMathRebuild(t *testing.T) {
	fx, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	// Same mathematics: identical canonical bytes → identical digest.
	if got := hex.EncodeToString(fx.FinalRef.Digest[:]); got != grFixtureFinalDigest {
		t.Fatalf("math identity differs from GR fixture final:\n got %s\nwant %s", got, grFixtureFinalDigest)
	}
	// Deterministic rebuild.
	fx2, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	if fx.FinalRef != fx2.FinalRef || fx.TheoremFWRef != fx2.TheoremFWRef {
		t.Fatal("rebuild not deterministic")
	}
}

func TestDistinctFrameworkIdentities(t *testing.T) {
	fx, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	// Every framework artifact lives under phys-fixture2 — never phys-gr.
	for name, ref := range map[string]artifact.ArtifactRef{
		"framework": fx.FrameworkRef, "assumption": fx.AssumptionRef,
		"definition": fx.DefinitionRef, "equation": fx.EquationFWRef,
		"interpretation": fx.InterpretationRef, "theorem": fx.TheoremFWRef,
	} {
		if ref.Namespace != NamespaceFixture2 {
			t.Fatalf("%s under %q", name, ref.Namespace)
		}
		if ref.Namespace == "phys-gr" {
			t.Fatalf("%s leaked into phys-gr namespace", name)
		}
	}
	// Category is framework-owned: the GR-closed vocabulary does not apply.
	// "test-postulate" is valid here and would be rejected by gr validators.
	asm := Assumption{AssumptionID: "x", Category: "test-postulate", Description: "d"}
	if err := asm.Validate(); err != nil {
		t.Fatalf("framework-owned Category rejected: %v", err)
	}
	// No global winner exists: there is no comparison API between frameworks.
	// Asserted structurally — distinct refs, distinct namespaces, shared math.
	if fx.AssumptionRef.Namespace == "phys-gr" {
		t.Fatal("assumption namespace collision")
	}
}

func TestClosureRejections(t *testing.T) {
	s := physmath.NewMemStore()
	fw, err := NewFramework()
	if err != nil {
		t.Fatal(err)
	}
	ownRef, err := fw.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	fams := map[artifact.ArtifactRef]physmath.FamilyID{ownRef: FamilyFramework}
	_ = fams
	// Self-reference rejected.
	self := fw
	self.BaseFrameworkRefs = []artifact.ArtifactRef{ownRef}
	if err := self.CheckClosure(s); err == nil {
		t.Fatal("self framework reference accepted")
	}
	// Unrelated namespace rejected by InClosure.
	otherAsm := physmath.FamilyID{Namespace: "phys-gr", ArtifactType: "assumption/1", SchemaVersion: "1"}
	if fw.InClosure(s, otherAsm) {
		t.Fatal("unrelated phys-gr assumption inside fixture2 closure")
	}
	ownAsm := physmath.FamilyID{Namespace: NamespaceFixture2, ArtifactType: "assumption/1", SchemaVersion: "1"}
	if !fw.InClosure(s, ownAsm) {
		t.Fatal("own assumption outside closure")
	}
	// Cross-namespace bases are the conforming case: a base framework/1 in
	// another (syntactically valid) namespace is accepted, and two frameworks
	// may share it (diamond, not cycle). Same-ref self-reference stays rejected.
	// A cross-namespace base is minted directly under its own family so
	// payload and family stay consistent (no fixture2 API mints foreign
	// namespaces — this is test data, valid syntactically per §5.1).
	basePayload, err := FrameworkArtifact{FrameworkID: "shared-base",
		PackageNamespace: "phys-other", FrameworkVersion: "1"}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	baseRef, err := s.Put(physmath.FamilyID{Namespace: "phys-other",
		ArtifactType: "framework/1", SchemaVersion: "1"}, basePayload)
	if err != nil {
		t.Fatal(err)
	}
	left, _ := NewFramework()
	left.FrameworkID = "left"
	left.BaseFrameworkRefs = []artifact.ArtifactRef{baseRef}
	if err := left.CheckClosure(s); err != nil {
		t.Fatalf("cross-namespace base rejected: %v", err)
	}
	right, _ := NewFramework()
	right.FrameworkID = "right"
	right.BaseFrameworkRefs = []artifact.ArtifactRef{baseRef}
	if err := right.CheckClosure(s); err != nil {
		t.Fatalf("shared base rejected: %v", err)
	}
	// InClosure admits the base namespace transitively.
	otherFam := physmath.FamilyID{Namespace: "phys-other", ArtifactType: "assumption/1", SchemaVersion: "1"}
	if !left.InClosure(s, otherFam) {
		t.Fatal("base-namespace family outside closure")
	}
	// Depth-2 parity: C bases→B, B bases→A-adjacent chain admitted at depth 2.
	deepCpayload, err := FrameworkArtifact{FrameworkID: "deep-c",
		PackageNamespace: "phys-deep-c", FrameworkVersion: "1"}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	deepC, err := s.Put(physmath.FamilyID{Namespace: "phys-deep-c",
		ArtifactType: "framework/1", SchemaVersion: "1"}, deepCpayload)
	if err != nil {
		t.Fatal(err)
	}
	deepBpayload, err := FrameworkArtifact{FrameworkID: "deep-b",
		PackageNamespace: "phys-deep-b", FrameworkVersion: "1",
		BaseFrameworkRefs: []artifact.ArtifactRef{deepC}}.Encode()
	if err != nil {
		t.Fatal(err)
	}
	deepB, err := s.Put(physmath.FamilyID{Namespace: "phys-deep-b",
		ArtifactType: "framework/1", SchemaVersion: "1"}, deepBpayload)
	if err != nil {
		t.Fatal(err)
	}
	deepTop, _ := NewFramework()
	deepTop.FrameworkID = "deep-top"
	deepTop.BaseFrameworkRefs = []artifact.ArtifactRef{deepB}
	if err := deepTop.CheckClosure(s); err != nil {
		t.Fatalf("depth-2 chain rejected: %v", err)
	}
	deepFam := physmath.FamilyID{Namespace: "phys-deep-c",
		ArtifactType: "assumption/1", SchemaVersion: "1"}
	if !deepTop.InClosure(s, deepFam) {
		t.Fatal("depth-2 base namespace not admitted")
	}
	// Non-framework base target rejected.
	rel, _ := artifact.ComputeRef("physmath", "1", "relation/1", []byte{0x09})
	weird, _ := NewFramework()
	weird.FrameworkID = "weird"
	weird.BaseFrameworkRefs = []artifact.ArtifactRef{rel}
	if err := weird.CheckClosure(s); err == nil {
		t.Fatal("non-framework base accepted")
	}
}

// TestIsolation proves the dependency direction: core and gr resolve without
// internal/conformance, and fixture2 resolves to stdlib + math + artifact only.
func TestIsolation(t *testing.T) {
	for _, pkg := range []string{
		"github.com/PithomLabs/phys-lib/core",
		"github.com/PithomLabs/phys-lib/gr",
	} {
		out, err := exec.Command("go", "list", "-deps", pkg).Output()
		if err != nil {
			t.Fatalf("go list -deps %s: %v", pkg, err)
		}
		for _, dep := range strings.Fields(string(out)) {
			if strings.Contains(dep, "/internal/conformance") {
				t.Fatalf("%s depends on conformance fixture: %s", pkg, dep)
			}
		}
	}
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	for _, dep := range strings.Fields(string(out)) {
		if strings.Contains(dep, "phys-lib/gr") {
			t.Fatalf("fixture2 imports phys-lib/gr: %s", dep)
		}
		if strings.Contains(dep, "PithomLabs/phys-gr") && !strings.Contains(dep, "phys-math") &&
			!strings.Contains(dep, "phys-artifact") && !strings.Contains(dep, "phys-lib") {
			t.Fatalf("legacy/frozen import: %s", dep)
		}
		first := strings.Split(dep, "/")[0]
		if strings.Contains(first, ".") {
			switch {
			case strings.HasPrefix(dep, "github.com/PithomLabs/phys-lib/internal/"):
			case strings.HasPrefix(dep, "github.com/PithomLabs/phys-math"):
			case strings.HasPrefix(dep, "github.com/PithomLabs/phys-artifact"):
			default:
				t.Fatalf("unexpected dependency: %s", dep)
			}
		}
	}
}
