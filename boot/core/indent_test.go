package core_test

import (
	"strings"
	"testing"

	"axlab.dev/bit/boot/core"
	"github.com/stretchr/testify/require"
)

func TestFormatWriter(t *testing.T) {
	const tabSize = 3

	test := require.New(t)

	checkTab := func(input, expected string) {
		output := strings.Builder{}
		writer := core.FormatWriterNew(&output)
		writer.UseSpaces(false)
		n, e := writer.WriteString(input)
		test.NoError(e)
		test.Equal(len(input), n)
		test.Equal(expected, output.String())
	}

	checkSpc := func(input, expected string) {
		output := strings.Builder{}
		writer := core.FormatWriterNew(&output)
		writer.SetTabSize(tabSize)
		writer.UseSpaces(true)
		n, e := writer.WriteString(input)
		test.NoError(e)
		test.Equal(len(input), n)
		test.Equal(expected, output.String())
	}

	check := func(input, expected string) {
		checkTab(input, expected)
		checkSpc(input, expected)
	}

	check("", "")
	check("abc", "abc")

	check("L1\nL2\nL3", "L1\nL2\nL3")
	check("L1\rL2\rL3", "L1\nL2\nL3")
	check("L1\r\nL2\r\nL3", "L1\nL2\nL3")

	check("L1\n", "L1\n")
	check("L1\r", "L1\n")
	check("L1\r\n", "L1\n")

	checkTab("L1\tL2\n\tL3", "L1\tL2\n\tL3")
	checkSpc("X\tY\tZ\r\nL1\tL2\n\tL3", "X  Y  Z\nL1 L2\n   L3")
}

func TestFormatWriterIndentTab(t *testing.T) {
	test := require.New(t)

	output := strings.Builder{}
	writer := core.FormatWriterNew(&output)

	out := func(str string) {
		n, e := writer.WriteString(str)
		test.NoError(e)
		test.Equal(len(str), n)
	}

	inc := func() { writer.Indent() }
	dec := func() { writer.Dedent() }

	inc()
	out("L1\nL2\n\tL3")
	dec()
	out("\nL4\n\tL5\n")
	inc()
	inc()
	out("L6\n\tL7")
	dec()
	out("\n\tL8\nL9\n")
	dec()
	out("L0\n")

	expected := []string{
		"\tL1",
		"\tL2",
		"\t\tL3",
		"L4",
		"\tL5",
		"\t\tL6",
		"\t\t\tL7",
		"\t\tL8",
		"\tL9",
		"L0",
		"",
	}

	actual := strings.Split(output.String(), "\n")
	test.Equal(expected, actual)
}

func TestFormatWriterIndentSpc(t *testing.T) {
	const tabSize = 3

	test := require.New(t)

	output := strings.Builder{}
	writer := core.FormatWriterNew(&output)
	writer.UseSpaces(true)
	writer.SetTabSize(tabSize)

	out := func(str string) {
		n, e := writer.WriteString(str)
		test.NoError(e)
		test.Equal(len(str), n)
	}

	inc := func() { writer.Indent() }
	dec := func() { writer.Dedent() }

	inc()
	out("L1\nL2\n\tL3")
	dec()
	out("\nL4\n\tL5\n")
	inc()
	inc()
	out("L6\n\tL7")
	dec()
	out("\n\tL8\nL9\n")
	dec()
	out("L0\n")

	expected := []string{
		"   L1",
		"   L2",
		"      L3",
		"L4",
		"   L5",
		"      L6",
		"         L7",
		"      L8",
		"   L9",
		"L0",
		"",
	}

	actual := strings.Split(output.String(), "\n")
	test.Equal(expected, actual)
}
