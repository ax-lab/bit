package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"axlab.dev/bit/logs"
)

type Dir struct {
	name string
	path string
	fs   fs.FS
}

func OpenDir(path string) Dir {
	fullPath := logs.Handle(filepath.Abs(path))
	return Dir{
		name: path,
		path: fullPath,
		fs:   os.DirFS(fullPath),
	}
}

func (dir Dir) MustExist(label string) Dir {
	if !IsDir(dir.FullPath()) {
		logs.Fatal("%s is not a valid directory: %s", label, dir.name)
	}
	return dir
}

func (dir Dir) Create(label string) Dir {
	if err := os.MkdirAll(dir.FullPath(), fs.ModePerm); err != nil {
		logs.Fatal("%s directory `%s` could not be created: %v", label, dir.name, err)
	}
	return dir
}

func (dir Dir) Name() string {
	return dir.name
}

func (dir Dir) FullPath() string {
	return dir.path
}

func (dir Dir) Write(name, text string) *DirFile {
	path, name := dir.ResolvePath(name)
	logs.Check(os.MkdirAll(filepath.Dir(path), os.ModePerm))
	logs.Check(os.WriteFile(path, []byte(text), os.ModePerm))
	return &DirFile{name: name, path: path, text: text}
}

func (dir Dir) MakeDir(name string) {
	path, _ := dir.ResolvePath(name)
	logs.Check(os.MkdirAll(path, os.ModePerm))
}

func (dir Dir) RemoveAll(name string) {
	path, name := dir.ResolvePath(name)
	if err := os.RemoveAll(path); err != nil {
		logs.Warn(err, "removing `%s` from `%s`", name, dir.name)
	}
}

func (dir Dir) ReadFile(name string) *DirFile {
	path, name := dir.ResolvePath(name)
	if data, err := os.ReadFile(path); err == nil {
		text := string(data)
		return &DirFile{name: name, path: path, text: text}
	} else if !errors.Is(err, fs.ErrNotExist) {
		logs.Warn(err, "reading `%s` in `%s`", name, dir.name)
	}
	return nil
}

func (dir Dir) GetFullPath(path string) string {
	path, _ = dir.ResolvePath(path)
	return path
}

func (dir Dir) ResolvePath(path string) (fullPath, relativeName string) {
	base := dir.FullPath()
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(base, path)
	} else {
		fullPath = filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(fullPath); err == nil {
		fullPath = resolved
	} else if !errors.Is(err, fs.ErrNotExist) {
		logs.Check(err)
	}

	filePath := logs.Handle(filepath.Rel(base, fullPath))
	if filePath == "" || strings.Contains(filePath, "..") {
		logs.Fatal("`%s` is not a valid path within directory `%s`", path, dir.Name())
	}

	filePath = strings.Replace(filePath, "\\", "/", -1)
	return fullPath, filePath
}

type DirFile struct {
	name string
	path string
	text string
}

func (file *DirFile) FullPath() string {
	return file.path
}

func (file *DirFile) Name() string {
	return file.name
}

func (file *DirFile) Text() string {
	return file.text
}