package lang

import (
	"axlab.dev/bit/core"
)

func Declare(compiler *core.Compiler) error {
	compiler.DeclareOp(OpSegment)
	compiler.SetOutput(NoOp)

	InitLexer(&compiler.Lexer)

	return nil
}

func NoOp(list core.NodeList) {
}

func InitLexer(lexer *core.Lexer) {
	lexer.AddBrackets("(", ")")
	lexer.AddBrackets("[", "]")
	lexer.AddBrackets("{", "}")
	lexer.AddSymbols(
		// punctuation
		".", "..", ",", ";", ":",
		// operators
		"!", "?",
		"=", "+", "-", "*", "/", "%",
		"==", "!=", "<", "<=", ">", ">=",
	)
}
