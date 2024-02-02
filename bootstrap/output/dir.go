package output

import (
	"os"
	"path/filepath"

	"axlab.dev/bit/errs"
)

type Dir struct {
	root string
}

func Open(rootDir string) *Dir {
	root := errs.Handle(filepath.Abs(rootDir))
	errs.Check(os.MkdirAll(rootDir, os.ModePerm))
	return &Dir{root: root}
}

func (dir *Dir) Root() string {
	return dir.root
}

func (dir *Dir) Write(name, text string) {
	path := filepath.Join(dir.root, name)
	errs.Check(os.MkdirAll(filepath.Dir(path), os.ModePerm))
	errs.Check(os.WriteFile(path, []byte(text), os.ModePerm))
}
