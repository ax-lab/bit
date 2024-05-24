package core

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Source interface {
	Name() string
	Text() string
	String() string
	Loader() *SourceLoader
}

type SourceLoader struct {
	baseDir string
	sync    sync.Mutex
	sources map[string]sourceEntry
}

func SourceLoaderNew(baseDir string) (*SourceLoader, error) {
	fullPath, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("source `%s`: path error (%v)", baseDir, err)
	}

	if stat, err := os.Stat(fullPath); err != nil {
		return nil, fmt.Errorf("source `%s`: stat error (%v)", baseDir, err)
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("source `%s`: not a directory", baseDir)
	}

	loader := &SourceLoader{
		baseDir: fullPath,
	}
	return loader, nil
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

	nameFull := filepath.Join(loader.baseDir, nameKey)
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
	out := path.Clean(name)

	valid := true
	if out == "" || strings.ContainsAny(out, "\r\n\t\\*?:|") {
		valid = false
	} else {
		for _, it := range strings.Split(out, "/") {
			if it == "." || it == ".." {
				valid = false
				break
			}
		}
	}

	if !valid {
		return "", fmt.Errorf("source name is not valid: %#v", name)
	} else if strings.HasPrefix(out, "/") {
		return "", fmt.Errorf("source name must be relative: %s", name)
	}

	return out, nil
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

func (src *source) String() string {
	return fmt.Sprintf("Source(%s)", src.name)
}

func (src *source) Loader() *SourceLoader {
	return src.loader
}
