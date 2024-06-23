package code_test

import (
	"strings"
	"testing"

	"axlab.dev/bit/code"
	"github.com/stretchr/testify/require"
)

func TestPrint(t *testing.T) {
	test := require.New(t)

	varAns := code.Var{
		Name: code.Id("ans"),
		Type: code.TypeNumber(),
	}

	block := code.ExprNew(code.Block{
		List: []code.Expr{
			code.ExprNew(code.Let{
				Decl: varAns,
				Init: code.ExprNew(code.Number{Value: 42}),
			}),
			code.ExprNew(code.Print{
				Args: []code.Expr{
					code.ExprNew(code.Str{Value: "The answer to life, the universe, and everything is"}),
					code.ExprNew(varAns),
				},
			}),
		},
	})

	stdOut := strings.Builder{}
	stdErr := strings.Builder{}
	rt := code.Runtime{
		StdOut: &stdOut,
		StdErr: &stdErr,
	}

	eval := code.MustCompile(block)
	ans, err := eval(&rt)

	test.Equal("The answer to life, the universe and everything is 42\n", stdOut)
	test.Empty(stdErr)
	test.EqualValues([]any{"The answer to life, the universe, and everything is", 42}, ans)
	test.NoError(err)
}
