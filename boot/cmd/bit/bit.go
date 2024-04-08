package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello from bit!!!")
	fmt.Println(os.Args)

	fmt.Printf("Input text: ")

	var input string
	fmt.Scanln(&input)
	fmt.Printf("> %s\n\n", input)

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
