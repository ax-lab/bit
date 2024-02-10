package common

import (
	"fmt"
	"os"
	"runtime"
)

func Handle[T any](value T, err error) T {
	doCheck(err, 1)
	return value
}

func Check(err error) {
	doCheck(err, 1)
}

func doCheck(err error, skip int) {
	if err != nil {
		fmt.Println()

		fmt.Fprintf(os.Stderr, "Error: %s", err)
		if _, file, line, ok := runtime.Caller(skip + 1); ok {
			fmt.Fprintf(os.Stderr, "\n\n       at %s:%d", file, line)
		}
		fmt.Fprintf(os.Stderr, "\n")

		fmt.Println()
		os.Exit(1)
	}
}
