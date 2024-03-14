package boot

import (
	"fmt"
	"os"
	"runtime"
)

func FatalAt(file string, line int, err error) {
	Fatal(fmt.Errorf("at %s:%d: %v", file, line, err))
}

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "\nfatal: %v\n\n", err)
	os.Exit(1)
}

func Caller(skip int, label string) string {
	if _, file, line, ok := runtime.Caller(skip + 1); ok {
		return fmt.Sprintf("%s%s:%d", label, file, line)
	}
	return ""
}
