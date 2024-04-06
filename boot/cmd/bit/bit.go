package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"axlab.dev/bit/boot/core"
)

func main() {
	cmd := core.Cmd("go", "run", "cmd/run/bit_run.go")
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nFatal: %v\n\n", err)
		os.Exit(-1)
	}

	if err := strings.TrimRightFunc(cmd.StdErr(), unicode.IsSpace); err != "" {
		fmt.Fprintf(os.Stderr, "\n>>> Error:\n\n%s\n", err)
	}

	if out := cmd.StdOut(); len(out) > 0 {
		fmt.Fprintf(os.Stdout, "\n>>> Output:\n\n%s<EOF>\n", out)
	}

	fmt.Printf("\n-> Exit code = %d\n\n", cmd.ExitCode())

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
