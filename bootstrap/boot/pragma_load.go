package boot

import (
	"fmt"

	"axlab.dev/bit/input"
)

const (
	PragmaLoadHeaderSize = 1024
	PragmaLoadPrefix     = "#load"
)

func (st *State) PragmaLoad(node Node, name string) error {
	name = input.Trim(name)
	if name == "" {
		return fmt.Errorf("%s -- missing name", PragmaLoadPrefix)
	}

	switch name {
	case "boot.lexer":
		return st.BootLoadLexer()
	case "boot.print":
		return st.BootLoadPrint()
	default:
		return fmt.Errorf("%s -- invalid package `%s`", PragmaLoadPrefix, name)
	}
}
