// Package gr is the canonical General Relativity physical-domain library
// (plan9.3 §7; spec §14): the first conforming phys-* framework.
//
// It owns all GR-specific semantics: framework identity, GR assumptions
// (conventions, regularity, model), GR-local algebra, and the framework
// interpretation of mathematical objects. It owns no pure mathematics
// (that is phys-math) and no mechanical identity (that is phys-artifact).
//
// GR-local mathematics that PhysMath cannot represent (trigonometry,
// unknown-function atoms, regime-aware rules) stays here as disjoint
// package-local types with an explicit fallible projection; it never
// extends the PhysMath AST or family registry (spec §15, plan9.3 D1/D9).
package gr
