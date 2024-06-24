package code

import (
	"io"
	"sync/atomic"
)

type Runtime struct {
	StdErr io.Writer
	StdOut io.Writer

	frameId atomic.Uint64
	frames  []stackFrame
}

func (rt *Runtime) GetVar(id VarId) any {
	last := len(rt.frames) - 1
	return rt.frames[last-int(id.frame)].vars[id.index]
}

func (rt *Runtime) SetVar(id VarId, val any) {
	last := len(rt.frames) - 1
	rt.frames[last-int(id.frame)].vars[id.index] = val
}

func (rt *Runtime) InitScope(scope *Scope) (cleanFn func()) {
	runId := rt.frameId.Add(1)
	frame := stackFrame{
		runId: runId,
		vars:  make([]any, scope.varCount),
	}

	rt.frames = append(rt.frames, frame)
	return func() {
		last := &rt.frames[len(rt.frames)-1]
		if last.runId != runId {
			panic("cleaning up invalid frame in the stack")
		}
		rt.frames = rt.frames[:len(rt.frames)-1]
	}
}

type stackFrame struct {
	runId uint64
	vars  []any
}
