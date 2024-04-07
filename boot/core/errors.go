package core

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func Errors(errs []error, msg string, args ...any) error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if len(errs) > 0 {
		out := strings.Builder{}
		out.WriteString(msg)
		out.WriteString(":")
		for _, err := range errs {
			out.WriteString("\n- ")
			out.WriteString(err.Error())
		}
		return errors.New(out.String())
	}

	return errors.New(msg)
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
