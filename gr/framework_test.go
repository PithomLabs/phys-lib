package gr

import (
	"testing"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

func testStore() physmath.Store { return physmath.NewMemStore() }

// Category-A vector: framework artifact serialization is deterministic and
// carries the pinned namespace/version.
func TestFrameworkIdentity(t *testing.T) {
	fw, err := NewFramework("general_relativity")
	if err != nil {
		t.Fatal(err)
	}
	if fw.PackageNamespace != NamespaceGR {
		t.Fatalf("PackageNamespace = %q, want phys-gr", fw.PackageNamespace)
	}
	if fw.FrameworkVersion != FrameworkVersion || FrameworkVersion != "1" {
		t.Fatal("FrameworkVersion not pinned to \"1\"")
	}
	if len(fw.BaseFrameworkRefs) != 0 {
		t.Fatal("Phase-1 BaseFrameworkRefs not empty")
	}
	s := testStore()
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
	fam, _, err := s.Lookup(r1)
	if err != nil {
		t.Fatal(err)
	}
	if fam != FamilyFramework {
		t.Fatal("minted under wrong family")
	}
	// Round-trip.
	_, payload, _ := s.Lookup(r1)
	back, err := DecodeFrameworkArtifact(artifact.NewDecoder(payload))
	if err != nil {
		t.Fatal(err)
	}
	if back.FrameworkID != "general_relativity" {
		t.Fatal("framework round trip changed identity")
	}
	// Closure: self-reference rejected (same identity, any bases revision).
	self := FrameworkArtifact{FrameworkID: "x", PackageNamespace: NamespaceGR,
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{r1}}
	_ = r1
	selfRef, err := self.Mint(s)
	if err != nil {
		t.Fatal(err)
	}
	_ = selfRef
	selfLoop := FrameworkArtifact{FrameworkID: "x", PackageNamespace: NamespaceGR,
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{mustMintFramework(t, s, "x", NamespaceGR, nil)}}
	if err := selfLoop.CheckClosure(s); err == nil {
		t.Fatal("self framework reference accepted")
	}
	// Closure: non-framework target rejected.
	rel, _ := artifact.ComputeRef("physmath", "1", "relation/1", []byte{0x09})
	if _, err := s.Put(physmath.FamilyRelation, []byte{0x09}); err != nil {
		t.Fatal(err)
	}
	other := FrameworkArtifact{FrameworkID: "x", PackageNamespace: NamespaceGR,
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{rel}}
	if err := other.CheckClosure(s); err == nil {
		t.Fatal("non-framework base reference accepted")
	}
	// Closure: dangling reference rejected.
	dangling, _ := artifact.ComputeRef("phys-other", "1", "framework/1", []byte("ghost"))
	ghost := FrameworkArtifact{FrameworkID: "x", PackageNamespace: NamespaceGR,
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{dangling}}
	if err := ghost.CheckClosure(s); err == nil {
		t.Fatal("dangling base reference accepted")
	}
	// Unrelated-namespace assumption is outside the closure.
	if fw.InClosure(s, physmath.FamilyID{Namespace: "phys-qm", ArtifactType: "assumption/1", SchemaVersion: "1"}) {
		t.Fatal("unrelated namespace inside closure")
	}
	if !fw.InClosure(s, FamilyAssumption) {
		t.Fatal("own assumption family outside closure")
	}
}

// mustMintFramework mints a foreign-namespace framework/1 artifact with
// consistent payload and family (test data, syntactically valid per §5.1).
func mustMintFramework(t *testing.T, s physmath.Store, id, ns string, bases []artifact.ArtifactRef) artifact.ArtifactRef {
	t.Helper()
	fw := FrameworkArtifact{FrameworkID: id, PackageNamespace: ns,
		FrameworkVersion: "1", BaseFrameworkRefs: bases}
	if ns == NamespaceGR {
		if err := fw.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	payload, err := fw.Encode()
	if err != nil {
		t.Fatal(err)
	}
	ref, err := s.Put(physmath.FamilyID{Namespace: ns, ArtifactType: "framework/1", SchemaVersion: "1"}, payload)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// TestClosureTransitive exercises depth 0-3 admission and every cycle class
// (§9.1 AGENTS.md: depth 0/1/2+, dangling, wrong family/namespace, cycles).
func TestClosureTransitive(t *testing.T) {
	s := testStore()
	// Chain: top → A → B → C (all foreign test namespaces).
	refC := mustMintFramework(t, s, "c", "phys-c", nil)
	refB := mustMintFramework(t, s, "b", "phys-b", []artifact.ArtifactRef{refC})
	refA := mustMintFramework(t, s, "a", "phys-a", []artifact.ArtifactRef{refB})
	top := FrameworkArtifact{FrameworkID: "top", PackageNamespace: NamespaceGR,
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{refA}}
	if err := top.CheckClosure(s); err != nil {
		t.Fatalf("valid 3-level chain rejected: %v", err)
	}
	// Depth 0 (self), 1, 2, 3 admission through the deepest base namespace.
	famAt := func(ns string) physmath.FamilyID {
		return physmath.FamilyID{Namespace: ns, ArtifactType: "assumption/1", SchemaVersion: "1"}
	}
	if !top.InClosure(s, famAt(NamespaceGR)) {
		t.Fatal("depth-0 namespace not admitted")
	}
	for _, ns := range []string{"phys-a", "phys-b", "phys-c"} {
		if !top.InClosure(s, famAt(ns)) {
			t.Fatalf("transitive base namespace %s not admitted", ns)
		}
	}
	if top.InClosure(s, famAt("phys-unrelated")) {
		t.Fatal("unrelated namespace admitted")
	}
	// Two-node cycle: X → Y → X.
	refX := mustMintFramework(t, s, "x", "phys-x", nil)
	refY := mustMintFramework(t, s, "y", "phys-y", []artifact.ArtifactRef{refX})
	cycX := FrameworkArtifact{FrameworkID: "x", PackageNamespace: "phys-x",
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{refY}}
	if err := cycX.CheckClosure(s); err == nil {
		t.Fatal("two-node cycle accepted")
	}
	// Three-node transitive cycle: P → Q → R → P.
	refP := mustMintFramework(t, s, "p", "phys-p", nil)
	refQ := mustMintFramework(t, s, "q", "phys-q", []artifact.ArtifactRef{refP})
	refR := mustMintFramework(t, s, "r", "phys-r", []artifact.ArtifactRef{refQ})
	cycP := FrameworkArtifact{FrameworkID: "p", PackageNamespace: "phys-p",
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{refR}}
	if err := cycP.CheckClosure(s); err == nil {
		t.Fatal("three-node transitive cycle accepted")
	}
	// Wrong family base rejected even when resolvable.
	relPayload := []byte{0x09}
	relRef, err := s.Put(physmath.FamilyRelation, relPayload)
	if err != nil {
		t.Fatal(err)
	}
	wrongFam := FrameworkArtifact{FrameworkID: "w", PackageNamespace: NamespaceGR,
		FrameworkVersion: "1", BaseFrameworkRefs: []artifact.ArtifactRef{relRef}}
	if err := wrongFam.CheckClosure(s); err == nil {
		t.Fatal("non-framework base accepted")
	}
}
