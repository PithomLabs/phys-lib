package gr

import (
	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// NamespaceGR is the durable GR framework namespace. It is pinned and
// unchanged by the Go package location (plan9.3 D12): phys-lib/gr is the
// package, phys-gr is the durable namespace.
const NamespaceGR = "phys-gr"

// SchemaVersion1 is the single Phase-1 schema version.
const SchemaVersion1 = "1"

// FrameworkVersion is the pinned opaque package-owned version string for the
// Phase-1 GR reference fixture (plan9.3 D24). It carries no semver ontology.
const FrameworkVersion = "1"

// GR durable family identifiers (spec §5.6: six families per conforming
// framework namespace).
var (
	FamilyFramework      = physmath.FamilyID{Namespace: NamespaceGR, ArtifactType: "framework/1", SchemaVersion: SchemaVersion1}
	FamilyAssumption     = physmath.FamilyID{Namespace: NamespaceGR, ArtifactType: "assumption/1", SchemaVersion: SchemaVersion1}
	FamilyDefinition     = physmath.FamilyID{Namespace: NamespaceGR, ArtifactType: "definition/1", SchemaVersion: SchemaVersion1}
	FamilyEquation       = physmath.FamilyID{Namespace: NamespaceGR, ArtifactType: "equation/1", SchemaVersion: SchemaVersion1}
	FamilyTheorem        = physmath.FamilyID{Namespace: NamespaceGR, ArtifactType: "theorem/1", SchemaVersion: SchemaVersion1}
	FamilyInterpretation = physmath.FamilyID{Namespace: NamespaceGR, ArtifactType: "interpretation/1", SchemaVersion: SchemaVersion1}
)

// AllGRFamilies returns the six GR-owned durable families.
func AllGRFamilies() []physmath.FamilyID {
	return []physmath.FamilyID{
		FamilyFramework, FamilyAssumption, FamilyDefinition,
		FamilyEquation, FamilyTheorem, FamilyInterpretation,
	}
}

// validateIdentifier enforces the free-form identifier grammar (spec §5.4),
// shared with the namespace grammar.
func validateIdentifier(what, id string) error {
	if err := artifact.ValidateNamespace(id); err != nil {
		return &Error{Op: what, Err: err}
	}
	return nil
}

// isGRFamily reports whether fam is one of the six GR-owned families.
func isGRFamily(fam physmath.FamilyID) bool {
	for _, f := range AllGRFamilies() {
		if fam == f {
			return true
		}
	}
	return false
}
