package code

import (
	"fmt"
	"strings"

	"axlab.dev/bit/common"
)

type Block struct {
	Id
	Decl    *Decl
	Body    []Stmt
	Expr    Expr
	OwnDecl bool
}

func NewBlockWithScope(scope *Scope) *Block {
	return &Block{
		Decl:    NewDecl(scope),
		OwnDecl: true,
	}
}

func NewBlock(decl *Decl) *Block {
	return &Block{Decl: decl, OwnDecl: false}
}

func (blk *Block) Scope() *Scope {
	return blk.Decl.scope
}

func (blk *Block) Eval(rt *Runtime) (Value, error) {
	if blk.OwnDecl {
		blk.Decl.Init(rt)
		defer blk.Decl.Drop(rt)
	}

	if err := blk.Exec(rt); err != nil {
		return nil, err
	}

	if blk.Expr != nil {
		return blk.Expr.Eval(rt)
	}

	return nil, nil
}

func (blk *Block) Exec(rt *Runtime) error {
	if blk.OwnDecl {
		blk.Decl.Init(rt)
		defer blk.Decl.Drop(rt)
	}
	for _, it := range blk.Body {
		if err := it.Exec(rt); err != nil {
			return err
		}
	}
	return nil
}

func (blk *Block) OutputCpp(ctx *CppContext) {
	ctx.Body.Push("{")
	ctx.Body.Indent()
	for _, it := range blk.Body {
		ctx.Body.EnsureBlank()
		it.OutputCpp(ctx)
	}
	ctx.Body.Dedent()
	ctx.Body.Push("}")
}

func (blk *Block) Repr(mode Repr) string {
	if len(blk.Body) == 0 && (!blk.OwnDecl || blk.Decl.Len() == 0) {
		return "{ }"
	}

	switch mode {
	case ReprLabel:
		return fmt.Sprintf("block{%d}", len(blk.Body))

	case ReprLine:
		max := MaxLine - 3
		out := strings.Builder{}
		out.WriteString("{")
		for n, it := range blk.Body {
			if n == 0 {
				out.WriteString(" ")
			} else {
				out.WriteString("; ")
			}

			txt := it.Repr(ReprLine)
			if n > 0 && out.Len()+len(txt) > max {
				out.WriteString("…")
			} else {
				out.WriteString(txt)
			}
		}
		out.WriteString(" }")
		return out.String()

	default:
		out := strings.Builder{}
		out.WriteString("{")
		if blk.OwnDecl {
			out.WriteString("\n")
			out.WriteString(common.Indent(blk.Decl.String()))
		}
		for _, it := range blk.Body {
			out.WriteString("\n")
			out.WriteString(common.Indent(it.Repr(mode)))
		}
		out.WriteString("\n}")
		return out.String()
	}
}
