package gr

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/PithomLabs/phys-math"
)

// ErrUnrepresentableProjection: a GR-local expression has no representation
// in any authorized PhysMath family. Projection fails explicitly; callers
// keep the expression GR-local (plan9.3 D9/R7). No synthetic PhysMath
// function identities are invented for GR-local content.
var ErrUnrepresentableProjection = errors.New("gr: expression not representable in PhysMath")

// GRKind is the GR-local expression node vocabulary. It is deliberately
// disjoint from phys-math tags: no shared interface, no tagged-union
// extension, no common expression abstraction (plan9.3 D9).
type GRKind byte

const (
	GSymbol   GRKind = 0x01
	GRational GRKind = 0x02
	GAdd      GRKind = 0x03
	GMul      GRKind = 0x04
	GNeg      GRKind = 0x05
	GPow      GRKind = 0x06
	GSin      GRKind = 0x07
	GCos      GRKind = 0x08
	GUnknown  GRKind = 0x09
)

// GRExpr is a GR-local mathematical expression. Sin/Cos/Unknown atoms and
// GR-specific simplification live here under the theory-local mechanism
// (spec §15); only the representable subset projects to PhysMath.
type GRExpr struct {
	Kind GRKind

	// Symbol name (display only; carries no PhysMath binding).
	Name string

	// Rational value (exact).
	Num *big.Int
	Den *big.Int

	// Add / Mul children.
	Children []GRExpr

	// Neg operand; Sin / Cos argument.
	Operand *GRExpr

	// Pow base and integer exponent.
	Base *GRExpr
	Exp  int

	// Unknown-function atom: opaque name plus argument expressions.
	FuncName string
	FuncArgs []GRExpr
}

// Constructors (package-local algebra; validation is structural only).

// NewGSymbol builds a GR-local named symbol.
func NewGSymbol(name string) GRExpr { return GRExpr{Kind: GSymbol, Name: name} }

// NewGRational builds an exact GR-local rational (denominator must be nonzero).
func NewGRational(num, den int64) (GRExpr, error) {
	if den == 0 {
		return GRExpr{}, fmt.Errorf("gr: zero rational denominator")
	}
	return GRExpr{Kind: GRational, Num: big.NewInt(num), Den: big.NewInt(den)}, nil
}

// NewGAdd builds a GR-local sum.
func NewGAdd(children ...GRExpr) GRExpr {
	return GRExpr{Kind: GAdd, Children: append([]GRExpr{}, children...)}
}

// NewGMul builds a GR-local product.
func NewGMul(children ...GRExpr) GRExpr {
	return GRExpr{Kind: GMul, Children: append([]GRExpr{}, children...)}
}

// NewGNeg builds a GR-local negation.
func NewGNeg(operand GRExpr) GRExpr {
	o := operand
	return GRExpr{Kind: GNeg, Operand: &o}
}

// NewGPow builds a GR-local integer power.
func NewGPow(base GRExpr, exp int) GRExpr {
	b := base
	return GRExpr{Kind: GPow, Base: &b, Exp: exp}
}

// NewGSin builds a GR-local sine application (never projected).
func NewGSin(arg GRExpr) GRExpr {
	a := arg
	return GRExpr{Kind: GSin, Operand: &a}
}

// NewGCos builds a GR-local cosine application (never projected).
func NewGCos(arg GRExpr) GRExpr {
	a := arg
	return GRExpr{Kind: GCos, Operand: &a}
}

// NewGUnknown builds a GR-local opaque function atom such as A(r): the shape
// PhysMath cannot durably express (plan9.3 D1). It is never projected.
func NewGUnknown(name string, args ...GRExpr) GRExpr {
	return GRExpr{Kind: GUnknown, FuncName: name, FuncArgs: append([]GRExpr{}, args...)}
}

// ToPhysMath projects the representable subset into a PhysMath expression:
// symbols (as ephemeral display symbols — binding is the caller's fixture
// responsibility), exact rationals, Add/Mul/Neg, and integer Pow. Sin, Cos,
// and Unknown atoms fail explicitly with ErrUnrepresentableProjection.
// There is deliberately no FromPhysMath: projection is one-way.
func (g GRExpr) ToPhysMath() (physmath.Expr, error) {
	switch g.Kind {
	case GSymbol:
		return physmath.NewEphemeralSymbol(g.Name), nil
	case GRational:
		if g.Den == nil || g.Den.Sign() == 0 {
			return physmath.Expr{}, fmt.Errorf("%w: zero denominator", ErrUnrepresentableProjection)
		}
		e, err := physmath.NewRational(g.Num, g.Den)
		if err != nil {
			return physmath.Expr{}, fmt.Errorf("%w: %v", ErrUnrepresentableProjection, err)
		}
		return e, nil
	case GAdd:
		children := make([]physmath.Expr, len(g.Children))
		for i, c := range g.Children {
			pc, err := c.ToPhysMath()
			if err != nil {
				return physmath.Expr{}, err
			}
			children[i] = pc
		}
		return physmath.Add(children...), nil
	case GMul:
		children := make([]physmath.Expr, len(g.Children))
		for i, c := range g.Children {
			pc, err := c.ToPhysMath()
			if err != nil {
				return physmath.Expr{}, err
			}
			children[i] = pc
		}
		return physmath.Mul(children...), nil
	case GNeg:
		o, err := g.Operand.ToPhysMath()
		if err != nil {
			return physmath.Expr{}, err
		}
		return physmath.Neg(o), nil
	case GPow:
		b, err := g.Base.ToPhysMath()
		if err != nil {
			return physmath.Expr{}, err
		}
		x, err := physmath.NewRational(big.NewInt(int64(g.Exp)), big.NewInt(1))
		if err != nil {
			return physmath.Expr{}, fmt.Errorf("%w: %v", ErrUnrepresentableProjection, err)
		}
		return physmath.Pow(b, x), nil
	case GSin, GCos, GUnknown:
		return physmath.Expr{}, fmt.Errorf("%w: GR-local %s has no PhysMath family",
			ErrUnrepresentableProjection, g.Kind)
	default:
		return physmath.Expr{}, fmt.Errorf("%w: unknown GR node", ErrUnrepresentableProjection)
	}
}

// String names the GR node kind (display only, never identity).
func (k GRKind) String() string {
	switch k {
	case GSymbol:
		return "Symbol"
	case GRational:
		return "Rational"
	case GAdd:
		return "Add"
	case GMul:
		return "Mul"
	case GNeg:
		return "Neg"
	case GPow:
		return "Pow"
	case GSin:
		return "Sin"
	case GCos:
		return "Cos"
	case GUnknown:
		return "Unknown"
	default:
		return "Invalid"
	}
}
