package code_test

import (
	"testing"

	"axlab.dev/bit/code"
	"github.com/stretchr/testify/require"
)

func TestEncodeIdentifier(t *testing.T) {
	test := require.New(t)
	test.Equal("_$", code.EncodeIdentifier(""))
	test.Equal("_$_", code.EncodeIdentifier("_"))
	test.Equal("_$0", code.EncodeIdentifier("0"))
	test.Equal("_$9", code.EncodeIdentifier("9"))

	test.Equal("_$if", code.EncodeIdentifier("if"))
	test.Equal("_$case", code.EncodeIdentifier("case"))
	test.Equal("_$char", code.EncodeIdentifier("char"))

	test.Equal("abc", code.EncodeIdentifier("abc"))
	test.Equal("_123", code.EncodeIdentifier("_123"))
	test.Equal("a123", code.EncodeIdentifier("a123"))
	test.Equal("abc_$123", code.EncodeIdentifier("abc-123"))
	test.Equal("abc_$_$123", code.EncodeIdentifier("abc--123"))
	test.Equal("_$u4E16__$u754C__$u4E00__$u306E__$u8A00__$u8A9E_", code.EncodeIdentifier("世界一の言語"))
	test.Equal("x_$u4E16__$u754C__$u4E00__$u306E__$u8A00__$u8A9E_y", code.EncodeIdentifier("x世界一の言語y"))
}

func TestNameMap(t *testing.T) {
	test := require.New(t)
	names := &code.NameMap{}
	test.True(names.DeclareGlobal("x"))
	test.True(names.DeclareGlobal("y"))
	test.False(names.DeclareGlobal("y"))

	a := names.NewChild()
	test.Equal("a1", a.DeclareUnique("a1"))
	test.Equal("a2", a.DeclareUnique("a2"))
	test.Equal("a1_$$1", a.DeclareUnique("a1"))
	test.Equal("a2_$$1", a.DeclareUnique("a2"))
	test.Equal("a2_$$2", a.DeclareUnique("a2"))
	test.Equal("a2_$$3", a.DeclareUnique("a2"))

	b := names.NewChild()
	test.Equal("a1", b.DeclareUnique("a1"))
	test.Equal("a2", b.DeclareUnique("a2"))
	test.Equal("x_$$1", b.DeclareUnique("x"))
	test.Equal("x_$$2", b.DeclareUnique("x"))
	test.Equal("y_$$1", b.DeclareUnique("y"))
}
