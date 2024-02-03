package bit

import (
	"axlab.dev/bit/files"
)

type Compiler struct {
	inputDir files.Dir
	buildDir files.Dir
}

func NewCompiler(inputPath, buildPath string) *Compiler {
	return &Compiler{
		inputDir: files.OpenDir(inputPath).MustExist("compiler input"),
		buildDir: files.OpenDir(buildPath).Create("compiler build"),
	}
}

func (comp *Compiler) InputDir() files.Dir {
	return comp.inputDir
}

func (comp *Compiler) BuildDir() files.Dir {
	return comp.buildDir
}
