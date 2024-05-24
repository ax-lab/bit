package main

import (
	"fmt"
	"os"

	"axlab.dev/bit/core"
)

func main() {
	fmt.Printf("\nHello bit %s\n\n", core.Version())

	if len(os.Args) > 1 {
		for idx, arg := range os.Args[1:] {
			fmt.Printf("[%d] %#v\n", idx, arg)
		}
		fmt.Printf("\n")
	}
}
