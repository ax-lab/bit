package logs

import (
	"fmt"
	"os"
	"runtime"
	"strings"
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

var (
	curSep = "\n"
)

func Break() {
	if len(curSep) > 1 {
		os.Stdout.WriteString(curSep[:1])
		curSep = curSep[1:]
	}
}

func Sep() {
	os.Stdout.WriteString(curSep)
	curSep = ""
}

func Out(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if len(msg) > 0 {
		os.Stdout.WriteString(msg)
		if strings.HasSuffix(msg, "\n\n") {
			curSep = ""
		} else if strings.HasSuffix(msg, "\n") {
			curSep = "\n"
		} else {
			curSep = "\n\n"
		}
	}
}

func Err(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if len(msg) > 0 {
		os.Stderr.WriteString(msg)
		if strings.HasSuffix(msg, "\n\n") {
			curSep = ""
		} else if strings.HasSuffix(msg, "\n") {
			curSep = "\n"
		} else {
			curSep = "\n\n"
		}
	}
}

func Warn(err error, msg string, args ...any) bool {
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		Sep()
		fmt.Printf("[WRN] %s: %v%s\n", msg, err, CallerSuffix(1))
		return false
	}
	return true
}

func Error(err error, msg string, args ...any) bool {
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		Sep()
		fmt.Printf("[ERR] %s: %v%s\n", msg, err, CallerSuffix(1))
		return false
	}
	return true
}

func Info(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	Sep()
	fmt.Printf("[INF] %s%s\n", msg, CallerSuffix(1))
}
