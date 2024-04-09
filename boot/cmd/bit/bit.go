package main

import (
	"fmt"
	"path/filepath"

	"axlab.dev/bit/boot/core"
)

const (
	DirSrc = "src"
)

func main() {
	rootDir := core.ProjectRoot()
	rootSrc := filepath.Join(rootDir, DirSrc)

	root := core.Check(core.FS(rootSrc))

	for _, it := range core.CheckErrs(root.Glob("*.bit")) {
		fmt.Println(it.Path())
	}

	//

	// prj := core.ProjectNew("bit")
	// program := bit.Program(prj)

	// prj.SetBase(core.RepoRoot())
	// prj.Depends("./boot/**.go")
	// prj.Source("./src/**.bit", bit.SourceLoader(program))
	// prj.Step(bit.Compiler(program))
	// prj.Step(bit.Builder(program))
	// prj.ExecMain(program.OutputExe(), os.Args[1:]...)
	// prj.Run()
}
