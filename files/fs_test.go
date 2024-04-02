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
