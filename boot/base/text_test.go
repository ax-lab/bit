package base_test

import (
	"testing"

	"axlab.dev/bit/base"
	"github.com/stretchr/testify/require"
)

func TestLines(t *testing.T) {
	test := require.New(t)

	var (
		actual   []string
		expected []string
	)

	expected = []string{"L1", "L2", "L3", "L4"}

	actual = base.Lines("L1\nL2\nL3\nL4")
	test.Equal(expected, actual)

	actual = base.Lines("L1\rL2\rL3\rL4")
	test.Equal(expected, actual)

	actual = base.Lines("L1\r\nL2\r\nL3\r\nL4")
	test.Equal(expected, actual)

	expected = append(expected, "")
	actual = base.Lines("L1\nL2\nL3\nL4\n")

	test.Equal(expected, actual)
}

func TestIndent(t *testing.T) {
	test := require.New(t)

	var (
		actual string
	)

	actual = base.Indent("L1\nL2")
	test.Equal("L1\n\tL2", actual)

	actual = base.Indent("L1\nL2", "  ")
	test.Equal("L1\n  L2", actual)

	actual = base.Indent("L1\nL2", ".. ")
	test.Equal(".. L1\n.. L2", actual)

	actual = base.Indent("L1\nL2\n\tL3\n\tL4\n\nL5\n")
	test.Equal("L1\n\tL2\n\t\tL3\n\t\tL4\n\n\tL5\n", actual)
}

func TestText(t *testing.T) {
	test := require.New(t)

	actual := base.Text("\r\tL1\r\t\tL2\r\tL3")
	test.Equal("L1\n\tL2\nL3\n", actual)
}
