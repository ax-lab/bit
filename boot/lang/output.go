package lang

import (
	"fmt"

	"axlab.dev/bit/code"
	"axlab.dev/bit/core"
)

func OutputCode(mod *core.Module, nodes core.NodeList) {
	compiler := mod.Compiler()
	seq := outputSequence(mod, nodes)
	compiler.OutputExpr(mod, seq)
}

func outputSequence(mod *core.Module, list core.NodeList) (out code.Seq) {
	out = code.SeqNew(list.Span())
	for _, it := range list.Nodes() {
		var (
			expr core.Expr
			err  error
		)
		switch val := it.Value().(type) {
		case Line:
			expr, err = outputLine(val)
		default:
			err = fmt.Errorf("cannot output node: %s", val.String())
		}

		if err != nil {
			err = core.ErrorAt(it.Span(), err)
			if stop := mod.Error(err); stop {
				break
			}
		}

		if expr != nil {
			out.Push(expr)
		}
	}

	return out
}

func outputLine(line Line) (out core.Expr, err error) {
	list := (core.NodeList)(line)
	if list.Len() == 0 {
		return
	} else if list.Len() != 1 {
		return nil, fmt.Errorf("invalid line expression -- lines should reduce to a single node when evaluated")
	}

	out, err = outputExpr(list.Get(0))
	return
}

func outputExpr(node core.Node) (out core.Expr, err error) {
	span := node.Span()
	switch val := node.Value().(type) {
	case PrintExpr:
		out, err = outputPrintExpr(span, val)
	case core.Literal:
		raw := val.Prefix == "r"
		text := ParseStringLiteral(val.RawText, raw, val.Delim)
		out = code.StrNew(span, text)
	default:
		err = fmt.Errorf("cannot output expression for node: %s", val.String())
	}

	if err != nil {
		err = core.ErrorAt(node.Span(), err)
	}
	return
}

func outputPrintExpr(span core.Span, expr PrintExpr) (out core.Expr, err error) {
	args := make([]core.Expr, 0, expr.Args.Len())
	for _, node := range expr.Args.Nodes() {
		expr, exprErr := outputExpr(node)
		if exprErr != nil {
			return nil, exprErr
		}
		if expr != nil {
			args = append(args, expr)
		}
	}

	out = code.PrintNew(span, args...)
	return
}
