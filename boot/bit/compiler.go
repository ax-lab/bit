package bit

import (
	"fmt"

	"axlab.dev/bit/boot/core"
)

type Compiler struct {
	rootDir  string
	mainFile string
}

func (comp *Compiler) SetRoot(dir string) {
	comp.rootDir = dir
}

func (comp *Compiler) SetMain(mainFile string) {
	comp.mainFile = mainFile
}

func (comp *Compiler) Run() (err error) {
	root := comp.rootDir
	if root == "" {
		root = "."
	}

	ctx := contextNew(root)
	if comp.mainFile == "" {
		ctx.AddError(fmt.Errorf("main file not set"))
	} else {
		ctx.LoadFile(comp.mainFile)
	}

	ctx.Eval()
	if ctx.HasErrors() {
		errs := ctx.Errors()
		if len(errs) == 1 {
			return errs[0]
		}

		core.ErrorSort(errs)
		return core.ErrorFromList(errs, "%d errors compiling `%s`:\n", len(errs), comp.mainFile)
	}

	return nil
}
