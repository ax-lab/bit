package code_tests

import (
	"testing"

	"axlab.dev/bit/code"
)

func TestPrint(t *testing.T) {
	test := NewTest(t)
	program := &test.Program

	varAns := code.Var{
		Name: code.Id("ans"),
		Type: program.Types().Scalar(code.TypeScalarNumber),
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

	program.Append(block)
	test.ExpectStdOut = "The answer to life, the universe, and everything is 42\n"
	test.ExpectResult = []any{"The answer to life, the universe, and everything is", int64(42)}
	test.Check()
}
