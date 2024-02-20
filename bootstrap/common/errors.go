package common

import (
	"cmp"
	"fmt"
	"os"
	"sort"
	"strings"
)

const MaxErrorOutput = 16

type ErrorWithLocation struct {
	Span    Span
	Message string
	Args    []any
}

func (err ErrorWithLocation) String() string {
	msg := err.Message
	if len(err.Args) > 0 {
		msg = fmt.Sprintf(msg, err.Args...)
	}
	loc := fmt.Sprintf("%s:%s", err.Span.Source().Name(), err.Span.Location().String())
	txt := err.Span.DisplayText(60)
	if len(txt) > 0 {
		txt = fmt.Sprintf("\n\n    | %s", txt)
	}
	return fmt.Sprintf("at %s: %s%s", loc, msg, txt)
}

func (err ErrorWithLocation) Error() string {
	return err.String()
}

func SortErrors(errs []error) {
	sort.Slice(errs, func(i, j int) bool {
		errA := errs[i]
		errB := errs[j]
		cmpErrA, okA := errA.(ErrorWithLocation)
		cmpErrB, okB := errB.(ErrorWithLocation)
		if okA != okB {
			return !okA // non-compilation errors first
		}

		if orderByMessage := !okA; orderByMessage {
			return cmp.Compare(errA.Error(), errB.Error()) < 0
		}

		return cmpErrA.Span.Compare(cmpErrB.Span) < 0
	})
}

func ShowErrors(errs []error) bool {
	if errs := ErrorsToString(errs, MaxErrorOutput); len(errs) > 0 {
		os.Stderr.WriteString(errs)
		return true
	}
	return false
}

func ErrorsToString(errs []error, max int) string {
	SortErrors(errs)
	txt := strings.Builder{}
	for n, err := range errs {
		if n > 0 {
			txt.WriteString("\n")
		}
		if max > 0 && n == max {
			txt.WriteString(fmt.Sprintf("Too many errors, omitting %d errors...\n", len(errs)-n))
			break
		}
		txt.WriteString(fmt.Sprintf("[%d of %d] ", n+1, len(errs)))
		txt.WriteString(err.Error())
		txt.WriteString("\n")
	}
	return txt.String()
}
