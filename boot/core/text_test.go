package core_test

import (
	"testing"

	"axlab.dev/bit/boot/core"
	"github.com/stretchr/testify/require"
)

func TestLines(t *testing.T) {
	test := require.New(t)
	test.Equal([]string{""}, core.Lines(""))
	test.Equal([]string{"A", "B", "C", ""}, core.Lines("A\nB\nC\n"))
	test.Equal([]string{"A", "B", "C", ""}, core.Lines("A\rB\rC\r"))
	test.Equal([]string{"A", "B", "C", ""}, core.Lines("A\r\nB\r\nC\r\n"))
	test.Equal([]string{"A", "B", "C", ""}, core.Lines("A\nB\rC\r\n"))
}

func TestText(t *testing.T) {
	test := require.New(t)
	text := core.Text(`
		L1
			L2
			L3
				L3.1
				L3.2
		L4
	`)
	test.Equal("L1\n\tL2\n\tL3\n\t\tL3.1\n\t\tL3.2\nL4\n", text)
}
