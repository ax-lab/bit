package tests

import (
	"fmt"
	"path"
	"strings"

	"axlab.dev/bit/core"
	"axlab.dev/bit/lang"
	"github.com/stretchr/testify/require"
)

type TestRun struct {
	test     *require.Assertions
	compiler *core.Compiler

	writerStdOut *strings.Builder
	writerStdErr *strings.Builder

	Result any
	Error  error
}

func (run TestRun) Success() bool {
	return !run.compiler.HasErrors() && run.Error == nil && run.StdErr() == ""
}

func (run TestRun) StdOut() string {
	return run.writerStdOut.String()
}

func (run TestRun) StdErr() string {
	return run.writerStdErr.String()
}

func (run TestRun) NoError() {
	if run.compiler.HasErrors() {
		fmt.Println(run.StdErr())
		run.test.Fail("Compilation failed with errors")
	} else {
		run.test.Empty(run.StdErr())
	}

	run.test.NoError(run.Error)
	run.test.True(run.Success())
}

func Run(name string, test *require.Assertions, input string) (out TestRun) {
	out = compileTest(name, test, input)
	if !out.Success() {
		return
	}

	rt := out.compiler.CreateRuntime()
	out.Result, out.Error = rt.Run()

	return
}

func RunC(name string, test *require.Assertions, input string) (out TestRun) {
	out = compileTest(name, test, input)
	if !out.Success() {
		return
	}

	output := out.compiler.Output.Get("cpp")
	cmd := lang.OutputC(out.compiler, output)
	if !out.Success() {
		out.compiler.OutputErrors()
		return
	}

	test.NotNil(cmd)

	out.Error = cmd.Run()
	out.Result = cmd.ExitCode()
	out.writerStdOut.WriteString(cmd.StdOut())
	out.writerStdErr.WriteString(cmd.StdErr())

	return
}

func compileTest(name string, test *require.Assertions, input string) (out TestRun) {
	out.test = test
	out.writerStdOut = &strings.Builder{}
	out.writerStdErr = &strings.Builder{}

	testDir := path.Join(core.ProjectRoot(), core.BuildCacheDir, "tests", name)

	compiler := &core.Compiler{}
	compiler.Output.BaseDir = testDir

	out.compiler = compiler

	main := name + ".bit"
	compiler.Sources.Preload(main, core.Text(input))

	err := lang.Declare(compiler)
	test.NoError(err)

	compiler.AddSource(main)

	compiler.RedirectStdOut(out.writerStdOut)
	compiler.RedirectStdErr(out.writerStdErr)

	ok := compiler.Execute()
	test.Equal(ok, !compiler.HasErrors())

	return out
}
