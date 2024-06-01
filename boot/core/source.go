package core

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type Source interface {
	Name() string
	Text() string
	Span() Span
	String() string
	TabSize() int
}

func SourceCompare(a, b Source) int {
	if a == nil && b == nil {
		return 0
	} else if a == nil {
		return -1
	} else if b == nil {
		return +1
	}
	res := cmp.Compare(a.Name(), b.Name())
	return res
}

type SourceLoader struct {
	sync sync.Mutex

	baseDir      string
	baseDirInit  bool
	baseDirPath  string
	baseDirError error

	sources map[string]sourceEntry
}

func (loader *SourceLoader) SetBaseDir(baseDir string) error {
	loader.sync.Lock()
	defer loader.sync.Unlock()
	if loader.baseDirInit {
		panic("SourceLoader: cannot change base directory after loading files")
	}

	loader.baseDir = baseDir
	_, err := loader.getBaseDir()
	return err
}

func (loader *SourceLoader) Preload(name, text string) Source {
	nameKey, err := loader.cleanName(name)
	if err != nil {
		panic(fmt.Sprintf("SourceLoader: preloading invalid source name: %v", err))
	}

	loader.sync.Lock()
	defer loader.sync.Unlock()
	if _, hasEntry := loader.sources[nameKey]; hasEntry {
		panic(fmt.Sprintf("SourceLoader: preloading source already loaded: %s", nameKey))
	}

	src := &source{
		name: nameKey,
		text: text,
	}

	if loader.sources == nil {
		loader.sources = make(map[string]sourceEntry)
	}
	loader.sources[nameKey] = sourceEntry{
		source: src,
	}

	return src
}

func (loader *SourceLoader) Load(name string) (Source, error) {
	nameKey, err := loader.cleanName(name)
	if err != nil {
		return nil, err
	}

	loader.sync.Lock()
	defer loader.sync.Unlock()
	if entry, hasEntry := loader.sources[nameKey]; hasEntry {
		return entry.source, entry.err
	}

	if loader.sources == nil {
		loader.sources = make(map[string]sourceEntry)
	}

	baseDir, err := loader.getBaseDir()
	if err != nil {
		return nil, err
	}

	nameFull := filepath.Join(baseDir, nameKey)
	data, err := os.ReadFile(nameFull)

	var entry sourceEntry
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			entry.err = fmt.Errorf("source file `%s` not found", nameKey)
		} else if stat, statErr := os.Stat(nameFull); statErr == nil && stat.IsDir() {
			entry.err = fmt.Errorf("source file `%s` is a directory", nameKey)
		} else {
			entry.err = fmt.Errorf("source file `%s`: %v", nameKey, err)
		}
	}

	if entry.err == nil {
		entry.source = &source{
			name:   nameKey,
			text:   string(data),
			loader: loader,
		}
	}
	loader.sources[nameKey] = entry

	return entry.source, entry.err
}

func (loader *SourceLoader) ResolveName(base, name string) (out string, err error) {
	base, err = loader.cleanName(base)
	if err != nil {
		panic(fmt.Sprintf("invalid base name: %v", err))
	}

	name, err = loader.cleanName(name)
	if err != nil {
		return "", err
	}

	fullName := path.Join(base, name)
	if clean, err := loader.cleanName(fullName); err != nil {
		panic(fmt.Sprintf("invalid resolved name: %v", err))
	} else if clean != fullName {
		panic(fmt.Sprintf("resolved name was not clean: %#v -> %#v", fullName, clean))
	}

	return fullName, nil
}

func (loader *SourceLoader) cleanName(name string) (string, error) {
	out, err := CleanPath(name)
	if err != nil {
		err = fmt.Errorf("source name: %v", err)
	}
	return out, err
}

type sourceEntry struct {
	source Source
	err    error
}

type source struct {
	name   string
	text   string
	loader *SourceLoader
}

func (src *source) Name() string {
	return src.name
}

func (src *source) Text() string {
	return src.text
}

func (src *source) Span() Span {
	return spanForSource(src)
}

func (src *source) String() string {
	return fmt.Sprintf("Source(%s)", src.name)
}

func (src *source) Loader() *SourceLoader {
	return src.loader
}

func (src *source) TabSize() int {
	return DefaultTabSize
}

func (loader *SourceLoader) getBaseDir() (string, error) {
	if loader.baseDirInit {
		return loader.baseDir, loader.baseDirError
	}

	baseDir := loader.baseDir
	if baseDir == "" {
		baseDir = "."
	}

	fullPath, err := filepath.Abs(baseDir)
	if err != nil {
		err = fmt.Errorf("source `%s`: path error (%v)", baseDir, err)
	} else if stat, statErr := os.Stat(fullPath); statErr != nil {
		err = fmt.Errorf("source `%s`: stat error (%v)", baseDir, statErr)
	} else if !stat.IsDir() {
		err = fmt.Errorf("source `%s`: not a directory", baseDir)
	}

	loader.baseDirPath = fullPath
	loader.baseDirError = err
	loader.baseDirInit = true
	return fullPath, err
}
