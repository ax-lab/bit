package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type DirEntry struct {
	data *dirEntryData
}

type dirEntryData struct {
	root *dirEntryData
	name string
	path string
	info fs.FileInfo
}

func Dir(dirPath string) (dir DirEntry, err error) {
	fullPath, err := filepath.Abs(dirPath)
	if err != nil {
		err = fmt.Errorf("opening dir `%s`: path error -- %v", dirPath, err)
		return dir, err
	}

	info, err := os.Stat(fullPath)
	if err == nil {
		if !info.IsDir() {
			err = fmt.Errorf("opening dir `%s`: not a directory", dirPath)
		}
	} else {
		if errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("opening dir `%s`: directory not found", dirPath)
		} else {
			err = fmt.Errorf("opening dir `%s`: stat error -- %v", dirPath, err)
		}
	}

	data := &dirEntryData{
		name: "",
		path: fullPath,
		info: info,
	}
	data.root = data

	return DirEntry{data}, err
}

func (dir DirEntry) Valid() bool {
	return dir.data != nil && dir.data.path != "" && dir.data.info != nil
}

func (dir DirEntry) Resolve(name string) (string, error) {
	dir.checkEntry()
	entryName, _, err := dir.resolveEntry(name)
	return entryName, err
}

func (dir DirEntry) Get(name string) (DirEntry, error) {
	dir.checkEntry()
	entryName, entryPath, err := dir.resolveEntry(name)
	if err != nil {
		return DirEntry{}, err
	}

	data := &dirEntryData{
		root: dir.data.root,
		name: entryName,
		path: entryPath,
	}

	info, err := os.Stat(entryPath)
	data.info = info
	if err != nil {
		err = fmt.Errorf("could not read `%s`: stat error -- %v", entryName, err)
	}
	return DirEntry{data}, err
}

func (dir DirEntry) Name() string {
	if dir.data == nil {
		return ""
	}
	return dir.data.name
}

func (dir DirEntry) Info() os.FileInfo {
	dir.checkEntry()
	return dir.data.info
}

func (dir DirEntry) Read() ([]byte, error) {
	dir.checkEntry()
	return os.ReadFile(dir.data.path)
}

func (dir DirEntry) checkEntry() {
	if !dir.Valid() {
		panic("invalid directory entry")
	}
}

func (dir DirEntry) resolveEntry(name string) (entryName, entryPath string, err error) {
	dir.checkEntry()
	entryName, entryPath, valid := fsPathJoin(dir.data.root.path, dir.data.name, name)
	if !valid {
		if dir.data.name != "" {
			err = fmt.Errorf("invalid path from `%s`: %s", dir.data.name, name)
		} else {
			err = fmt.Errorf("invalid path: %s", name)
		}
	}
	return
}
