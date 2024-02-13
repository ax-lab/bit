package bit

func (program *Program) InitCore() {
	if !program.coreInit.CompareAndSwap(false, true) {
		return
	}

	program.DeclareGlobal(TokenBreak, SplitLines{})
	program.DeclareGlobal(Indented{}, ParseIndent{})
	program.DeclareGlobal(IndentedGroup{}, ParseBlocks{})

	program.DeclareGlobal(Symbol("("), ParseBrackets{"(", ")"})
	program.DeclareGlobal(Symbol(")"), ParseBrackets{"(", ")"})

	program.DeclareGlobal(Symbol("["), ParseBrackets{"[", "]"})
	program.DeclareGlobal(Symbol("]"), ParseBrackets{"[", "]"})

	program.DeclareGlobal(Symbol("{"), ParseBrackets{"{", "}"})
	program.DeclareGlobal(Symbol("}"), ParseBrackets{"{", "}"})

	program.DeclareGlobal(Word("print"), ParsePrint{})
	program.DeclareGlobal(Word("let"), ParseLet{})
	program.DeclareGlobal(Word("if"), ParseIf{})

	program.DeclareGlobal(TokenString, Replace{ParseString{}})
	program.DeclareGlobal(TokenInteger, Replace{ParseInt{}})

	program.OutputAll(Str(""))
	program.OutputAll(Int(0))

	program.DeclareGlobal(Line{}, Simplify{})
	program.DeclareGlobal(Group{}, Simplify{})

	program.OutputAll(Module{})
	program.OutputAll(Block{})
	program.OutputAll(Print{})
	program.OutputAll(Var{})
	program.OutputAll(Let{})
	program.OutputAll(If{})
}

func (program *Program) OutputAll(key Key) {
	program.DeclareGlobal(key, Output{})
}
