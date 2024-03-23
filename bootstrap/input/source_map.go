package input

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type SourceMap struct {
	mutex sync.Mutex
	index atomic.Uint64
	files map[string]*sourceData
}

func (srcMap *SourceMap) NewSource(name, text string) Source {
	srcData := &sourceData{
		index:  srcMap.index.Add(1),
		parent: srcMap,
		name:   name,
		text:   text,
	}
	return Source{srcData}
}

func (srcMap *SourceMap) LoadFile(file string) (Source, error) {
	srcMap.mutex.Lock()
	defer srcMap.mutex.Unlock()

	fullPath, err := filepath.Abs(file)
	if err != nil {
		return Source{}, err
	}

	if srcData, loaded := srcMap.files[fullPath]; loaded {
		if srcData.err != nil {
			return Source{}, srcData.err
		}
		return Source{srcData}, nil
	}

	srcData := &sourceData{
		index:  srcMap.index.Add(1),
		parent: srcMap,
		name:   file,
	}
	if srcMap.files == nil {
		srcMap.files = make(map[string]*sourceData)
	}
	srcMap.files[fullPath] = srcData

	bytes, err := os.ReadFile(file)
	if err != nil {
		srcData.err = err
		return Source{}, err
	}

	srcData.text = string(bytes)
	return Source{srcData}, nil
}
