package code

import (
	"fmt"
	"io"
	"strings"

	"axlab.dev/bit/core"
)

type Print struct {
	Args []core.Expr
}

func (expr Print) String() string {
	out := strings.Builder{}
	out.WriteString("Print(")
	for n, it := range expr.Args {
		if n > 0 {
			out.WriteString(", ")
		}
		out.WriteString(it.String())
	}
	out.WriteString(")")
	return out.String()
}

func (expr Print) Eval(rt *core.Runtime) (core.Value, error) {
	args := make([]core.Value, 0, len(expr.Args))
	for _, it := range expr.Args {
		if val, err := it.Eval(rt); err == nil {
			args = append(args, val)
		} else {
			return nil, err
		}
	}

	out := rt.StdOut()

	write := func(out io.Writer, str string) error {
		_, err := out.Write([]byte(str))
		if err != nil {
			err = fmt.Errorf("print error: %v", err)
		}
		return err
	}

	hasOutput := false
	for _, val := range args {
		if val == nil {
			continue
		}

		txt := val.String()
		if len(txt) == 0 {
			continue
		}

		if hasOutput {
			if err := write(out, " "); err != nil {
				return nil, err
			}
		}

		hasOutput = true
		if err := write(out, txt); err != nil {
			return nil, err
		}
	}

	err := write(out, "\n")
	return nil, err
}
