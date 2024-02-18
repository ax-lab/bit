package bit_core

import (
	"axlab.dev/bit/bit"
	"axlab.dev/bit/code"
)

// TODO: move span and sources to a common library to enable interop with code

type Key = bit.Key
type Node = bit.Node
type Span = bit.Span
type Symbol = bit.Symbol
type Type = code.Type
type Value = bit.Value
type Word = bit.Word

func InitCompiler(program *bit.Program) {
	program.DeclareGlobal(bit.TokenBreak, SplitLines{})
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

	program.DeclareGlobal(bit.TokenString, Replace{ParseString{}})
	program.DeclareGlobal(bit.TokenInteger, Replace{ParseInt{}})
	program.DeclareGlobal(Word("true"), Replace{ParseBool{}})
	program.DeclareGlobal(Word("false"), Replace{ParseBool{}})

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

func outputAll(program *bit.Program, key Key) {
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
