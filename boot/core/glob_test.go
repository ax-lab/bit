package core_test

import (
	"testing"

	"axlab.dev/bit/boot/core"
	"github.com/stretchr/testify/require"
)

func TestGlob(t *testing.T) {
	test := require.New(t)

	test.True(core.GlobMatch("abc.txt", "abc.txt"))
	test.True(core.GlobMatch("ABC.txt", "abc.txt"))
	test.False(core.GlobMatch("123.txt", "abc.txt"))

	test.True(core.GlobMatch("abc1.txt", "abc?.txt"))
	test.True(core.GlobMatch("abc2.txt", "abc?.txt"))
	test.True(core.GlobMatch("abcX.txt", "abc?.txt"))
	test.False(core.GlobMatch("abc.txt", "abc?.txt"))

	test.True(core.GlobMatch("abc123.txt", "abc*.txt"))
	test.True(core.GlobMatch("abcXYZ.txt", "abc*.txt"))
	test.True(core.GlobMatch("abc.txt", "abc*.txt"))

	test.True(core.GlobMatch("p1/p2/name.txt", "*/*/*.txt"))
	test.True(core.GlobMatch("pre/p1/p2/name.txt", "*/*/*.txt"))
	test.False(core.GlobMatch("only/name.txt", "*/*/*.txt"))

	test.True(core.GlobMatch("name.txt", "**.txt"))
	test.True(core.GlobMatch("a/name.txt", "**/*.txt"))
	test.True(core.GlobMatch("a/b/name.txt", "**/*.txt"))
	test.False(core.GlobMatch("name.txt", "**/*.txt"))

	test.True(core.GlobMatch("p1\\p2\\name.txt", "*/*/*.txt"))
	test.True(core.GlobMatch("pre\\p1\\p2\\name.txt", "*/*/*.txt"))
}
