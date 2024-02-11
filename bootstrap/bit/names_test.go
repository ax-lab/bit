package bit_test

import (
	"testing"

	"axlab.dev/bit/bit"
	"github.com/stretchr/testify/require"
)

func TestEncodeIdentifier(t *testing.T) {
	test := require.New(t)
	test.Equal("_$", bit.EncodeIdentifier(""))
	test.Equal("_$0", bit.EncodeIdentifier("0"))
	test.Equal("_$9", bit.EncodeIdentifier("9"))

	test.Equal("abc", bit.EncodeIdentifier("abc"))
	test.Equal("_123", bit.EncodeIdentifier("_123"))
	test.Equal("a123", bit.EncodeIdentifier("a123"))
	test.Equal("abc_$123", bit.EncodeIdentifier("abc-123"))
	test.Equal("abc_$_$123", bit.EncodeIdentifier("abc--123"))
	test.Equal("_$u4E16__$u754C__$u4E00__$u306E__$u8A00__$u8A9E_", bit.EncodeIdentifier("世界一の言語"))
	test.Equal("x_$u4E16__$u754C__$u4E00__$u306E__$u8A00__$u8A9E_y", bit.EncodeIdentifier("x世界一の言語y"))
}

func TestNameMap(t *testing.T) {
	test := require.New(t)
	names := &bit.NameMap{}
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
