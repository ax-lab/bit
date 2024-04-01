package files_test

import (
	"testing"

	"axlab.dev/test/files"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	test := require.New(t)

	// empty
	test.Equal(files.Path(""), files.PathNew[string]())
	test.Equal(files.Path(""), files.PathNew(""))
	test.Equal(files.Path(""), files.PathNew("", ""))

	// single component
	test.Equal(files.Path("."), files.PathNew("."))
	test.Equal(files.Path(".."), files.PathNew(".."))
	test.Equal(files.Path("a"), files.PathNew("a"))
	test.Equal(files.Path("abc"), files.PathNew("abc"))

	// rooted
	test.Equal(files.Path("/"), files.PathNew("/"))
	test.Equal(files.Path("/a/b"), files.PathNew("/a", "b"))
	test.Equal(files.Path("/"), files.PathNew("x", "y", "/"))
	test.Equal(files.Path("/a/b"), files.PathNew("x", "/", "y", "/", "a", "b"))

	// empty components
	test.Equal(files.Path("abc"), files.PathNew("", "abc", ""))
	test.Equal(files.Path("a/b/c"), files.PathNew("", "a", "", "b", "c", ""))

	// multiple slashes
	test.Equal(files.Path("a/b/c/d"), files.PathNew("a//b/c///d"))

	// trailing slash
	test.Equal(files.Path("a/b/c/d"), files.PathNew("a/", "b/c/", "d/"))

	// dot-dot
	test.Equal(files.Path("a/b/c"), files.PathNew("a", "b", "c", "d/.."))
	test.Equal(files.Path("../a"), files.PathNew("x", "../..", "a", "b", "c", "d", "../../.."))

	// rooted dot-dot
	test.Equal(files.Path("/a"), files.PathNew("/..", "a"))
	test.Equal(files.Path("/a"), files.PathNew("/../..", "a"))
	test.Equal(files.Path("/a"), files.PathNew("//..", "a"))

	// reverse slash
	test.Equal(files.Path("a/b/c"), files.PathNew("a\\b", "c"))
	test.Equal(files.Path("a/b/c"), files.PathNew("a\\\\b\\c", ""))
	test.Equal(files.Path("a/b/c"), files.PathNew("a\\b\\c", "d\\.."))
	test.Equal(files.Path("../a"), files.PathNew("x", "..\\..", "a\\b", "c\\d", "..\\..\\.."))
	test.Equal(files.Path("/"), files.PathNew("\\"))
	test.Equal(files.Path("/a/b"), files.PathNew("\\a", "b"))

	test.Equal(files.Path("/"), files.PathNew("x", "y", "\\"))
	test.Equal(files.Path("/a/b"), files.PathNew("x", "\\", "y", "\\", "a", "b"))
}

func TestPathIs(t *testing.T) {
	test := require.New(t)

	test.True(files.Path("/").IsRoot())
	test.True(files.Path("/a/b").IsRoot())

	test.False(files.Path("a/b").IsRoot())
	test.False(files.Path(".").IsRoot())
	test.False(files.Path("").IsRoot())

	test.True(files.Path("..").IsOutside())
	test.True(files.Path("../").IsOutside())
	test.True(files.Path("../abc").IsOutside())

	test.False(files.Path("/").IsOutside())
	test.False(files.Path(".").IsOutside())
	test.False(files.Path("./abc").IsOutside())
}

func TestPathClean(t *testing.T) {
	test := require.New(t)
	test.Equal(files.Path("a/b/c"), files.Path(".//a\\b\\x/.././//c/").Clean())
}
