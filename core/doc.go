// Package core is the theory-agnostic physical common denominator of the
// phys-lib module (plan9.3 §§6, 8; D5, D6, D20).
//
// Phase 1: intentionally empty. A capability belongs here only when its
// physical meaning is independent of any particular theory, its semantics
// are stable across frameworks, its API encodes no theory-specific
// assumptions, and multi-framework consumer evidence exists. Promotion
// requires a classification record under ../classifications/ with the full
// evidence fields; "useful to us" is insufficient. Ambiguous candidates stay
// in their specific framework package (default: phys-lib/gr).
//
// An empty core at Phase-1 completion is a conforming outcome, not a gap.
package core
