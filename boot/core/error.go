package core

import (
	"cmp"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
)

func Error(msg string, args ...any) ErrorWithLocation {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return ErrorWithLocation{
		text: msg,
	}
}

func ErrorAt(err error, at LocationPos) error {
	if err == nil {
		return nil
	}

	if pos, ok := err.(ErrorWithLocation); ok {
		pos.loc = at
		return pos
	}

	return Error(err.Error()).AtLocation(at)
}

type ErrorWithLocation struct {
	loc  LocationPos
	text string
}

func (err ErrorWithLocation) String() string {
	return err.Error()
}

func (err ErrorWithLocation) Error() string {
	if !err.loc.Valid() {
		return err.text
	}

	out := strings.Builder{}
	out.WriteString("in ")
	out.WriteString(err.loc.String())
	out.WriteString("\n\n")
	out.WriteString(Indent(err.text))

	return out.String()
}

func (err ErrorWithLocation) At(file string, pos ...int) ErrorWithLocation {
	err.loc = Location(file, pos...)
	return err
}

func (err ErrorWithLocation) AtLocation(loc LocationPos) ErrorWithLocation {
	err.loc = loc
	return err
}

func Check[T any](val T, err error) T {
	if err != nil {
		doFatal(err, 1)
	}
	return val
}

func CheckErrs[T any](val T, errs []error) T {
	errCount := 0
	for _, err := range errs {
		if err != nil {
			errCount++
			if errCount == 1 {
				fmt.Fprintf(os.Stderr, "\n")
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	if errCount > 0 {
		var fatalErr error
		if errCount > 1 {
			fatalErr = fmt.Errorf("exiting due to %d errors", errCount)
		} else {
			fatalErr = fmt.Errorf("exiting due to error")
		}
		doFatal(fatalErr, 1)
	}
	return val
}

func Handle(err error) {
	if err != nil {
		doFatal(err, 1)
	}
}

func Fatal(err error) {
	doFatal(err, 1)
}

func doFatal(err error, skip int) {
	fmt.Fprintf(os.Stderr, "\nFatal: %v\n", err)
	if _, file, line, ok := runtime.Caller(skip + 1); ok {
		fmt.Fprintf(os.Stderr, "       (at %s:%d)\n\n", file, line)
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}
	os.Exit(1)
}

func ErrorSort(errors []error) {
	sort.Slice(errors, func(a, b int) bool {
		return ErrorCompare(errors[a], errors[b]) < 0
	})
}

func ErrorCompare(a, b error) int {
	pa, hasA := a.(ErrorWithLocation)
	pb, hasB := b.(ErrorWithLocation)
	if hasA != hasB {
		if !hasA {
			return -1 // without location first
		} else {
			return +1
		}
	} else if hasA {
		return pa.loc.Compare(pb.loc)
	}

	sa, sb := "", ""
	if a != nil {
		sa = a.Error()
	}
	if b != nil {
		sb = b.Error()
	}
	return cmp.Compare(sa, sb)
}
