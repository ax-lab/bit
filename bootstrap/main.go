package main

import (
	"fmt"

	"axlab.dev/bit/output"
	"axlab.dev/bit/proc"
	"axlab.dev/bit/text"
)

func main() {
	build := output.Open("./build")
	build.Write("src/main.c", text.Cleanup(`
		#include <stdio.h>

		int main() {
			printf("hello world\n");
			return 0;
		}
	`))

	if proc.Run("CC", "gcc", "./build/src/main.c", "-o", "./build/output.exe") {
		proc.Replace("./build/output.exe")
	} else {
		fmt.Println("\nCompilation failed")
	}

	fmt.Printf("\n")
}
