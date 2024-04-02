package files

import (
	"io/fs"
	"sync"
	"sync/atomic"
	"time"
)

type fsEntry struct {
	statMutex  sync.Mutex
	statLoaded atomic.Bool
	statError  error

	cache *fsCache
	path  Path
	name  string

	size int64
	time time.Time
	mode fs.FileMode

	dataMutex  sync.Mutex
	dataLoaded atomic.Bool
	dataError  error

	data []byte
}

func fsEntryNew(cache *fsCache, path Path, name string) *fsEntry {
	return &fsEntry{cache: cache, path: path, name: name}
}

func (entry *fsEntry) Open() (file File, err error) {
	entry.ensureStatLoaded()

	err = entry.statError
	if err == nil {
		file = fsFileNew(entry)
	}
	return
}

func (entry *fsEntry) Info() (fs.FileInfo, error) {
	entry.ensureStatLoaded()
	return entry, entry.statError
}

func (entry *fsEntry) Name() string {
	return entry.name
}

func (entry *fsEntry) Size() int64 {
	if entry.dataLoaded.Load() && entry.dataError == nil {
		return int64(len(entry.data))
	}

	entry.ensureStatLoaded()
	return entry.size
}

func (entry *fsEntry) Type() fs.FileMode {
	return entry.Mode().Type()
}

func (entry *fsEntry) Mode() fs.FileMode {
	entry.ensureStatLoaded()
	return entry.mode
}

func (entry *fsEntry) ModTime() time.Time {
	return entry.time
}

func (entry *fsEntry) IsDir() bool {
	return entry.Mode().IsDir()
}

func (entry *fsEntry) Sys() any {
	return nil
}

func (entry *fsEntry) ReadFile() ([]byte, error) {
	if entry.dataLoaded.Load() {
		return entry.data, entry.dataError
	}

	entry.dataMutex.Lock()
	defer entry.dataMutex.Unlock()
	if !entry.dataLoaded.Load() {
		entry.data, entry.dataError = entry.cache.loader.ReadFile(entry.path)
		entry.dataLoaded.Store(true)
	}

	return entry.data, entry.dataError
}

func (entry *fsEntry) ensureStatLoaded() {
	if entry.statLoaded.Load() {
		return
	}

	entry.statMutex.Lock()
	defer entry.statMutex.Unlock()

	if entry.statLoaded.Load() {
		return
	}

	info, err := entry.cache.loader.Stat(entry.path)

	entry.size = info.Size()
	entry.time = info.ModTime()
	entry.mode = info.Mode()
	entry.statError = err
	entry.statLoaded.Store(true)

}
