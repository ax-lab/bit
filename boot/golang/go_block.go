package golang

import (
	"fmt"
	"strings"
	"sync/atomic"
	"unicode"

	"axlab.dev/bit/core"
)

const (
	TempVarPrefix = "tmp_"
)

type Var string

type Block struct {
	core.CodeText

	context *Context

	tmp  atomic.Int32
	decl core.CodeText
}

func (block *Block) Context() *Context {
	return block.context
}

func (block *Block) Declare(code string, args ...any) {
	block.decl.NewLine()
	block.decl.WriteLine(code, args...)
}

func (block *Block) Import(includes ...string) {
	block.context.Import(includes...)
}

func (block *Block) NewVar(name ...string) Var {
	if len(name) > 1 {
		panic("Block: invalid arguments to NewVar")
	}

	var prefix string
	if len(name) > 0 {
		prefix = name[0]
	}

	var out string
	next := block.tmp.Add(1)
	if len(prefix) > 0 {
		out = fmt.Sprintf("%s%s_%04d", TempVarPrefix, prefix, next)
	} else {
		out = fmt.Sprintf("%s%04d", TempVarPrefix, next)
	}

	return Var(out)
}

func (block *Block) String() string {
	out := strings.Builder{}

	if decl := block.decl.String(); len(decl) > 0 {
		out.WriteString(decl)
	}

	body := block.CodeText.String()
	body = strings.TrimRightFunc(body, unicode.IsSpace)
	out.WriteString(body)

	return out.String()
}
