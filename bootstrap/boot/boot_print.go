package boot

import "fmt"

func (st *State) BootLoadPrint() error {
	fmt.Println("Loaded boot.print")
	return nil
}
