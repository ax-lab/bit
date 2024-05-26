package core

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

func SortErrors(errors []error) {
	slices.SortFunc(errors, ErrorCompare)
}

func ErrorCompare(a, b error) int {
	if a == nil && b == nil {
		return 0
	} else if a == nil {
		return -1
	} else if b == nil {
		return +1
	}

	locA, okA := a.(ErrorAtSpan)
	locB, okB := b.(ErrorAtSpan)
	if okA != okB {
		if okB { // errors without location first
			return -1
		} else {
			return +1
		}
	} else if okA {
		if res := locA.span.Compare(locB.span); res != 0 {
			return res
		}
	}

	return cmp.Compare(a.Error(), b.Error())
}

type ErrorAtSpan struct {
	span  Span
	inner error
}

func Errorf(span Span, msg string, args ...any) error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	err := errors.New(msg)
	if !span.Valid() {
		return err
	}

	return ErrorAtSpan{span, err}
}

func ErrorAt(span Span, err error) error {
	if err == nil {
		return nil
	} else if !span.Valid() {
		return err
	}

	if errAt, ok := err.(ErrorAtSpan); ok && errAt.span.Valid() {
		return errAt
	}

	return ErrorAtSpan{span, err}
}

func (err ErrorAtSpan) Span() Span {
	return err.span
}

func (err ErrorAtSpan) Error() string {
	return err.String()
}

func (err ErrorAtSpan) String() string {
	return fmt.Sprintf("%s: %v", err.span.Location(), err.inner)
}
