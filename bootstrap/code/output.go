package code

import "fmt"

type Output struct {
	errors []error
	main   *Block
}

func NewOutput(scope *Scope) *Output {
	return &Output{
		main: NewBlockWithScope(scope),
	}
}

func (output *Output) NewContext() *OutputContext {
	return &OutputContext{
		source: output,
		block:  output.main,
	}
}

func (output *Output) Valid() bool {
	return len(output.errors) == 0
}

func (output *Output) Errors() []error {
	return output.errors
}

func (output *Output) Eval(rt *Runtime) (Value, error) {
	if len(output.errors) > 0 {
		return nil, fmt.Errorf("cannot evaluate output code with errors")
	}
	return output.main.Eval(rt)
}

func (output *Output) Repr(repr Repr) string {
	return fmt.Sprintf("Output(%s)", output.main.Repr(repr))
}

func (output *Output) OutputCpp(mainFile string) map[string]string {
	ctx := NewCppContext()
	return ctx.GetOutputFiles(mainFile)
}

type OutputContext struct {
	expr  Expr
	block *Block

	source *Output
}

func (ctx *OutputContext) WithScope(scope *Scope) *OutputContext {
	out := *ctx
	out.block = NewBlockWithScope(scope)
	return &out
}

func (ctx *OutputContext) NewBlock() *OutputContext {
	out := *ctx
	out.block = NewBlock(ctx.block.Decl)
	return &out
}

func (ctx *OutputContext) GetDecl() *Decl {
	return ctx.block.Decl
}

func (ctx *OutputContext) TempVar(name string, typ Type, source any) *Variable {
	scope := ctx.block.Scope()
	v := scope.DeclareUnique(name, typ, source)
	ctx.block.Decl.Add(v)
	return v
}

func (ctx *OutputContext) Valid() bool {
	return len(ctx.source.errors) == 0
}

func (ctx *OutputContext) Error(err error) {
	ctx.source.errors = append(ctx.source.errors, err)
}

func (ctx *OutputContext) OutputExpr(expr Expr) {
	ctx.expr = expr
}

func (ctx *OutputContext) Output(stmt ...Stmt) {
	ctx.block.Body = append(ctx.block.Body, stmt...)
}

func (ctx *OutputContext) LastExpr() Expr {
	return ctx.expr
}

func (ctx *OutputContext) Body() []Stmt {
	return ctx.block.Body
}

func (ctx *OutputContext) Block() *Block {
	return ctx.block
}
