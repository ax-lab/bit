package bit_lang

import (
	"axlab.dev/bit/bit"
)

func InitCompiler(program *bit.Program) {
	program.DeclareGlobal(bit.TokenBreak, SplitLines{})
	program.DeclareGlobal(Indented{}, ParseIndent{})
	program.DeclareGlobal(IndentedGroup{}, ParseBlocks{})

	program.DeclareGlobal(bit.Symbol("("), ParseBrackets{"(", ")"})
	program.DeclareGlobal(bit.Symbol(")"), ParseBrackets{"(", ")"})

	program.DeclareGlobal(bit.Symbol("["), ParseBrackets{"[", "]"})
	program.DeclareGlobal(bit.Symbol("]"), ParseBrackets{"[", "]"})

	program.DeclareGlobal(bit.Symbol("{"), ParseBrackets{"{", "}"})
	program.DeclareGlobal(bit.Symbol("}"), ParseBrackets{"{", "}"})

	program.DeclareGlobal(bit.Word("print"), ParsePrint{})
	program.DeclareGlobal(bit.Word("let"), ParseLet{})
	program.DeclareGlobal(bit.Word("if"), ParseIf{})

	program.DeclareGlobal(bit.TokenString, Replace{ParseString{}})
	program.DeclareGlobal(bit.TokenInteger, Replace{ParseInt{}})
	program.DeclareGlobal(bit.Word("true"), Replace{ParseBool{}})
	program.DeclareGlobal(bit.Word("false"), Replace{ParseBool{}})

	outputAll(program, Str(""))
	outputAll(program, Int(0))
	outputAll(program, Bool(false))

	program.DeclareGlobal(Line{}, Simplify{})
	program.DeclareGlobal(Group{}, Simplify{})

	outputAll(program, bit.Module{})
	outputAll(program, Block{})
	outputAll(program, Print{})
	outputAll(program, Var{})
	outputAll(program, Let{})
	outputAll(program, If{})
}

func outputAll(program *bit.Program, key bit.Key) {
	program.DeclareGlobal(key, Output{})
}

func init() {
	assertCode[bit.Module]()
	assertCode[Block]()
	assertCode[Group]()
	assertCode[Line]()
	assertCode[Var]()
	assertCode[Let]()
	assertCode[Bool]()
	assertCode[Int]()
	assertCode[Str]()
	assertCode[Print]()
	assertCode[If]()
}

func assertCode[T bit.HasOutput]() {}
