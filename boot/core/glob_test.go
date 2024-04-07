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

	test.True(core.GlobMatch("abc0.txt", "abc[0-9]+.txt"))
	test.True(core.GlobMatch("abc99.txt", "abc[0-9]+.txt"))
	test.True(core.GlobMatch("abc1234567890.txt", "abc[0-9]+.txt"))
	test.False(core.GlobMatch("abc.txt", "abc[0-9]+.txt"))
}

func TestGlobParsePrefix(t *testing.T) {
	test := require.New(t)

	prefix := func(pattern string) string {
		pre, _ := core.GlobParse(pattern)
		return pre
	}

	test.Equal("", prefix(""))
	test.Equal("abc", prefix("abc"))
	test.Equal("abc.def", prefix("abc.def"))
	test.Equal("path/abc.def", prefix("path/abc.def"))
	test.Equal("path/sub/file.ext", prefix("path/sub/file.ext"))

	test.Equal("path/", prefix("path/*/file.ext"))
	test.Equal("path/", prefix("path/?/file.ext"))

	test.Equal("path/", prefix("path/abc*/file.ext"))
	test.Equal("path/", prefix("path/abc?/file.ext"))

	test.Equal("path/*/", prefix("path/\\*/[abc]file.ext"))
	test.Equal("path/?/", prefix("path/\\?/[abc]file.ext"))
}
