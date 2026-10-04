package fixture2

import (
	"github.com/PithomLabs/phys-artifact"
	"github.com/PithomLabs/phys-math"
)

// NamespaceFixture2 is the pinned durable test namespace (plan9.3 D12).
// Invariant across runs and implementations; never phys-gr.
const NamespaceFixture2 = "phys-fixture2"

// FrameworkVersion is the pinned opaque version string (plan9.3 D24).
const FrameworkVersion = "1"

// Fixture2 durable family identifiers: all six framework-owned families.
var (
	FamilyFramework      = physmath.FamilyID{Namespace: NamespaceFixture2, ArtifactType: "framework/1", SchemaVersion: "1"}
	FamilyAssumption     = physmath.FamilyID{Namespace: NamespaceFixture2, ArtifactType: "assumption/1", SchemaVersion: "1"}
	FamilyDefinition     = physmath.FamilyID{Namespace: NamespaceFixture2, ArtifactType: "definition/1", SchemaVersion: "1"}
	FamilyEquation       = physmath.FamilyID{Namespace: NamespaceFixture2, ArtifactType: "equation/1", SchemaVersion: "1"}
	FamilyTheorem        = physmath.FamilyID{Namespace: NamespaceFixture2, ArtifactType: "theorem/1", SchemaVersion: "1"}
	FamilyInterpretation = physmath.FamilyID{Namespace: NamespaceFixture2, ArtifactType: "interpretation/1", SchemaVersion: "1"}
)

func validateID(what, id string) error {
	if err := artifact.ValidateNamespace(id); err != nil {
		return err
	}
	return nil
}
