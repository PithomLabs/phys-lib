package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// FrameworkArtifact is the GR framework identity artifact (spec §14.1).
// One framework version has exactly one framework identity artifact.
// Family phys-gr/framework/1.
type FrameworkArtifact struct {
	FrameworkID       string
	PackageNamespace  string
	FrameworkVersion  string
	BaseFrameworkRefs []artifact.ArtifactRef // SET
}

// Family returns FamilyFramework.
func (FrameworkArtifact) Family() physmath.FamilyID { return FamilyFramework }

// NewFramework returns the Phase-1 GR framework identity: PackageNamespace
// "phys-gr", empty BaseFrameworkRefs, pinned opaque version.
func NewFramework(frameworkID string) (FrameworkArtifact, error) {
	fw := FrameworkArtifact{
		FrameworkID:      frameworkID,
		PackageNamespace: NamespaceGR,
		FrameworkVersion: FrameworkVersion,
	}
	if err := fw.Validate(); err != nil {
		return FrameworkArtifact{}, err
	}
	return fw, nil
}

// Validate checks identity fields (no reference resolution).
func (f FrameworkArtifact) Validate() error {
	if err := validateIdentifier("FrameworkID", f.FrameworkID); err != nil {
		return err
	}
	if f.PackageNamespace != NamespaceGR {
		return &Error{Op: "Framework.PackageNamespace",
			Err: fmt.Errorf("must be %q, got %q", NamespaceGR, f.PackageNamespace)}
	}
	if f.FrameworkVersion == "" || len(f.FrameworkVersion) > 128 {
		return &Error{Op: "Framework.FrameworkVersion",
			Err: fmt.Errorf("must be non-empty, max 128 bytes")}
	}
	seen := map[artifact.ArtifactRef]bool{}
	for _, r := range f.BaseFrameworkRefs {
		if seen[r] {
			return &Error{Op: "Framework.BaseFrameworkRefs",
				Err: fmt.Errorf("duplicate reference")}
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

// DecodeFrameworkArtifact decodes one canonical framework payload.
func DecodeFrameworkArtifact(d *artifact.Decoder) (FrameworkArtifact, error) {
	id, err := d.String()
	if err != nil {
		return FrameworkArtifact{}, err
	}
	ns, err := d.String()
	if err != nil {
		return FrameworkArtifact{}, err
	}
	ver, err := d.String()
	if err != nil {
		return FrameworkArtifact{}, err
	}
	n, err := d.SequenceCount()
	if err != nil {
		return FrameworkArtifact{}, err
	}
	refs := make([]artifact.ArtifactRef, 0, n)
	for i := 0; i < n; i++ {
		r, err := artifact.DecodeArtifactRef(d)
		if err != nil {
			return FrameworkArtifact{}, err
		}
		refs = append(refs, r)
	}
	fw := FrameworkArtifact{FrameworkID: id, PackageNamespace: ns,
		FrameworkVersion: ver, BaseFrameworkRefs: refs}
	if err := fw.Validate(); err != nil {
		return FrameworkArtifact{}, err
	}
	return fw, nil
}

// Mint stores the canonical payload in a phys-math Store and returns its
// content-addressed reference.
func (f FrameworkArtifact) Mint(store physmath.Store) (artifact.ArtifactRef, error) {
	payload, err := f.Encode()
	if err != nil {
		return artifact.ArtifactRef{}, err
	}
	return store.Put(f.Family(), payload)
}

// NOTE (S-1): fixture2 carries an intentionally independent copy of this
// traversal for test isolation. Do not unify them into a shared helper;
// shared ownership would create a cross-framework dependency. See
// phys-lib/internal/conformance/fixture2/framework.go.
// closureNamespaces resolves the transitive BaseFrameworkRefs closure and
// returns every admitted namespace: the owning namespace plus every
// transitively reachable base namespace. It rejects dangling refs,
// non-framework/1 targets, self-references, and cycles of any length
// (spec §14.9). This single traversal backs both CheckClosure and InClosure
// so the two cannot diverge (H-1/M-3).
func (f FrameworkArtifact) closureNamespaces(store physmath.Store) (map[string]bool, error) {
	namespaces := map[string]bool{f.PackageNamespace: true}
	seenID := map[string]bool{f.FrameworkID + "\x00" + f.PackageNamespace: true}
	visited := map[artifact.ArtifactRef]bool{}
	var visit func(fw FrameworkArtifact) error
	visit = func(fw FrameworkArtifact) error {
		for _, base := range fw.BaseFrameworkRefs {
			fam, raw, err := store.Lookup(base)
			if err != nil {
				return &Error{Op: "Framework.BaseFrameworkRefs",
					Err: fmt.Errorf("dangling reference: %w", err)}
			}
			if fam.ArtifactType != "framework/1" || fam.SchemaVersion != "1" {
				return &Error{Op: "Framework.BaseFrameworkRefs",
					Err: fmt.Errorf("target is not a framework/1 artifact")}
			}
			bfw, err := decodeFrameworkPayload(raw)
			if err != nil {
				return &Error{Op: "Framework.BaseFrameworkRefs", Err: err}
			}
			// Itself = same framework identity (one version, one artifact).
			if bfw.FrameworkID == f.FrameworkID && bfw.PackageNamespace == f.PackageNamespace {
				return &Error{Op: "Framework.BaseFrameworkRefs", Err: ErrSelfFramework}
			}
			id := bfw.FrameworkID + "\x00" + bfw.PackageNamespace
			if seenID[id] {
				return &Error{Op: "Framework.BaseFrameworkRefs",
					Err: fmt.Errorf("cyclic framework dependency")}
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

// CheckClosure validates the transitive BaseFrameworkRefs closure:
// acyclic, all targets framework/1, no self-reference (spec §14.9).
// BaseFrameworkRefs is artifact/data dependency only and never authorizes
// Go imports (D13).
func (f FrameworkArtifact) CheckClosure(store physmath.Store) error {
	_, err := f.closureNamespaces(store)
	return err
}

// InClosure reports whether fam is usable from this framework: its own
// GR-owned families, or families of any framework in the transitive
// BaseFrameworkRefs closure. Unrelated namespaces are rejected (spec §14.9).
// Family-kind gating (assumption vs interpretation) happens at use sites
// via TargetChecks; this function admits namespaces.
func (f FrameworkArtifact) InClosure(store physmath.Store, fam physmath.FamilyID) bool {
	if fam.Namespace == f.PackageNamespace {
		return isGRFamily(fam)
	}
	namespaces, err := f.closureNamespaces(store)
	if err != nil {
		return false
	}
	return namespaces[fam.Namespace]
}

// decodeFrameworkPayload decodes one canonical framework payload.
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
		return fw, fmt.Errorf("gr: trailing bytes in framework payload")
	}
	return fw, nil
}
