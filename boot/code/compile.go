package code

import (
	"fmt"
	"log"
)

type EvalFunc func(rt *Runtime) (out any, err error)

func MustCompile(expr Expr) EvalFunc {
	global := &Scope{}
	out, err := Compile(global, expr)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func Compile(scope *Scope, expr Expr) (eval EvalFunc, err error) {
	switch val := expr.Value().(type) {

	case Block:
		var (
			code []EvalFunc
		)

		blockScope := scope.NewChild()
		for _, it := range val.List {
			eval, err := Compile(blockScope, it)
			if err != nil {
				return nil, err
			}
			code = append(code, eval)
		}

		eval = func(rt *Runtime) (out any, err error) {
			cleanup := rt.InitScope(blockScope)
			defer cleanup()

			for _, it := range code {
				if out, err = it(rt); err != nil {
					return nil, err
				}
			}
			return out, nil
		}

	case Let:
		init, err := Compile(scope, val.Init)
		if err != nil {
			return nil, err
		}

		id, err := scope.Declare(val.Decl)
		if err != nil {
			return nil, err
		}

		eval = func(rt *Runtime) (out any, err error) {
			if val, err := init(rt); err == nil {
				rt.SetVar(id, val)
			}
			return val, err
		}

	case Var:

		id, err := scope.Resolve(val)
		if err != nil {
			return nil, err
		}

		eval = func(rt *Runtime) (out any, err error) {
			out = rt.GetVar(id)
			return
		}

	case Number:

		eval = func(rt *Runtime) (out any, err error) {
			out = val.Value
			return out, nil
		}

	case Str:

		eval = func(rt *Runtime) (out any, err error) {
			out = val.Value
			return out, nil
		}

	case Print:

		args := make([]EvalFunc, 0, len(val.Args))
		for _, arg := range val.Args {
			if fn, err := Compile(scope, arg); err == nil {
				args = append(args, fn)
			} else {
				return nil, err
			}
		}

		eval = func(rt *Runtime) (out any, err error) {
			if len(args) > 0 {
				vals := make([]any, len(args))
				for argIdx, argEval := range args {
					if val, err := argEval(rt); err == nil {
						vals[argIdx] = val
					} else {
						return nil, err
					}
				}
				out = vals

				sep := ""
				for _, it := range vals {
					if _, err := fmt.Fprintf(rt.StdOut, "%s%v", sep, it); err != nil {
						return nil, fmt.Errorf("print: %w", err)
					}
					sep = " "
				}
			}

			if _, err := fmt.Fprintln(rt.StdOut); err != nil {
				return nil, fmt.Errorf("print: %w", err)
			}

			return out, nil
		}

	default:
		return nil, fmt.Errorf("cannot compile expression: %s", expr)
	}

	return eval, err
}
