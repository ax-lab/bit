package cpp

import (
	"fmt"
	"sync"

	"axlab.dev/bit/core"
)

type Context struct {
	sync sync.Mutex

	includeLocalSet  map[string]bool
	includeLocalList []string

	includeSystemSet  map[string]bool
	includeSystemList []string

	declHead []string
	declBody []*Block

	mainDecl Block
	mainBody Block
}

func (ctx *Context) IncludeLocal(includes ...string) {
	ctx.sync.Lock()
	defer ctx.sync.Unlock()
	if ctx.includeLocalSet == nil {
		ctx.includeLocalSet = make(map[string]bool)
	}
	for _, it := range includes {
		if !ctx.includeLocalSet[it] {
			ctx.includeLocalSet[it] = true
			ctx.includeLocalList = append(ctx.includeLocalList, it)
		}
	}
}

func (ctx *Context) IncludeSystem(includes ...string) {
	ctx.sync.Lock()
	defer ctx.sync.Unlock()
	if ctx.includeSystemSet == nil {
		ctx.includeSystemSet = make(map[string]bool)
	}
	for _, it := range includes {
		if !ctx.includeSystemSet[it] {
			ctx.includeSystemSet[it] = true
			ctx.includeSystemList = append(ctx.includeSystemList, it)
		}
	}
}

func (ctx *Context) Main() *Block {
	ctx.sync.Lock()
	defer ctx.sync.Unlock()

	ctx.initBlocks()
	return &ctx.mainBody
}

func (ctx *Context) DeclareFunction(header string, args ...any) *Block {
	ctx.sync.Lock()
	defer ctx.sync.Unlock()

	if len(args) > 0 {
		header = fmt.Sprintf(header, args...)
	}

	ctx.mainDecl.BlankLine()
	ctx.mainDecl.WriteLine("%s;", header)

	body := &Block{context: ctx}
	ctx.declHead = append(ctx.declHead, header)
	ctx.declBody = append(ctx.declBody, body)

	return body
}

func (ctx *Context) GenerateOutput(mainFile string, output *core.OutputSet) {
	file := core.CodeText{}

	for _, it := range ctx.includeSystemList {
		file.WriteLine(`#include <%s>`, it)
	}
	if len(ctx.includeSystemList) > 0 {
		file.BlankLine()
	}

	for _, it := range ctx.includeLocalList {
		file.WriteLine(`#include "%s"`, it)
	}
	if len(ctx.includeLocalList) > 0 {
		file.BlankLine()
	}

	file.WriteLine(ctx.mainDecl.String())

	for idx, header := range ctx.declHead {
		body := ctx.declBody[idx]
		file.BlankLine()
		file.WriteLine("%s {", header)
		file.Indent()
		file.WriteLine(body.String())
		file.Dedent()
		file.WriteLine("}")
	}

	file.BlankLine()
	file.WriteLine(`int main() {`)
	file.Indent()
	file.WriteLine(ctx.mainBody.String())
	file.Dedent()
	file.WriteLine(`}`)

	text := file.String()
	output.Add(mainFile, text)
}

func (ctx *Context) initBlocks() {
	ctx.mainDecl.context = ctx
	ctx.mainBody.context = ctx
}
