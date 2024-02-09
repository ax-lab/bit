package bit

type Runtime struct{}

type Code interface {
	Output(node *Node) Expr
}

type Expr interface {
	Eval(rt *Runtime) Result
}

type Result interface {
	String() string
}
