package core

import (
	"cmp"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Source struct {
	data *sourceData
}

func (src Source) Valid() bool {
	return src.data != nil && src.data.name != "" && src.data.err == nil
}

func (src Source) Name() string {
	if src.data == nil {
		return ""
	}
	return src.data.name
}

func (src Source) Text() string {
	src.checkValid()
	return src.data.text
}

func (src Source) TabSize() int {
	src.checkValid()
	tabs := src.data.tabs.Load()
	if tabs <= 0 {
		return int(DefaultTabSize)
	}
	return int(tabs)
}

func (src Source) SetTabSize(size int) {
	src.checkValid()
	if size <= 0 || size > 32 {
		panic("invalid tab size")
	}
	src.data.tabs.Store(int32(size))
}

func (src Source) Compare(other Source) int {
	va, vb := src.Valid(), other.Valid()
	if va != vb {
		if !va {
			return -1
		} else {
			return +1
		}
	}

	if res := cmp.Compare(src.Name(), other.Name()); res != 0 {
		return res
	}

	if res := cmp.Compare(len(src.Text()), len(other.Text())); res != 0 {
		return res
	}

	return cmp.Compare(uintptr(unsafe.Pointer(src.data)), uintptr(unsafe.Pointer(other.data)))
}

func (src Source) checkValid() {
	if !src.Valid() {
		panic("invalid source")
	}
}

type sourceData struct {
	err  error
	name string
	text string
	tabs atomic.Int32
}

type SourceMap struct {
	rootDir DirEntry

	sourceSync sync.Mutex
	sourceMap  map[string]*sourceData
}

func SourceMapNew(rootDir DirEntry) SourceMap {
	return SourceMap{
		rootDir: rootDir,
	}
}

func (src *SourceMap) LoadFile(path string) (Source, error) {
	name, err := src.rootDir.Resolve(path)
	if err != nil {
		return Source{}, fmt.Errorf("loading source file: %v", err)
	}

	src.sourceSync.Lock()
	defer src.sourceSync.Unlock()

	if entry := src.sourceMap[name]; entry != nil {
		return Source{entry}, entry.err
	}

	entry := &sourceData{
		name: name,
	}

	if src.sourceMap == nil {
		src.sourceMap = make(map[string]*sourceData)
	}
	src.sourceMap[name] = entry

	file, err := src.rootDir.Get(name)
	if err != nil {
		entry.err = err
	} else {
		if file.Info().IsDir() {
			entry.err = fmt.Errorf("loading source file: `%s` is a directory", name)
		} else {
			text, err := file.Read()
			if err == nil {
				entry.text = string(text)
			} else {
				entry.err = fmt.Errorf("loading source file: could not read `%s` -- %v", name, err)
			}
		}
	}

	return Source{entry}, entry.err
}

func (src *SourceMap) LoadString(name, text string) Source {
	if name == "" {
		panic("source name cannot be empty")
	}

	data := &sourceData{
		name: name,
		text: text,
	}
	return Source{data}
}
