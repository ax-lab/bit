package bot

import "fmt"

type Code interface {
	Type() Type
	Eval(rt *Runtime) (Value, error)
	Output(cw *CodeWriter) CodeValue
}

type CodePrint struct {
	Arg Code
}

func (expr CodePrint) Type() Type {
	return expr.Arg.Type()
}

func (expr CodePrint) Eval(rt *Runtime) (Value, error) {
	arg, err := expr.Arg.Eval(rt)
	if err != nil {
		return arg, err
	}

	typ := arg.Type()
	err = typ.Print(arg, rt.Out())
	return arg, err
}

func (expr CodePrint) Output(cw *CodeWriter) CodeValue {
	arg := expr.Arg.Output(cw)
	cw.Import("fmt")
	cw.Push("fmt.Print(%s)", arg)
	return arg
}

type CodeBool bool

func (expr CodeBool) Type() Type {
	return TypeBool()
}

func (expr CodeBool) Eval(rt *Runtime) (Value, error) {
	return Bool(expr), nil
}

func (expr CodeBool) Output(cw *CodeWriter) CodeValue {
	var out string
	if expr {
		out = "true"
	} else {
		out = "false"
	}
	return cw.PushExpr(out)
}

type CodeInt int

func (expr CodeInt) Type() Type {
	return TypeInt()
}

func (expr CodeInt) Eval(rt *Runtime) (Value, error) {
	return Int(expr), nil
}

func (expr CodeInt) Output(cw *CodeWriter) CodeValue {
	out := fmt.Sprintf("%v", expr)
	return cw.PushExpr(out)
}

type CodeStr string

func (expr CodeStr) Type() Type {
	return TypeStr()
}

func (expr CodeStr) Eval(rt *Runtime) (Value, error) {
	return Str(expr), nil
}

func (expr CodeStr) Output(cw *CodeWriter) CodeValue {
	out := fmt.Sprintf("%#v", expr)
	return cw.PushExpr(out)
}
