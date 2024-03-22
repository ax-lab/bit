package bot

import (
	"io"
	"os"
)

type Runtime struct {
	out RuntimeWriter
	err RuntimeWriter
}

func (rt *Runtime) Out() RuntimeWriter {
	if rt.out == nil {
		rt.out = runtimeWriter{os.Stdout}
	}
	return rt.out
}

func (rt *Runtime) Err() RuntimeWriter {
	if rt.err == nil {
		rt.err = runtimeWriter{os.Stderr}
	}
	return rt.err
}

func (rt *Runtime) Eval(code Code) (Value, error) {
	return code.Eval(rt)
}

type RuntimeWriter interface {
	Write(str string) error
}

type runtimeWriter struct {
	inner io.Writer
}

func (rw runtimeWriter) Write(str string) error {
	_, err := rw.inner.Write([]byte(str))
	return err
}
