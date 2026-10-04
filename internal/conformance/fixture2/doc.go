// Package fixture2 is the minimal second-framework conformance fixture
// (plan9.3 D30/R4; Phase 8). It is TEST-ONLY: never imported by production
// code, never a dependency of core or gr, never a production theory package.
//
// It is nevertheless a FULLY CONFORMING MINIMAL framework under its own
// pinned durable namespace phys-fixture2: a valid FrameworkArtifact plus
// minimal implementations of all six framework-owned families. Its sole
// purpose is proving framework independence: same PhysMath structures,
// different physical assumptions and meaning, no global winner.
//
// This package imports phys-math and phys-artifact ONLY. It must never
// import phys-lib/gr (not even in tests): the independent mathematical
// rebuild below proves the math stands alone.
package fixture2
