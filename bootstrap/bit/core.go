package bit

func (program *Program) InitCore() {
	program.DeclareGlobal(TokenBreak, SplitLines{})
	program.DeclareGlobal(Indented{}, ParseIndent{})

	program.DeclareGlobal(Symbol("("), ParseBrackets{"(", ")"})
	program.DeclareGlobal(Symbol(")"), ParseBrackets{"(", ")"})

	program.DeclareGlobal(Symbol("["), ParseBrackets{"[", "]"})
	program.DeclareGlobal(Symbol("]"), ParseBrackets{"[", "]"})

	program.DeclareGlobal(Symbol("{"), ParseBrackets{"{", "}"})
	program.DeclareGlobal(Symbol("}"), ParseBrackets{"{", "}"})

	program.DeclareGlobal(Module{}, Output{})
	program.DeclareGlobal(Line{}, Output{})
	program.DeclareGlobal(TokenString, Output{})
}

type Group struct{}

func (val Group) Bind(node *Node) {
	node.Bind(Group{})
}

func (val Group) String() string {
	return "Group"
}

func (val Group) IsEqual(other Key) bool {
	if v, ok := other.(Group); ok {
		return val == v
	}
	return false
}

type CanOutput interface {
	HasOutput() bool
}

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
		if _, ok := it.Value().(CanOutput); !ok {
			it.AddError("cannot output value: %s", it.Value().String())
		}
	}
}

func (op Output) String() string {
	return "Output"
}
