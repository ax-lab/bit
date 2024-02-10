package tester_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"axlab.dev/bit/tester"
	"github.com/stretchr/testify/require"
)

func TestTempDir(t *testing.T) {
	test := require.New(t)

	dir, err := tester.TryMakeDir("tester-dir", map[string]string{
		"a/a1.txt":       "this is A1",
		"a/a2.txt":       "this is A2",
		"a/a3.txt":       "this is A3",
		"a/sub/some.txt": "some file under A",
		"b/b1.txt":       "this is B1",
		"b/b2.txt":       "this is B2",
		"some/path/with/multiple/levels/file.txt": "deeply nested file",
		"text.txt": `
			Line 1
			Line 2
				Line 3
				Line 4
			Line 5
				Line 6
		`,
	})

	test.NoError(err)
	test.DirExists(dir.DirPath())
	test.Contains(dir.DirPath(), "tester-dir")
	test.True(strings.HasPrefix(dir.DirPath(), os.TempDir()))

	check := func(name, text string) {
		path := filepath.Join(dir.DirPath(), name)
		test.FileExists(path)
		data, err := os.ReadFile(path)
		test.NoError(err)

		fileText := string(data)
		test.Equal(text, fileText)
	}

	check("a/a1.txt", "this is A1\n")
	check("a/a2.txt", "this is A2\n")
	check("a/a3.txt", "this is A3\n")
	check("a/sub/some.txt", "some file under A\n")
	check("b/b1.txt", "this is B1\n")
	check("b/b2.txt", "this is B2\n")
	check("some/path/with/multiple/levels/file.txt", "deeply nested file\n")
	check("text.txt", "Line 1\nLine 2\n\tLine 3\n\tLine 4\nLine 5\n\tLine 6\n")

	dir.Delete()
	test.NoDirExists(dir.DirPath())
}
