package gr

import (
	"fmt"

	"github.com/PithomLabs/phys-math"
)

// B2: framework-local field-tensor representation with explicit,
// fallible projection. A FieldTensor is GR-local data: indexed components
// as GRExprs. Projection to PhysMath succeeds only per component and only
// for the representable subset (symbols, rationals, Add/Mul/Neg, integer
// Pow). Anything else is reported, never invented: there is no synthetic
// PhysMath identity for tensor structure, trigonometric/coordinate
// functions, or opaque field components (plan9.3 D9).

// FieldTensorBound caps stored field-tensor components (storage bound).
const FieldTensorBound = 1 << 20

// FieldTensor is a named rank-r tensor over a dim-dimensional slot space
// with row-major GRExpr components. Rank/dimension/index geometry are
// GR-local bookkeeping; only projected components touch PhysMath.
type FieldTensor struct {
	Name       string
	Rank       uint16
	Dimension  uint16
	Components []GRExpr // row-major, len == Dimension^Rank
}

// Validate checks shape and identifier grammar. It proves nothing about
// physical meaning or projectability: see ProjectReport.
func (f FieldTensor) Validate() error {
	if err := validateIdentifier("FieldTensor.Name", f.Name); err != nil {
		return err
	}
	if f.Rank < 1 {
		return fmt.Errorf("gr: FieldTensor rank must be >= 1")
	}
	if f.Dimension < 1 {
		return fmt.Errorf("gr: FieldTensor dimension must be >= 1")
	}
	want := uint64(1)
	for i := uint64(0); i < uint64(f.Rank); i++ {
		want *= uint64(f.Dimension)
		if want > FieldTensorBound {
			return fmt.Errorf("gr: FieldTensor shape exceeds bound")
		}
	}
	if uint64(len(f.Components)) != want {
		return fmt.Errorf("gr: FieldTensor Components %d != dim^rank %d",
			len(f.Components), want)
	}
	return nil
}

// ProjectionReport is the per-component outcome of attempting projection.
// Projected[i] holds the PhysMath expression for component i on success;
// Failures lists component indices with reasons. AllProjected is true
// only when every component projected. A report with failures is a
// successful fallibility record, not an error: callers keep the tensor
// GR-local for the failed parts.
type ProjectionReport struct {
	Projected    []physmath.Expr
	Failed       []int
	Reasons      []string
	AllProjected bool
}

// ProjectReport attempts ToPhysMath per component and records the exact
// outcome. It invents nothing: unprojectable components stay local.
func (f FieldTensor) ProjectReport() ProjectionReport {
	rep := ProjectionReport{
		Projected: make([]physmath.Expr, len(f.Components)),
	}
	ok := true
	for i, c := range f.Components {
		e, err := c.ToPhysMath()
		if err != nil {
			ok = false
			rep.Failed = append(rep.Failed, i)
			rep.Reasons = append(rep.Reasons, err.Error())
			continue
		}
		rep.Projected[i] = e
	}
	rep.AllProjected = ok
	return rep
}
