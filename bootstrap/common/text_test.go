package common_test

import (
	"testing"

	"axlab.dev/bit/common"
	"github.com/stretchr/testify/require"
)

func TestCleanupText(t *testing.T) {
	test := require.New(t)
	test.Equal("",
		common.Cleanup(""))
	test.Equal("A\nB\nC\nD\n",
		common.Cleanup("A\nB\r\nC\rD\n"))
	test.Equal("A\nB\nC\n",
		common.Cleanup("A  \nB  \nC  \n  "))
	test.Equal("A\nB\n  C\n  D\nE\n",
		common.Cleanup("  \n  A\n  B\n    C\n    D\n  E\n  "))
	test.Equal("A\nB\n    C\n    D\nE\n",
		common.Cleanup(common.ExpandTabs("  \n\tA\n\tB\n\t\tC\n\t\tD\n\tE\n  ", 4)))
}

func TestIndent(t *testing.T) {
	tab := "  "
	test := require.New(t)
	test.Equal("", common.Indent(""))
	test.Equal("  ABC", common.Indent("ABC", tab))
	test.Equal("  L1\n  L2", common.Indent("L1\nL2", tab))
	test.Equal("", common.Indent(""))
	test.Equal("  ABC", common.Indent("ABC", tab))
	test.Equal("  L1\n  L2\n\n  L3", common.Indent("L1\nL2\n\nL3", tab))
}

func TestExpandTabs(t *testing.T) {
	test := require.New(t)
	test.Equal("  L1\n    L2\n  L3a\tL3b", common.ExpandTabs("\tL1\n\t\tL2\n\tL3a\tL3b", 2))
}

func TestAddTrailingNewLine(t *testing.T) {
	test := require.New(t)
	test.Equal("", common.AddTrailingNewLine(""))
	test.Equal("A\n", common.AddTrailingNewLine("A"))
	test.Equal("A\nB\n", common.AddTrailingNewLine("A\nB"))
	test.Equal("A\nB\n", common.AddTrailingNewLine("A\nB\n"))
	test.Equal("A\nB\r", common.AddTrailingNewLine("A\nB\r"))
	test.Equal("A\nB\r\n", common.AddTrailingNewLine("A\nB\r\n"))
	test.Equal("A\nB\r\n\n", common.AddTrailingNewLine("A\nB\r\n\n"))
}
