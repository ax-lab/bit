package files_test

import (
	"io"
	"io/fs"
	"testing"

	"axlab.dev/test/files"
	"github.com/stretchr/testify/require"
)

func TestFS(t *testing.T) {
	test := require.New(t)
	root, err := files.Root(".")
	test.NoError(err)

	var ok bool

	_, ok = root.(files.FS)
	test.True(ok, "must implement files.FS")

	_, ok = root.(fs.SubFS)
	test.True(ok, "must implement fs.SubFS")

	_, ok = root.(fs.StatFS)
	test.True(ok, "must implement fs.StatFS")

	_, ok = root.(fs.ReadDirFS)
	test.True(ok, "must implement fs.ReadDirFS")

	_, ok = root.(fs.ReadFileFS)
	test.True(ok, "must implement fs.ReadFileFS")

	_, ok = root.(fs.GlobFS)
	test.True(ok, "must implement fs.GlobFS")

	file, err := root.Open(".")
	test.NoError(err)

	_, ok = file.(io.Reader)
	test.True(ok, "file must implement io.Reader")

	_, ok = file.(io.ReaderAt)
	test.True(ok, "file must implement io.ReaderAt")

	_, ok = file.(io.Seeker)
	test.True(ok, "file must implement io.Seeker")

	_, ok = file.(fs.ReadDirFile)
	test.True(ok, "file must implement io.ReadDirFile")
}
