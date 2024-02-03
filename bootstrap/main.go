package main

import (
	"axlab.dev/bit/bit"
	"axlab.dev/bit/logs"
	"axlab.dev/bit/proc"
	"axlab.dev/bit/text"
)

const sampleC = false

func main() {
	proc.Bootstrap()

	compiler := bit.NewCompiler("sample", "build")
	inputDir := compiler.InputDir()
	buildDir := compiler.BuildDir()

	logs.Sep()
	logs.Out("-> Input: %s\n", inputDir.FullPath())
	logs.Out("-> Build: %s\n", buildDir.FullPath())
	logs.Sep()

	compiler.Watch()

	if sampleC {
		main := buildDir.Write("src/main.c", text.Cleanup(`
			#include <stdio.h>

			int main() {
				printf("hello world\n");
				return 42;
			}
		`))

		output := buildDir.GetFullPath("output.exe")
		if proc.Run("CC", "gcc", main.FullPath(), "-o", output) {
			logs.Sep()
			if exitCode := proc.Spawn(output); exitCode != 0 {
				logs.Out("\n(exited with %d)\n", exitCode)
			} else {
				logs.Out("\n")
			}
		} else {
			logs.Out("\nCompilation failed\n")
		}
	}

	logs.Sep()
}
