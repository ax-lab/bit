package text_test

import (
	"testing"

	"axlab.dev/bit/text"
	"github.com/stretchr/testify/require"
)

func TestCleanupText(t *testing.T) {
	test := require.New(t)
	test.Equal("",
		text.Cleanup(""))
	test.Equal("A\nB\nC\nD\n",
		text.Cleanup("A\nB\r\nC\rD\n"))
	test.Equal("A\nB\nC\n",
		text.Cleanup("A  \nB  \nC  \n  "))
	test.Equal("A\nB\n  C\n  D\nE\n",
		text.Cleanup("  \n  A\n  B\n    C\n    D\n  E\n  "))
	test.Equal("A\nB\n    C\n    D\nE\n",
		text.Cleanup(text.ExpandTabs("  \n\tA\n\tB\n\t\tC\n\t\tD\n\tE\n  ", 4)))
}

func TestIndent(t *testing.T) {
	tab := "  "
	test := require.New(t)
	test.Equal("", text.Indent(""))
	test.Equal("  ABC", text.Indent("ABC", tab))
	test.Equal("  L1\n  L2", text.Indent("L1\nL2", tab))
	test.Equal("", text.Indent(""))
	test.Equal("  ABC", text.Indent("ABC", tab))
	test.Equal("  L1\n  L2\n\n  L3", text.Indent("L1\nL2\n\nL3", tab))
}

func TestExpandTabs(t *testing.T) {
	test := require.New(t)
	test.Equal("  L1\n    L2\n  L3a\tL3b", text.ExpandTabs("\tL1\n\t\tL2\n\tL3a\tL3b", 2))
}

func TestAddTrailingNewLine(t *testing.T) {
	test := require.New(t)
	test.Equal("", text.AddTrailingNewLine(""))
	test.Equal("A\n", text.AddTrailingNewLine("A"))
	test.Equal("A\nB\n", text.AddTrailingNewLine("A\nB"))
	test.Equal("A\nB\n", text.AddTrailingNewLine("A\nB\n"))
	test.Equal("A\nB\r", text.AddTrailingNewLine("A\nB\r"))
	test.Equal("A\nB\r\n", text.AddTrailingNewLine("A\nB\r\n"))
	test.Equal("A\nB\r\n\n", text.AddTrailingNewLine("A\nB\r\n\n"))
}
