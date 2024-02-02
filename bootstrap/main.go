package main

import (
	"fmt"
	"os"

	"axlab.dev/bit/output"
	"axlab.dev/bit/proc"
	"axlab.dev/bit/text"
)

func main() {
	proc.Bootstrap()

	fmt.Printf("-> WorkDir: %s\n", proc.WorkingDir())
	fmt.Printf("-> Args:    %v\n", os.Args)

	fmt.Printf("-> Main:    %s\n", proc.FileName())
	fmt.Printf("-> Exe:     %s\n", proc.GetBootstrapExe())
	fmt.Printf("-> Project: %s\n", proc.ProjectDir())

	build := output.Open("./build")
	build.Write("src/main.c", text.Cleanup(`
		#include <stdio.h>

		int main() {
			printf("hello world\n");
			return 0;
		}
	`))

	if proc.Run("CC", "gcc", "./build/src/main.c", "-o", "./build/output.exe") {
		exitCode := proc.Spawn("./build/output.exe")
		fmt.Printf("\nexited with %d\n", exitCode)
	} else {
		fmt.Printf("\nCompilation failed\n")
	}

	fmt.Printf("\n")

}
