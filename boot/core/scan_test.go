package core_test

import (
	"testing"

	"axlab.dev/bit/boot/core"
	"github.com/stretchr/testify/require"
)

func TestScanRead(t *testing.T) {
	test := require.New(t)

	const INPUT = "abc123"
	scan := core.ScannerNew(INPUT)

	for _, expected := range INPUT {
		chr, read := scan.Read()
		test.True(read)
		test.Equal(expected, chr)
	}

	_, read := scan.Read()
	test.False(read)
}

func TestScanPeek(t *testing.T) {
	test := require.New(t)

	const INPUT = "abc123"
	scan := core.ScannerNew(INPUT)

	for _, expected := range INPUT {
		test.Equal(expected, scan.Peek())
		scan.Read()
	}
	test.Equal(rune(0), scan.Peek())
}

func TestScanText(t *testing.T) {
	test := require.New(t)

	const INPUT = "abc123"
	scan := core.ScannerNew(INPUT)

	for n := range INPUT {
		rest := INPUT[n:]
		test.Equal(rest, scan.Text())
		test.Equal(len(rest), scan.Len())
		scan.Read()
	}

	test.Equal("", scan.Text())
	test.Equal(0, scan.Len())
}

func TestScanReadChars(t *testing.T) {
	test := require.New(t)

	const INPUT = "a12abc1234"
	scan := core.ScannerNew(INPUT)

	test.Equal("", scan.ReadChars(0))
	test.Equal("a", scan.ReadChars(1))
	test.Equal("12", scan.ReadChars(2))
	test.Equal("abc", scan.ReadChars(3))
	test.Equal("1234", scan.ReadChars(4))
	test.Equal("", scan.ReadChars(999))
}

func TestScanReadWhile(t *testing.T) {
	test := require.New(t)

	const INPUT = "12345ABCDE"
	scan := core.ScannerNew(INPUT)

	mT := core.MatchRegex(`\d`)
	mF := core.MatchRegex(`[^\d]`)
	test.Equal("", scan.ReadWhile(mF))
	test.Equal("12345", scan.ReadWhile(mT))
	test.Equal("ABCDE", scan.ReadWhile(mF))
}

func TestScanReadUntil(t *testing.T) {
	test := require.New(t)

	const INPUT = "12345ABCDE"
	scan := core.ScannerNew(INPUT)

	mT := core.MatchRegex(`\d`)
	mF := core.MatchRegex(`[^\d]`)
	test.Equal("", scan.ReadUntil(mT))
	test.Equal("12345", scan.ReadUntil(mF))
	test.Equal("ABCDE", scan.ReadUntil(mT))
}

func TestScanReadMatch(t *testing.T) {
	test := require.New(t)

	const INPUT = "42a123b456c"
	scan := core.ScannerNew(INPUT)

	var pre, txt string
	m := core.MatchRegex(`\d+`)

	pre, txt = scan.ReadMatch(m)
	test.Equal("", pre)
	test.Equal("42", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("a", pre)
	test.Equal("123", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("b", pre)
	test.Equal("456", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("c", pre)
	test.Equal("", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("", pre)
	test.Equal("", txt)
}

func TestScanPos(t *testing.T) {
	test := require.New(t)

	const INPUT = "123456"

	scan := core.ScannerNew(INPUT)
	read := func() (rune, int) {
		chr, ok := scan.Read()
		test.True(ok)
		return chr, scan.Pos()
	}

	test.Equal(0, scan.Pos())
	a0, p0 := read()
	a1, p1 := read()
	a2, p2 := read()
	a3, p3 := read()
	a4, p4 := read()
	a5, p5 := read()

	test.Equal('1', a0)
	test.Equal('2', a1)
	test.Equal('3', a2)
	test.Equal('4', a3)
	test.Equal('5', a4)
	test.Equal('6', a5)

	test.Equal(1, p0)
	test.Equal(2, p1)
	test.Equal(3, p2)
	test.Equal(4, p3)
	test.Equal(5, p4)
	test.Equal(6, p5)
}
