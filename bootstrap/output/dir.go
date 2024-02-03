package output

import (
	"os"
	"path/filepath"

	"axlab.dev/bit/logs"
)

type Dir struct {
	root string
}

func Open(rootDir string) *Dir {
	root := logs.Handle(filepath.Abs(rootDir))
	logs.Check(os.MkdirAll(rootDir, os.ModePerm))
	return &Dir{root: root}
}

func (dir *Dir) Root() string {
	return dir.root
}
