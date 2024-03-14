package boot

import "fmt"

func (st *State) BootLoadLexer() error {
	fmt.Println("Loaded boot.lexer")
	return nil
}
