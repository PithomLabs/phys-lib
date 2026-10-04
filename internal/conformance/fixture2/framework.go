package fixture2

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// FrameworkArtifact is the fixture2 framework identity. It mirrors the
// specification's framework contract under the fixture2 namespace; it shares
// no code with phys-lib/gr by design (isolation proof).
type FrameworkArtifact struct {
	FrameworkID       string
	PackageNamespace  string
	FrameworkVersion  string
	BaseFrameworkRefs []artifact.ArtifactRef // SET
}

// NewFramework returns the minimal fixture2 identity.
func NewFramework() (FrameworkArtifact, error) {
	fw := FrameworkArtifact{
		FrameworkID:      "fixture2",
		PackageNamespace: NamespaceFixture2,
		FrameworkVersion: FrameworkVersion,
	}
	if err := fw.Validate(); err != nil {
		return FrameworkArtifact{}, err
	}
	return fw, nil
}

// Validate checks identity fields.
func (f FrameworkArtifact) Validate() error {
	if err := validateID("FrameworkID", f.FrameworkID); err != nil {
		return err
	}
	if f.PackageNamespace != NamespaceFixture2 {
		return fmt.Errorf("fixture2: PackageNamespace must be %q", NamespaceFixture2)
	}
	if f.FrameworkVersion == "" || len(f.FrameworkVersion) > 128 {
		return fmt.Errorf("fixture2: FrameworkVersion must be non-empty, max 128 bytes")
	}
	seen := map[artifact.ArtifactRef]bool{}
	for _, r := range f.BaseFrameworkRefs {
		if seen[r] {
			return fmt.Errorf("fixture2: duplicate BaseFrameworkRef")
		}
		seen[r] = true
	}
	return nil
}

// Encode returns the canonical payload in declaration order.
func (f FrameworkArtifact) Encode() ([]byte, error) {
	id, err := artifact.EncodeString(f.FrameworkID)
	if err != nil {
		return nil, err
	}
	ns, err := artifact.EncodeString(f.PackageNamespace)
	if err != nil {
		return nil, err
	}
	ver, err := artifact.EncodeString(f.FrameworkVersion)
	if err != nil {
		return nil, err
	}
	out := append(append(id, ns...), ver...)
	elems := make([][]byte, len(f.BaseFrameworkRefs))
	for i, r := range f.BaseFrameworkRefs {
		elems[i] = r.Encode()
	}
	set, err := artifact.EncodeSet(elems)
	if err != nil {
		return nil, err
	}
	return append(out, set...), nil
}

// Mint stores the canonical payload and returns its reference.
func (f FrameworkArtifact) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := f.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(FamilyFramework, payload)
}

// CheckClosure validates BaseFrameworkRefs with transitive cycle detection:
// every base must resolve to a */framework/1 artifact of another namespace,
// and the dependency graph must be acyclic (spec §14.9).
func (f FrameworkArtifact) CheckClosure(store physmath.Store) error {
	// Resolve own ref for self-comparison.
	payload, err := f.Encode()
	if err != nil {
		return err
	}
	own, err := artifact.ComputeRef(FamilyFramework.Namespace, FamilyFramework.SchemaVersion,
		FamilyFramework.ArtifactType, payload)
	if err != nil {
		return err
	}
	seenID := map[string]bool{f.FrameworkID + "\x00" + f.PackageNamespace: true}
	visited := map[artifact.ArtifactRef]bool{own: true}
	var visit func(ref artifact.ArtifactRef) error
	visit = func(ref artifact.ArtifactRef) error {
		if visited[ref] {
			return fmt.Errorf("fixture2: cyclic framework dependency")
		}
		visited[ref] = true
		fam, raw, err := store.Lookup(ref)
		if err != nil {
			return fmt.Errorf("fixture2: unresolved base framework: %w", err)
		}
		if fam.ArtifactType != "framework/1" || fam.SchemaVersion != "1" {
			return fmt.Errorf("fixture2: base target is not framework/1")
		}
		base, err := decodeFrameworkPayload(raw)
		if err != nil {
			return err
		}
		// Itself = same framework identity (one version, one artifact).
		if base.FrameworkID == f.FrameworkID && base.PackageNamespace == f.PackageNamespace {
			return fmt.Errorf("fixture2: framework may not reference itself")
		}
		id := base.FrameworkID + "\x00" + base.PackageNamespace
		if seenID[id] {
			return fmt.Errorf("fixture2: cyclic framework dependency")
		}
		seenID[id] = true
		// Recurse into the base's own bases.
		for _, r := range base.BaseFrameworkRefs {
			if err := visit(r); err != nil {
				return err
			}
		}
		delete(visited, ref)
		delete(seenID, id)
		return nil
	}
	for _, base := range f.BaseFrameworkRefs {
		if err := visit(base); err != nil {
			return err
		}
	}
	return nil
}

