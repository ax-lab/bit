package core

type Expr interface {
	Span() Span
	Eval(rt *Runtime) (Value, error)
	String() string
}
