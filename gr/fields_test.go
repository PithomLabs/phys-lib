package gr

import "testing"

// Projection must be per-component honest: algebra projects, trig and
// opaque components fail explicitly with reasons, nothing invented.
func TestFieldTensorProjectionReport(t *testing.T) {
	x := NewGSymbol("x")
	two, err := NewGRational(2, 1)
	if err != nil {
		t.Fatal(err)
	}
	sin := NewGSin(NewGSymbol("tau"))
	unknown := NewGUnknown("T_ab", NewGSymbol("a"), NewGSymbol("b"))
	f := FieldTensor{Name: "probe", Rank: 1, Dimension: 4,
		Components: []GRExpr{x, two, sin, unknown}}
	if err := f.Validate(); err != nil {
		t.Fatalf("valid field tensor rejected: %v", err)
	}
	rep := f.ProjectReport()
	if rep.AllProjected {
		t.Fatal("partially unprojectable tensor reported fully projected")
	}
	if len(rep.Failed) != 2 || rep.Failed[0] != 2 || rep.Failed[1] != 3 {
		t.Fatalf("wrong failure set: %v", rep.Failed)
	}
	if len(rep.Reasons) != 2 {
		t.Fatal("missing failure reasons")
	}
	if rep.Projected[0].Kind == 0 || rep.Projected[1].Kind == 0 {
		t.Fatal("projectable components not projected")
	}
	// Pure algebra projects cleanly.
	g := FieldTensor{Name: "alg", Rank: 1, Dimension: 2,
		Components: []GRExpr{NewGAdd(x, two), NewGMul(x, two)}}
	if rep := g.ProjectReport(); !rep.AllProjected {
		t.Fatalf("algebra tensor failed projection: %v", rep.Reasons)
	}
	bad := FieldTensor{Name: "bad", Rank: 1, Dimension: 3,
		Components: []GRExpr{x, two}}
	if err := bad.Validate(); err == nil {
		t.Fatal("component-count mismatch accepted")
	}
	zero := FieldTensor{Name: "zero", Rank: 0, Dimension: 2,
		Components: []GRExpr{x, two, x, two}}
	if err := zero.Validate(); err == nil {
		t.Fatal("zero rank accepted")
	}
}
