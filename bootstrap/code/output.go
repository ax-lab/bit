package code

import "fmt"

type Output struct {
	errors []error
	main   *Block
}

func NewOutput(scope *Scope, typ Type, source any) *Output {
	out := &Output{
		main: NewBlockWithScope(scope),
	}
	return out
}

func (output *Output) NewContext() *OutputContext {
	return &OutputContext{
		source: output,
		main:   output.main,
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

	// TODO: improve or simplify name allocation?
	output.main.Scope().ProcessNames()
	output.main.OutputCpp(ctx)
	return ctx.GetOutputFiles(mainFile)
}

type OutputContext struct {
	main   *Block
	source *Output
}

func (ctx *OutputContext) NewScope(scope *Scope) *OutputContext {
	out := *ctx
	out.main = NewBlockWithScope(scope)
	return &out
}

func (ctx *OutputContext) NewBlock() *OutputContext {
	out := *ctx
	out.main = NewBlock(ctx.main.Decl)
	return &out
}

func (ctx *OutputContext) GetDecl() *Decl {
	return ctx.main.Decl
}

func (ctx *OutputContext) TempVar(name string, typ Type, source any) *Variable {
	scope := ctx.main.Scope()
	v := scope.DeclareUnique(name, typ, source)
	ctx.main.Decl.Add(v)
	return v
}

func (ctx *OutputContext) Valid() bool {
	return len(ctx.source.errors) == 0
}

func (ctx *OutputContext) Error(err error) {
	ctx.source.errors = append(ctx.source.errors, err)
}

func (ctx *OutputContext) Output(stmt ...Stmt) {
	for _, it := range stmt {
		if it != nil {
			ctx.main.Body = append(ctx.main.Body, it)
		}
	}
}

func (ctx *OutputContext) Block() Stmt {
	return ctx.main
}
