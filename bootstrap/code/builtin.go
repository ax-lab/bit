package code

import "fmt"

func NewBool(v bool) *Bool {
	return &Bool{Val: v}
}

func NewInt(v int) *Int {
	return &Int{Val: v}
}

func NewStr(v string) *Str {
	return &Str{Val: v}
}

type Bool struct {
	Id
	Val bool
}

func (v *Bool) Type() Type {
	return BoolType()
}

func (v *Bool) Eval(rt *Runtime) (Value, error) {
	return v, nil
}

func (v *Bool) Bool() bool {
	return v.Val
}

func (v *Bool) String() string {
	if v.Val {
		return "true"
	} else {
		return "false"
	}
}

func (v *Bool) OutputCpp(ctx *CppContext) {
	ctx.IncludeSystem("stdbool.h")
	ctx.Body.Write(v.String())
}

func (v *Bool) Repr(mode Repr) string {
	return v.String()
}

type Int struct {
	Id
	Val int
}

func (v *Int) Type() Type {
	return IntType()
}

func (v *Int) Eval(rt *Runtime) (Value, error) {
	return v, nil
}

func (v *Int) Bool() bool {
	return v.Val != 0
}

func (v *Int) String() string {
	return fmt.Sprint(v.Val)
}

func (v *Int) OutputCpp(ctx *CppContext) {
	ctx.Body.WriteFmt("%d", v)
}

func (v *Int) Repr(mode Repr) string {
	return fmt.Sprintf("int(%d)", v)
}

type Str struct {
	Id
	Val string
}

func (v *Str) Type() Type {
	return StrType()
}

func (v *Str) Eval(rt *Runtime) (Value, error) {
	return v, nil
}

func (v *Str) Bool() bool {
	return v.Val != ""
}

func (v *Str) String() string {
	return v.Val
}

func (v *Str) OutputCpp(ctx *CppContext) {
	WriteLiteralString(ctx.Body, v.Val)
}

func (v *Str) Repr(mode Repr) string {
	if mode == ReprLabel {
		return "string"
	}
	return fmt.Sprintf("%#v", v.Val)
}
