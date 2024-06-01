package golang

import (
	"fmt"
	"slices"
	"sync"

	"axlab.dev/bit/core"
)

const (
	GoVersion = "1.21.3"
)

type Context struct {
	sync sync.Mutex

	importSet  map[string]bool
	importList []string

	declHead []string
	declBody []*Block

	mainBody Block
}

func (ctx *Context) Main() *Block {
	ctx.initBlocks()
	return &ctx.mainBody
}

func (ctx *Context) Import(includes ...string) {
	ctx.sync.Lock()
	defer ctx.sync.Unlock()
	if ctx.importSet == nil {
		ctx.importSet = make(map[string]bool)
	}
	for _, it := range includes {
		if !ctx.importSet[it] {
			ctx.importSet[it] = true
			ctx.importList = append(ctx.importList, it)
		}
	}
}

func (ctx *Context) DeclareBlock(header string, args ...any) *Block {
	ctx.sync.Lock()
	defer ctx.sync.Unlock()

	if len(args) > 0 {
		header = fmt.Sprintf(header, args...)
	}

	body := &Block{context: ctx}
	ctx.declHead = append(ctx.declHead, header)
	ctx.declBody = append(ctx.declBody, body)

	return body
}

func (ctx *Context) GenerateOutput(module string, mainFile string, output *core.OutputSet) {
	file := core.CodeText{}

	file.WriteLine(`package main`)
	file.BlankLine()

	if len(ctx.importList) > 0 {
		slices.Sort(ctx.importList)
		file.WriteLine(`import (`)
		file.Indent()
		for _, it := range ctx.importList {
			file.WriteLine(StringLiteral(it))
		}
		file.Dedent()
		file.WriteLine(`)`)
		file.BlankLine()
	}

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
	file.WriteLine(`func main() {`)
	file.Indent()
	file.WriteLine(ctx.mainBody.String())
	file.Dedent()
	file.WriteLine(`}`)

	text := file.String()
	output.Add(mainFile, text)

	output.Add("go.work", fmt.Sprintf("go %s\n\nuse (\n\t.\n)\n", GoVersion))
	output.Add("go.mod", fmt.Sprintf("module %s\n\ngo %s\n", module, GoVersion))
}

func (ctx *Context) initBlocks() {
	ctx.mainBody.context = ctx
}
