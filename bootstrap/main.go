package main

import (
	"os"

	"axlab.dev/bit/files"
	"axlab.dev/bit/logs"
	"axlab.dev/bit/output"
	"axlab.dev/bit/proc"
	"axlab.dev/bit/text"
)

func main() {
	proc.Bootstrap()

	logs.Out("-> WorkDir: %s\n", proc.WorkingDir())
	logs.Out("-> Args:    %v\n", os.Args)

	logs.Out("-> Main:    %s\n", proc.FileName())
	logs.Out("-> Exe:     %s\n", proc.GetBootstrapExe())
	logs.Out("-> Project: %s\n", proc.ProjectDir())

	build := output.Open("./build")
	build.Write("src/main.c", text.Cleanup(`
		#include <stdio.h>

		int main() {
			printf("hello world\n");
			return 0;
		}
	`))

	if proc.Run("CC", "gcc", "./build/src/main.c", "-o", "./build/output.exe") {
		logs.Sep()
		exitCode := proc.Spawn("./build/output.exe")
		logs.Out("\nexited with %d\n", exitCode)
	} else {
		logs.Out("\nCompilation failed\n")
	}

	logs.Sep()
	for _, it := range files.List(".") {
		logs.Break()
		logs.Out("%s", it.String())
	}

	logs.Sep()
}
