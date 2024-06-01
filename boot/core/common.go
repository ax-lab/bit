package core

import (
	"fmt"
)

const (
	BuildCacheDir = ".build"
)

func Version() string {
	return "0.1.0"
}

func Not[F ~func(T) bool, T any](pred F) func(T) bool {
	return func(t T) bool {
		return !pred(t)
	}
}

func Msg(args ...any) string {
	if len(args) == 0 {
		return ""
	}

	msg, ok := args[0].(string)
	if ok {
		args = args[1:]
	} else {
		msg = fmt.Sprint(args...)
		args = nil
	}

	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	return msg
}

func NoError(err error, args ...any) {
	if err != nil {
		msg := Msg(args...)
		if len(msg) > 0 {
			msg = fmt.Sprintf("%s: %v", msg, err)
		} else {
			msg = err.Error()
		}
		panic(msg)
	}
}

func Try[T any](res T, err error) T {
	NoError(err)
	return res
}
