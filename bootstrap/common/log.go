package common

import (
	"fmt"
	"os"
	"runtime"
)

const (
	ERR  = "error"
	WARN = "warning"
)

func Caller(skip int, label string) string {
	if _, file, line, ok := runtime.Caller(skip + 1); ok {
		return fmt.Sprintf("%s%s:%d", label, file, line)
	}
	return ""
}

func CallerSuffix(skip int) string {
	return Caller(skip+1, "  -- ")
}

func Out(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if len(msg) > 0 {
		os.Stdout.WriteString(ExpandTabs(msg, -1))
	}
}

func Err(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if len(msg) > 0 {
		os.Stderr.WriteString(ExpandTabs(msg, -1))
	}
}

func Fatal(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	Err("\n[FATAL] %s\n", msg)
	os.Exit(1)
}

func Warn(err error, msg string, args ...any) bool {
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		fmt.Printf("\n[WRN] %s: %v%s\n", msg, err, CallerSuffix(1))
		return false
	}
	return true
}

func Error(err error, msg string, args ...any) bool {
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		fmt.Printf("\n[ERR] %s: %v%s\n", msg, err, CallerSuffix(1))
		return false
	}
	return true
}

func Info(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if len(msg) > 0 {
		fmt.Printf("\n[INF] %s%s\n", msg, CallerSuffix(1))
	}
}
