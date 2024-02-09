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
	// only flag nodes that can be output as-is as done
}

func (op Output) String() string {
	return "Output"
}
