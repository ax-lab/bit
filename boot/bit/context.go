package bit

import (
	"fmt"
	"sync"
	"time"

	"axlab.dev/bit/boot/core"
)

type Context struct {
	core.ErrorList

	queue *EvalQueue

	lexer   Lexer
	rootDir core.DirEntry
	sources core.SourceMap

	loading sync.WaitGroup

	moduleSync   sync.Mutex
	moduleBySrc  map[core.Source]*Module
	moduleByName map[string]*Module
	moduleByFile map[string]*Module
	moduleList   []*Module
}

func contextNew(rootDir string, lexer Lexer) (*Context, error) {
	root, err := core.Dir(rootDir)
	if err != nil {
		return nil, err
	}

	ctx := &Context{
		rootDir: root,
		lexer:   lexer,
		sources: core.SourceMapNew(root),

		moduleBySrc:  make(map[core.Source]*Module),
		moduleByName: make(map[string]*Module),
		moduleByFile: make(map[string]*Module),
	}

	ctx.queue = evalQueueNew(ctx)

	return ctx, nil
}

func (ctx *Context) Queue() *EvalQueue {
	return ctx.queue
}

func (ctx *Context) GetModule(name string) *Module {
	ctx.moduleSync.Lock()
	defer ctx.moduleSync.Unlock()

	mod := ctx.moduleByName[name]
	if mod == nil {
		mod = ctx.doLoadFile(name, false)
		ctx.moduleByName[name] = mod
	}

	return mod
}

func (ctx *Context) DeclareModule(name, text string) *Module {
	if name == "" {
		panic("DeclareModule: name cannot be empty")
	}

	ctx.moduleSync.Lock()
	defer ctx.moduleSync.Unlock()

	src := ctx.sources.LoadString(name, text)
	modBySrc := ctx.doLoadSource(name, src, nil)

	modByName := ctx.moduleByName[name]
	if modByName == nil {
		ctx.moduleByName[name] = modBySrc
	} else if modBySrc != modByName {
		ctx.AddError(fmt.Errorf("module `%s` was declared multiple times for context", name))
	}

	return modBySrc
}

func (ctx *Context) LoadFile(fileName string) *Module {
	return ctx.doLoadFile(fileName, true)
}

func (ctx *Context) Eval() (out Result) {
	loaded := make(chan struct{})

	go func() {
		defer close(loaded)
		ctx.loading.Wait()
		ctx.queue.Start()
		ctx.queue.Wait()
	}()

	timeout := time.After(30 * time.Second)
	select {
	case <-timeout:
		out.AddError(fmt.Errorf("context loading timed out"))
		return
	case <-loaded:
	}

	var modules []*Module
	ctx.moduleSync.Lock()
	modules = append(modules, ctx.moduleList...)
	ctx.moduleSync.Unlock()

	out.MergeErrors(&ctx.ErrorList)
	for _, mod := range modules {
		mod.Wait()
		out.MergeErrors(&mod.ErrorList)
	}

	return
}

func (ctx *Context) doLoadFile(fileName string, lock bool) *Module {
	if lock {
		ctx.moduleSync.Lock()
		defer ctx.moduleSync.Unlock()
	}

	if mod := ctx.moduleByFile[fileName]; mod != nil {
		return mod
	}

	src, err := ctx.sources.LoadFile(fileName)
	name := fileName
	if src.Valid() {
		name = src.Name()
	}

	mod := ctx.doLoadSource(name, src, err)
	ctx.moduleByFile[fileName] = mod

	return mod
}

func (ctx *Context) doLoadSource(name string, src core.Source, err error) *Module {
	if err != nil {
		mod := moduleNew(name, src, ctx.lexer.Copy())
		mod.initError(ctx, err)
		ctx.moduleList = append(ctx.moduleList, mod)
		return mod
	}

	if !src.Valid() {
		panic("loading module from invalid source")
	}

	mod := ctx.moduleBySrc[src]
	if mod == nil {
		mod = moduleNew(name, src, ctx.lexer.Copy())
		ctx.moduleList = append(ctx.moduleList, mod)
		ctx.moduleBySrc[src] = mod

		ctx.loading.Add(1)
		go func() {
			defer ctx.loading.Done()
			mod.init(ctx)
		}()
	}

	return mod
}
