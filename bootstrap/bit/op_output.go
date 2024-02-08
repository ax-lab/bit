package bit

type Output struct{}

func (op Output) IsSame(other Binding) bool {
	if v, ok := other.(Output); ok {
		return v == op
	}
	return false
}

func (op Output) Precedence() Precedence {
	return PrecOutput
}

func (op Output) Process(args *BindArgs) {
	for _, it := range args.Nodes {
		if _, ok := it.Value().(Expr); !ok {
			it.AddError("cannot output value: %s", it.Value().Repr())
		}
	}
}

func (op Output) String() string {
	return "Output"
}
