package bit

import (
	"cmp"
	"fmt"
	"sort"
)

type RuntimeError struct {
	Span    Span
	Message string
	Args    []any
}

func (err RuntimeError) String() string {
	msg := err.Message
	if len(err.Args) > 0 {
		msg = fmt.Sprintf(msg, err.Args...)
	}
	loc := fmt.Sprintf("%s:%s", err.Span.Source().Name(), err.Span.Location().String())
	return fmt.Sprintf("Runtime error: at %s: %s", loc, msg)
}

func (err RuntimeError) Error() string {
	return err.String()
}

type CompilerError struct {
	Span    Span
	Message string
	Args    []any
}

func (err CompilerError) String() string {
	msg := err.Message
	if len(err.Args) > 0 {
		msg = fmt.Sprintf(msg, err.Args...)
	}
	loc := fmt.Sprintf("%s:%s", err.Span.Source().Name(), err.Span.Location().String())
	txt := err.Span.DisplayText(0)
	if len(txt) > 0 {
		txt = fmt.Sprintf("\n\n    | %s", txt)
	}
	return fmt.Sprintf("at %s: %s%s", loc, msg, txt)
}

func (err CompilerError) Error() string {
	return err.String()
}

func SortErrors(errs []error) {
	sort.Slice(errs, func(i, j int) bool {
		errA := errs[i]
		errB := errs[j]
		cmpErrA, okA := errA.(CompilerError)
		cmpErrB, okB := errB.(CompilerError)
		if okA != okB {
			return !okA // non-compilation errors first
		}

		if orderByMessage := !okA; orderByMessage {
			return cmp.Compare(errA.Error(), errB.Error()) < 0
		}

		return cmpErrA.Span.Compare(cmpErrB.Span) < 0
	})
}
