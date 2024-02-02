package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"axlab.dev/bit/logs"
)

type Entry struct {
	Name  string
	Path  string
	IsDir bool

	fullPath string
	entry    fs.DirEntry
	file     fs.FileInfo
	fileErr  error
}

func (entry *Entry) Info() (fs.FileInfo, error) {
	if entry.file == nil && entry.fileErr == nil {
		entry.file, entry.fileErr = entry.entry.Info()
	}
	return entry.file, entry.fileErr
}

func (entry *Entry) FullPath() string {
	return entry.fullPath
}

func (entry *Entry) ModTime() time.Time {
	if file, err := entry.Info(); err == nil {
		return file.ModTime()
	} else {
		return time.Time{}
	}
}

type ListOptions struct {
	Hidden bool
	Filter func(entry *Entry) bool
}

func (info *Entry) String() string {
	mode := "F"
	if info.IsDir {
		mode = "D"
	}
	return fmt.Sprintf("%s %s", mode, info.Path)
}

func List(dirPath string, options ListOptions) (out []*Entry) {
	dir := os.DirFS(dirPath)
	basePath := logs.Handle(filepath.Abs(dirPath))
	err := fs.WalkDir(dir, ".", func(entryPath string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			logs.Warn(err, "listing `%s`", dirPath)
			return nil
		}

		name := dirEntry.Name()
		if name == "." {
			return nil
		}

		entry := &Entry{
			Name:  dirEntry.Name(),
			Path:  path.Join(dirPath, entryPath),
			IsDir: dirEntry.IsDir(),
			entry: dirEntry,
		}
		entry.fullPath = filepath.Join(basePath, entryPath)

		skip := !options.Hidden && strings.HasPrefix(name, ".")
		skip = skip || options.Filter != nil && !options.Filter(entry)

		if !skip {
			if _, err := entry.Info(); err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					logs.Warn(err, "listing `%s`", dirPath)
				}
				skip = true
			}
		}

		if skip {
			if entry.IsDir {
				return fs.SkipDir
			} else {
				return nil
			}
		}

		out = append(out, entry)
		return nil
	})
	logs.Check(err)
	return out
}
