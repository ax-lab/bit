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

func (dir *Dir) Write(name, text string) {
	path := filepath.Join(dir.root, name)
	logs.Check(os.MkdirAll(filepath.Dir(path), os.ModePerm))
	logs.Check(os.WriteFile(path, []byte(text), os.ModePerm))
}
