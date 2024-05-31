package core

type Expr interface {
	Eval(rt *Runtime) (Value, error)
	String() string
}
