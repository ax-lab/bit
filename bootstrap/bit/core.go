package bit

func (program *Program) InitCore() {
	program.DeclareGlobal(TokenBreak, SplitLines{})
	program.DeclareGlobal(Line{}, ParseIndent{})

	program.DeclareGlobal(Symbol("("), ParseBrackets{"(", ")"})
	program.DeclareGlobal(Symbol(")"), ParseBrackets{"(", ")"})

	program.DeclareGlobal(Symbol("["), ParseBrackets{"[", "]"})
	program.DeclareGlobal(Symbol("]"), ParseBrackets{"[", "]"})

	program.DeclareGlobal(Symbol("{"), ParseBrackets{"{", "}"})
	program.DeclareGlobal(Symbol("}"), ParseBrackets{"{", "}"})
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
