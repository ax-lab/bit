package bit

func (program *Program) InitCore() {
	if !program.coreInit.CompareAndSwap(false, true) {
		return
	}

	program.DeclareGlobal(TokenBreak, SplitLines{})
	program.DeclareGlobal(Indented{}, ParseIndent{})

	program.DeclareGlobal(Symbol("("), ParseBrackets{"(", ")"})
	program.DeclareGlobal(Symbol(")"), ParseBrackets{"(", ")"})

	program.DeclareGlobal(Symbol("["), ParseBrackets{"[", "]"})
	program.DeclareGlobal(Symbol("]"), ParseBrackets{"[", "]"})

	program.DeclareGlobal(Symbol("{"), ParseBrackets{"{", "}"})
	program.DeclareGlobal(Symbol("}"), ParseBrackets{"{", "}"})

	program.DeclareGlobal(Word("print"), ParsePrint{})
	program.DeclareGlobal(Word("let"), ParseLet{})

	program.DeclareGlobal(TokenString, Replace{ParseString{}})
	program.DeclareGlobal(TokenInteger, Replace{ParseInt{}})

	program.OutputAll(Str(""))
	program.OutputAll(Int(0))

	program.OutputAll(Module{})
	program.OutputAll(Line{})
	program.OutputAll(Print{})
	program.OutputAll(Var{})
	program.OutputAll(Let{})
}

func (program *Program) OutputAll(key Key) {
	program.DeclareGlobal(key, Output{})
}

type Group struct{}

func (val Group) Bind(node *Node) {
	node.Bind(Group{})
}

func (val Group) Repr(oneline bool) string {
	return "Group"
}

func (val Group) IsEqual(other Key) bool {
	if v, ok := other.(Group); ok {
		return val == v
	}
	return false
}
