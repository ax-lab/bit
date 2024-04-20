package bit

import (
	"fmt"
	"sync"

	"axlab.dev/bit/boot/core"
)

type Context struct {
	core.ErrorList

	sources core.SourceMap

	moduleSync sync.Mutex
	moduleMap  map[core.Source]*Module
	moduleList []*Module
}

func contextNew(rootDir string) *Context {
	root, err := core.Dir(rootDir)

	out := &Context{
		sources: core.SourceMapNew(root),
	}
	if err != nil {
		out.AddError(err)
	}
	return out
}

func (ctx *Context) LoadFile(fileName string) {
	src, err := ctx.sources.LoadFile(fileName)
	if err != nil {
		ctx.AddError(err)
	}

	ctx.moduleSync.Lock()
	defer ctx.moduleSync.Unlock()

	mod := ctx.moduleMap[src]
	if mod == nil {
		mod = &Module{src: src}
		if ctx.moduleMap == nil {
			ctx.moduleMap = make(map[core.Source]*Module)
		}
		ctx.moduleMap[src] = mod
		ctx.moduleList = append(ctx.moduleList, mod)
	}
}

func (ctx *Context) Eval() {
	ctx.AddError(fmt.Errorf("Eval not implemented"))
}
