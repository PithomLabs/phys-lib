package gr

import "errors"
import "fmt"

// Error wraps a framework validation failure with the failing operation.
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string { return fmt.Sprintf("gr %s: %v", e.Op, e.Err) }
func (e *Error) Unwrap() error { return e.Err }

// Sentinel errors (semantics unchanged; names are implementation detail).
var (
	ErrUnknownCategory    = errors.New("gr: unknown assumption Category")
	ErrSelfFramework      = errors.New("gr: framework may not reference itself")
	ErrUnrelatedFramework = errors.New("gr: reference outside dependency closure")
)
