package bit

import (
	"fmt"

	"axlab.dev/bit/boot/core"
)

type Compiler struct {
	Lexer    Lexer
	rootDir  string
	mainFile string
}

func (comp *Compiler) SetRoot(dir string) {
	comp.rootDir = dir
}

func (comp *Compiler) SetMain(mainFile string) {
	comp.mainFile = mainFile
}

func (comp *Compiler) NewContext() (*Context, error) {
	root := comp.rootDir
	if root == "" {
		root = "."
	}

	ctx, err := contextNew(root, comp.Lexer.Copy())
	return ctx, err
}

func (comp *Compiler) Run() (err error) {
	ctx, err := comp.NewContext()
	if err != nil {
		return err
	}

	if comp.mainFile == "" {
		return fmt.Errorf("main file not set")
	}

	ctx.LoadFile(comp.mainFile)

	res := ctx.Eval()
	if res.HasErrors() {
		errs := res.Errors()
		if len(errs) == 1 {
			return errs[0]
		}

		core.ErrorSort(errs)
		return core.ErrorFromList(errs, "%d errors compiling `%s`", len(errs), comp.mainFile)
	}

	return nil
}
