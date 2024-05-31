package tests

import (
	"strings"

	"axlab.dev/bit/core"
	"axlab.dev/bit/lang"
	"github.com/stretchr/testify/require"
)

type TestRun struct {
	test *require.Assertions

	Success bool
	Result  any
	Error   error
	StdOut  string
	StdErr  string
}

func (run TestRun) NoError() {
	run.test.Empty(run.StdErr)
	run.test.NoError(run.Error)
	run.test.True(run.Success)
}

func Run(test *require.Assertions, input string) (out TestRun) {
	out.test = test

	compiler := core.Compiler{}
	compiler.Sources.Preload("main.bit", core.Text(input))

	err := lang.Declare(&compiler)
	test.NoError(err)

	compiler.AddSource("main.bit")

	stdOut := strings.Builder{}
	stdErr := strings.Builder{}
	compiler.RedirectStdOut(&stdOut)
	compiler.RedirectStdErr(&stdErr)
	out.Success = compiler.Execute()

	rt := compiler.CreateRuntime()
	if out.Success {
		out.Result, out.Error = rt.Run()
		out.Success = out.Error == nil
	}

	out.StdOut = stdOut.String()
	out.StdErr = stdErr.String()
	return
}
