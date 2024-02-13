package core

import "axlab.dev/bit/bit"

type Program = bit.Program
type Symbol = bit.Symbol
type Key = bit.Key
type Word = bit.Word
type Node = bit.Node
type Binding = bit.Binding
type BindArgs = bit.BindArgs
type Precedence = bit.Precedence
type Code = bit.Code
type CodeContext = bit.CodeContext
type RuntimeContext = bit.RuntimeContext
type CppContext = bit.CppContext
type TokenType = bit.TokenType
type Value = bit.Value
type Type = bit.Type
type Span = bit.Span

func InitCompiler(program *Program) {
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

	outputAll(program, Str(""))
	outputAll(program, Int(0))

	program.DeclareGlobal(Line{}, Simplify{})
	program.DeclareGlobal(Group{}, Simplify{})

	outputAll(program, bit.Module{})
	outputAll(program, Block{})
	outputAll(program, Print{})
	outputAll(program, Var{})
	outputAll(program, Let{})
	outputAll(program, If{})
}

func outputAll(program *Program, key Key) {
	program.DeclareGlobal(key, Output{})
}
