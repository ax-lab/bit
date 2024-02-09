package bit

import (
	"fmt"
	"os"
)

type Result interface {
	String() string
}

func IsError(res Result) bool {
	_, ok := res.(error)
	return ok
}

type RuntimeContext struct {
	Parent *RuntimeContext
	Result Result
	Source *Node
}

func (rt *RuntimeContext) Done() bool {
	return IsError(rt.Result)
}

func (rt *RuntimeContext) EmptyResult() Result {
	return emptyResult{}
}

func (rt *RuntimeContext) OutputStd(text string) {
	os.Stdout.WriteString(text)
}

func (rt *RuntimeContext) Eval(code Code) Result {
	if rt.Done() {
		return rt.EmptyResult()
	}

	sub := RuntimeContext{
		Parent: rt,
		Result: rt.EmptyResult(),
		Source: code.Node,
	}
	code.Expr.Eval(&sub)
	return sub.Result
}

func (rt *RuntimeContext) Panic(msg string, args ...any) {
	loc := rt.Location()
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	panic(fmt.Sprintf("%s (at %s)", msg, loc))
}

func (rt *RuntimeContext) Todo() {
	rt.Panic("not implemented")
}

func (rt *RuntimeContext) Location() string {
	return rt.Source.Span().String()
}

type emptyResult struct{}

func (emptyResult) String() string {
	return ""
}
