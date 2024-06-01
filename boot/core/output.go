package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type OutputMap struct {
	BaseDir string

	sync sync.Mutex
	sets map[string]*OutputSet
}

func (out *OutputMap) Get(name string) *OutputSet {
	if name != "" {
		name = Try(CleanPath(name))
	}

	out.sync.Lock()
	defer out.sync.Unlock()
	if out.sets == nil {
		out.sets = make(map[string]*OutputSet)
	}

	set := out.sets[name]
	if set == nil {
		set = &OutputSet{out: out, name: name}
		out.sets[name] = set
	}
	return set
}

func (out *OutputMap) WriteOutput() error {
	out.sync.Lock()
	defer out.sync.Unlock()

	for _, set := range out.sets {
		if err := set.WriteOutput(); err != nil {
			return err
		}
	}

	return nil
}

type OutputSet struct {
	sync  sync.Mutex
	out   *OutputMap
	name  string
	files map[string]string
}

func (set *OutputSet) Add(path, text string) {
	path, err := CleanPath(path)
	NoError(err, "output file path")

	set.sync.Lock()
	defer set.sync.Unlock()
	if _, dup := set.files[path]; dup {
		panic(fmt.Sprintf("OutputSet: duplicated file: %s", path))
	}

	if set.files == nil {
		set.files = make(map[string]string)
	}
	set.files[path] = text
}

func (set *OutputSet) GetFullPath(name string) (string, error) {
	path, err := CleanPath(name)
	if err != nil {
		return "", err
	}

	base := set.out.BaseDir
	if base == "" {
		base = "."
	}

	full, err := filepath.Abs(filepath.Join(base, set.name, path))
	return full, err
}

func (set *OutputSet) WriteOutput() error {
	set.sync.Lock()
	defer set.sync.Unlock()

	baseDir := set.out.BaseDir
	if baseDir == "" {
		baseDir = "."
	}

	name := set.name
	if name == "" {
		name = "(root)"
	}

	setDir := path.Join(baseDir, set.name)
	if err := os.MkdirAll(setDir, os.ModePerm); err != nil {
		err = fmt.Errorf("could not create `%s` output dir: %v (base=%s)", name, err, baseDir)
		return err
	}

	if err := set.writeTo(setDir); err != nil {
		err = fmt.Errorf("outputting `%s`: %v (base=%s)", name, err, baseDir)
		return err
	}

	return nil
}

func (set *OutputSet) writeTo(basePath string) error {
	hasDir := make(map[string]bool)
	for name, text := range set.files {
		full := path.Join(basePath, name)
		if data, err := os.ReadFile(full); err == nil {
			if len(data) == len(text) && string(data) == text {
				continue
			}
		}

		dirName := path.Dir(name)
		if dirName != "" && dirName != "." && !hasDir[dirName] {
			fullDir := path.Join(basePath, dirName)
			if err := os.MkdirAll(fullDir, os.ModePerm); err != nil {
				err = fmt.Errorf("could not create dir `%s/%s`: %v", set.name, dirName, err)
				return err
			}
			hasDir[dirName] = true
		}

		if err := os.WriteFile(full, ([]byte)(text), os.ModePerm); err != nil {
			err = fmt.Errorf("could not write `%s/%s`: %v", set.name, name, err)
			return err
		}
	}

	return nil
}
