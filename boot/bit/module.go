package bit

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"axlab.dev/bit/boot/core"
)

type Module struct {
	core.ErrorList

	name string

	src core.Source
	lex Lexer

	ctx  *Context
	load chan struct{}

	depSync sync.RWMutex
	depMap  map[*Module]bool
}

func moduleNew(name string, src core.Source, lex Lexer) *Module {
	if name == "" {
		panic("module name cannot be empty")
	}

	out := &Module{
		name:   name,
		src:    src,
		lex:    lex,
		load:   make(chan struct{}),
		depMap: make(map[*Module]bool),
	}
	return out
}

func (mod *Module) IsLoaded() bool {
	select {
	case <-mod.load:
		return true
	default:
		return false
	}
}

func (mod *Module) Name() string {
	return mod.name
}

func (mod *Module) Wait() {
	<-mod.load
}

func (mod *Module) LoadModule(name string) (*Module, error) {
	mod.checkModule()

	modPath := mod.src.Path()
	if modPath != "" {
		modPath = path.Join(modPath, name)
	} else {
		modPath = name
	}

	if modPath == "" || modPath == "." {
		return nil, fmt.Errorf("loading module: invalid module name")
	}

	modPath, err := mod.ctx.rootDir.Resolve(modPath)
	if err != nil {
		return nil, err
	}

	loadedMod := mod.ctx.GetModule(modPath)
	if loadedMod == mod {
		return nil, fmt.Errorf("loading module: cannot load own module")
	}

	mod.depSync.Lock()
	mod.depMap[loadedMod] = true
	mod.depSync.Unlock()

	if err := mod.checkCycle(loadedMod); err != nil {
		return nil, err
	}

	return loadedMod, nil
}

func (mod *Module) initError(ctx *Context, err error) {
	mod.ctx = ctx
	mod.AddError(err)
	close(mod.load)
}

func (mod *Module) init(ctx *Context) {
	defer close(mod.load)
	mod.ctx = ctx
	mod.checkModule()

	for _, line := range strings.Split(mod.src.Text(), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSpace(line[1:])
		cmd, arg, ok := strings.Cut(line, " ")
		if ok {
			arg = strings.TrimSpace(arg)
			switch cmd {
			case "error":
				mod.AddError(fmt.Errorf("module %s: #error %v", mod.Name(), arg))
			case "load":
				if loaded, err := mod.LoadModule(arg); err != nil {
					mod.AddError(fmt.Errorf("module %s: loading `%s`: %v", mod.Name(), arg, err))
				} else {
					fmt.Printf("[module %s] loaded %s\n", mod.Name(), loaded.Name())
					loaded.Wait()
				}
			case "print":
				fmt.Printf("[module %s] %s\n", mod.Name(), arg)
			}
		}
	}
}

func (mod *Module) checkModule() {
	if mod.ctx == nil {
		panic("module not initialized")
	}
	if !mod.src.Valid() {
		panic("module with invalid source")
	}
}

func (mod *Module) checkCycle(loadedMod *Module) error {
	checked := make(map[*Module]bool)
	queue := []*Module{loadedMod}
	for len(queue) > 0 {
		last := len(queue) - 1
		next := queue[last]
		queue = queue[:last]

		if next == mod {
			return fmt.Errorf("loading `%s` from `%s` creates a cyclic dependency", loadedMod.Name(), mod.Name())
		}

		next.depSync.Lock()
		for it := range next.depMap {
			if checked[it] {
				panic(fmt.Sprintf("detected cyclic dependency between `%s` and `%s`", loadedMod.Name(), it.Name()))
			}
			queue = append(queue, it)
			checked[it] = true
		}
		next.depSync.Unlock()
	}

	return nil
}
