package lang

import (
	"axlab.dev/bit/core"
)

func Declare(compiler *core.Compiler) error {
	compiler.DeclareOp(OpSegment)
	InitLexer(&compiler.Lexer)

	return nil
}

func InitLexer(lexer *core.Lexer) {
	lexer.AddMatcher(MatchString)
	lexer.AddMatcher(MatchWord)
	lexer.AddMatcher(MatchNumber)
	lexer.AddMatcher(MatcherLineComment("#", "//"))
	lexer.AddMatcher(MatcherBlockComment("/* */", "/# #/"))

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
