package code

import (
	"fmt"
	"io"
	"os"
)

type Value interface {
	String() string
	Bool() bool
}

type Runtime struct {
	Stack  []Value
	StdOut io.Writer
	StdErr io.Writer
}

func NewRuntime() *Runtime {
	return &Runtime{
		StdOut: os.Stdout,
		StdErr: os.Stderr,
	}
}

func (rt *Runtime) SlotIndex(slot int) int {
	return len(rt.Stack) - slot - 1
}

func (rt *Runtime) Out(txt string, args ...any) {
	if len(args) > 0 {
		txt = fmt.Sprintf(txt, args...)
	}
	io.WriteString(rt.StdOut, txt)
}

func (rt *Runtime) Err(txt string, args ...any) {
	if len(args) > 0 {
		txt = fmt.Sprintf(txt, args...)
	}
	io.WriteString(rt.StdErr, txt)
}
