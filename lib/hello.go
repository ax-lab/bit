package lib

import "fmt"

func Answer() int {
	return 42
}

func SayHello() {
	fmt.Printf("\nThe answer to life, the universe, and everything is %d\n\n", Answer())
}
