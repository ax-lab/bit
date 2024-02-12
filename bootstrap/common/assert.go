package common

import (
	"fmt"
	"os"
	"runtime"
)

func Try[T any](value T, err error) T {
	doCheck(err, 1)
	return value
}

func Check(err error, msg ...any) {
	doCheck(err, 1)
}

func doCheck(err error, skip int, msg ...any) {
	if err != nil {
		fmt.Println()

		txt := Fmt(msg...)
		if txt != "" {
			txt = " -- " + txt
		}

		fmt.Fprintf(os.Stderr, "Error: %s%s", err, txt)
		if _, file, line, ok := runtime.Caller(skip + 1); ok {
			fmt.Fprintf(os.Stderr, "\n\n       at %s:%d", file, line)
		}
		fmt.Fprintf(os.Stderr, "\n")

		fmt.Println()
		os.Exit(1)
	}
}

func Fmt(args ...any) string {
	switch len(args) {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(args[0])
	default:
		if msg, ok := args[0].(string); ok {
			return fmt.Sprintf(msg, args[1:]...)
		} else {
			return fmt.Sprint(args...)
		}
	}
}

func ExeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func NoError(err error, msg string) {
	if err != nil {
		if msg != "" {
			fmt.Fprintf(os.Stderr, "\nfatal error: %s - %v\n\n", msg, err)
		} else {
			fmt.Fprintf(os.Stderr, "\nfatal error: %v\n\n", err)
		}
		os.Exit(3)
	}
}

func Assert(cond bool, msg string, args ...any) {
	if !cond {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		if msg == "" {
			msg = "assertion failed"
		}
		panic(msg)
	}
}

type msgWithArgs struct {
	msg  string
	args []any
}

func (m msgWithArgs) String() string {
	if len(m.args) == 0 {
		return m.msg
	} else {
		return fmt.Sprintf(m.msg, m.args...)
	}
}

func Msg(msg string, args ...any) fmt.Stringer {
	return msgWithArgs{msg, args}
}
