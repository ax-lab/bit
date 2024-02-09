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

	program.DeclareGlobal(Word("print"), ParsePrint{})

	program.DeclareGlobal(TokenString, Replace{ParseString{}})

	program.OutputAll(Module{})
	program.OutputAll(Line{})
	program.OutputAll(Print{})
	program.OutputAll(Str(""))
}

func (program *Program) OutputAll(key Key) {
	program.DeclareGlobal(key, Output{})
}

type Group struct{}

func (val Group) Bind(node *Node) {
	node.Bind(Group{})
}

func (val Group) Repr() string {
	return "Group"
}

func (val Group) IsEqual(other Key) bool {
	if v, ok := other.(Group); ok {
		return val == v
	}
	return false
}
