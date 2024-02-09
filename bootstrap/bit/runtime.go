package bit

import (
	"fmt"
	"io"
	"os"
)

type Result interface {
	String() string
}

func ResultRepr(res Result) string {
	if IsError(res) {
		return fmt.Sprintf("Error(%s)", res.String())
	} else if _, empty := res.(emptyResult); empty {
		return "(none)"
	} else {
		return fmt.Sprintf("%#v", res.String())
	}
}

func IsError(res Result) bool {
	_, ok := res.(error)
	return ok
}

func NewRuntime(node *Node) *RuntimeContext {
	return &RuntimeContext{
		Parent: nil,
		Source: node,
		StdOut: os.Stdout,
		StdErr: os.Stderr,
	}
}

type RuntimeContext struct {
	Parent *RuntimeContext
	Result Result
	Source *Node
	StdOut io.Writer
	StdErr io.Writer
}

func (rt *RuntimeContext) Done() bool {
	return IsError(rt.Result)
}

func (rt *RuntimeContext) Error(msg string, args ...any) {
	rt.Result = rt.ErrorResult(msg, args...)
}

func (rt *RuntimeContext) ErrorResult(msg string, args ...any) Result {
	return RuntimeError{Span: rt.Source.Span(), Message: msg, Args: args}
}

func (rt *RuntimeContext) EmptyResult() Result {
	return emptyResult{}
}

func (rt *RuntimeContext) OutputStd(text string) {
	io.WriteString(rt.StdOut, text)
}

func (rt *RuntimeContext) Eval(code Code) Result {
	if rt.Done() {
		return rt.EmptyResult()
	}

	sub := *rt
	sub.Parent = rt
	sub.Result = rt.EmptyResult()
	sub.Source = code.Node
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
