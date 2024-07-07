package code_tests

import (
	"strings"
	"testing"

	"axlab.dev/bit/code"
	"github.com/stretchr/testify/require"
)

type Test struct {
	*require.Assertions

	Program code.Program

	ExpectStdOut string
	ExpectStdErr string
	ExpectResult any
}

func NewTest(t *testing.T) *Test {
	test := &Test{Assertions: require.New(t)}
	return test
}

func (test *Test) Check() {
	if test.Program.HasErrors() {
		test.Fail("program with errors: %s", test.Program.Errors.String())
	}

	eval, err := test.Program.Compile()
	test.NoError(err, "program compilation error")

	stdOut := strings.Builder{}
	stdErr := strings.Builder{}
	rt := code.Runtime{
		StdOut: &stdOut,
		StdErr: &stdErr,
	}

	ans, err := eval(&rt)
	test.NoError(err, "program evaluation error")

	test.Equal(test.ExpectStdOut, stdOut.String())

	if test.ExpectStdErr != "" {
		test.Equal(test.ExpectStdErr, stdErr.String())
	} else {
		test.Empty(stdErr.String(), "expected no error output")
	}

	if test.ExpectResult != nil {
		test.EqualValues(test.ExpectResult, ans)
	}
}
