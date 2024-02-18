package code

import "fmt"

type Variable struct {
	Id

	Source any

	typ    Type
	scope  *Scope
	name   string
	offset int

	slot      *varSlot
	counter   int
	processed string
}

func (v *Variable) SetType(typ Type) {
	v.typ = typ
}

func (v *Variable) Type() Type {
	return v.typ
}

func (v *Variable) Eval(rt *Runtime) (Value, error) {
	index := rt.SlotIndex(v.slot.index)
	return rt.Stack[index], nil
}

func (v *Variable) OutputCpp(ctx *CppContext) {
	ctx.Body.Write(v.OutputName())
}

func (v *Variable) Repr(mode Repr) string {
	return v.Name()
}

func (v *Variable) Offset() int {
	return v.offset
}

func (v *Variable) Name() string {
	return v.name
}

func (v *Variable) OutputName() string {
	if v.processed == "" {
		panic("variable output name not processed")
	}
	return v.processed
}

func (v *Variable) CheckBound() {
	if v.slot == nil {
		panic(fmt.Sprintf("variable `%s` was not bound", v.name))
	}
}

func (v *Variable) SetVar(expr Expr) *SetVar {
	return &SetVar{variable: v, Expr: expr}
}

type SetVar struct {
	Id
	variable *Variable
	Expr     Expr
}

func (v *SetVar) Exec(rt *Runtime) (err error) {
	index := rt.SlotIndex(v.variable.slot.index)
	rt.Stack[index], err = v.Expr.Eval(rt)
	return err
}

func (v *SetVar) OutputCpp(ctx *CppContext) {
	ctx.Body.WriteFmt("%s = ", v.variable.OutputName())
	v.Expr.OutputCpp(ctx)
	ctx.Body.Write(";")
	ctx.Body.EnsureBlank()
}

func (v *SetVar) Repr(mode Repr) string {
	return fmt.Sprintf("%s = %s", v.variable.Name(), v.Expr.Repr(mode))
}