// InClosure mirrors the dependency-closure rule: own-namespace families or
// families of any framework in the transitive base closure. Used to reject
// unrelated refs. Shares the traversal discipline with CheckClosure.
func (f FrameworkArtifact) InClosure(store physmath.Store, fam physmath.FamilyID) bool {
	if fam.Namespace == f.PackageNamespace {
		return true
	}
	namespaces, err := f.transitiveBaseNamespaces(store)
	if err != nil {
		return false
	}
	return namespaces[fam.Namespace]
}

// transitiveBaseNamespaces returns every transitively reachable base
// namespace via the same cycle-safe traversal CheckClosure enforces.
//
// NOTE (S-1 test-harness duplication): this traversal intentionally mirrors
// gr/framework.go closureNamespaces without sharing code. fixture2 is
// test-only isolation by design; the duplication must remain independent and
// must never be promoted into phys-lib/core or a shared production helper.

func (f FrameworkArtifact) transitiveBaseNamespaces(store physmath.Store) (map[string]bool, error) {
	namespaces := map[string]bool{}
	seenID := map[string]bool{f.FrameworkID + "\x00" + f.PackageNamespace: true}
	visited := map[artifact.ArtifactRef]bool{}
	var visit func(fw FrameworkArtifact) error
	visit = func(fw FrameworkArtifact) error {
		for _, base := range fw.BaseFrameworkRefs {
			_, raw, err := store.Lookup(base)
			if err != nil {
				return err
			}
			bfw, err := decodeFrameworkPayload(raw)
			if err != nil {
				return err
			}
			id := bfw.FrameworkID + "\x00" + bfw.PackageNamespace
			if seenID[id] {
				return fmt.Errorf("fixture2: cyclic framework dependency")
			}
			if visited[base] {
				continue
			}
			visited[base] = true
			seenID[id] = true
			namespaces[bfw.PackageNamespace] = true
			if err := visit(bfw); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(f); err != nil {
		return nil, err
	}
	return namespaces, nil
}

// decodeFrameworkPayload decodes a framework canonical payload.
func decodeFrameworkPayload(raw []byte) (FrameworkArtifact, error) {
	d := artifact.NewDecoder(raw)
	var fw FrameworkArtifact
	var err error
	if fw.FrameworkID, err = d.String(); err != nil {
		return fw, err
	}
	if fw.PackageNamespace, err = d.String(); err != nil {
		return fw, err
	}
	if fw.FrameworkVersion, err = d.String(); err != nil {
		return fw, err
	}
	n, err := d.SequenceCount()
	if err != nil {
		return fw, err
	}
	for i := 0; i < n; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return fw, err
		}
		fw.BaseFrameworkRefs = append(fw.BaseFrameworkRefs, r)
	}
	if !d.Exhausted() {
		return fw, fmt.Errorf("fixture2: trailing bytes in framework")
	}
	return fw, nil
}
