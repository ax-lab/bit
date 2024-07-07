package code

import (
	"fmt"
)

type EvalFunc func(rt *Runtime) (out any, err error)

func (program *Program) Compile() (eval EvalFunc, err error) {
	program.codeSync.Lock()
	defer program.codeSync.Unlock()
	eval, err = compileList(&program.scope, program.codeList)
	return
}

func compileList(scope *Scope, list []Expr) (eval EvalFunc, err error) {
	var code []EvalFunc
	for _, it := range list {
		eval, err := compileExpr(scope, it)
		if err != nil {
			return nil, err
		}
		code = append(code, eval)
	}

	eval = func(rt *Runtime) (out any, err error) {
		for _, it := range code {
			if out, err = it(rt); err != nil {
				return nil, err
			}
		}
		return out, nil
	}

	return eval, nil
}

func compileExpr(scope *Scope, expr Expr) (eval EvalFunc, err error) {
	switch val := expr.Value().(type) {

	case Block:
		blockScope := scope.NewChild()
		if evalInner, err := compileList(blockScope, val.List); err != nil {
			return nil, err
		} else {
			eval = func(rt *Runtime) (out any, err error) {
				cleanup := rt.InitScope(blockScope)
				defer cleanup()
				return evalInner(rt)
			}
		}

	case Let:
		init, err := compileExpr(scope, val.Init)
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
			if fn, err := compileExpr(scope, arg); err == nil {
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
