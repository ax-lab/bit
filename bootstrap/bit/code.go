package bit

type Runtime struct{}

type Expr interface {
	Eval(rt *Runtime) (any, error)
}
