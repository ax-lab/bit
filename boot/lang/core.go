package lang

import (
	"fmt"

	"axlab.dev/bit/core"
)

func Declare(compiler *core.Compiler) error {
	compiler.DeclareOp(OpSegment)
	compiler.DeclareOp(OpComment)
	compiler.DeclareOp(OpPrint)
	compiler.SetOutput(OutputCode)
	InitLexer(&compiler.Lexer)

	return nil
}

func InitLexer(lexer *core.Lexer) {
	lexer.SetSegmenter(ParseLine)

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

func OpComment(mod *core.Module, list core.NodeList) {
	list.RemoveIf(func(node core.Node) bool {
		_, isComment := node.Value().(core.Comment)
		return isComment
	})
}

func OpPrint(mod *core.Module, list core.NodeList) {
	if list.Len() == 0 {
		return
	}

	head := list.Get(0)
	if sym, ok := head.Value().(core.Word); !ok || sym != core.Word("print") {
		return
	}

	span := list.GetSpan()
	list.Remove(0)

	args := list.TakeList()

	node := core.NodeNew(span, PrintExpr{args})
	list.Push(node)
}

type PrintExpr struct {
	Args core.NodeList
}

func (expr PrintExpr) String() string {
	return fmt.Sprintf("PrintExpr(%s)", core.IndentBlock(expr.Args.String()))
}
