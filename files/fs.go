package files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type FS interface {
	fs.FS

	Resolve(name string) (Path, error)
}

type fsRoot struct {
	cache *fsCache
	path  Path
	root  Path
}

func Root(baseDir string) (fs.FS, error) {
	fullPath, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("opening root `%s`: %v", baseDir, err)
	}

	fs := &fsLoaderFromRootDir{fullPath}
	return fsRoot{&fsCache{loader: fs}, "", ""}, nil
}

func (root fsRoot) Open(name string) (fs.File, error) {
	path, err := root.doResolve(name, "open file")
	if err != nil {
		return nil, err
	}

	entry := root.cache.Get(path)
	file, err := entry.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open file `%s`: %v", name, err)
	}
	return file, nil
}

func (root fsRoot) Sub(dir string) (fs.FS, error) {
	file, err := root.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("could not open sub dir `%s`: %v", dir, err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not open sub dir `%s`: %v", dir, err)
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("cannot open sub dir `%s`: not a directory", dir)
	}

	entry := file.(*fsFile).entry
	out := fsRoot{cache: root.cache, path: entry.path, root: entry.path}
	return out, nil
}

func (root fsRoot) Stat(name string) (fs.FileInfo, error) {
	path, err := root.doResolve(name, "stat")
	if err != nil {
		return nil, err
	}

	entry := root.cache.Get(path)
	stat, err := entry.Info()
	if err != nil {
		return nil, fmt.Errorf("could not stat file `%s`: %v", name, err)
	}
	return stat, nil
}

func (root fsRoot) ReadFile(name string) ([]byte, error) {
	path, err := root.doResolve(name, "read file")
	if err != nil {
		return nil, err
	}

	entry := root.cache.Get(path)
	data, err := entry.ReadFile()
	if err != nil {
		return nil, fmt.Errorf("could not read file `%s`: %v", name, err)
	}
	return data, nil
}

func (root fsRoot) ReadDir(name string) ([]fs.DirEntry, error) {
	path, err := root.doResolve(name, "read dir")
	if err != nil {
		return nil, err
	}

	list, err := root.cache.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("could not read dir `%s`: %v", name, err)
	}
	return list, nil
}

func (root fsRoot) Glob(pattern string) ([]string, error) {
	panic("TODO: glob")
}

func (root fsRoot) Resolve(name string) (Path, error) {
	return root.doResolve(name, "resolve")
}

func (root fsRoot) doResolve(name, op string) (Path, error) {
	out := root.path.Push(name)
	isRoot := out.IsRoot()
	if isRoot {
		out = out[1:]
	}

	if out.IsOutside() || !out.HasPrefix(root.path) {
		var err error
		if root.path.Len() > 0 && !isRoot {
			err = fmt.Errorf("%s: invalid path relative to `%s`: %s", op, root.path, name)
		} else {
			err = fmt.Errorf("%s: invalid path: %s", op, name)
		}
		return "", err
	}

	return out, nil
}

type fsLoader interface {
	Stat(path Path) (fs.FileInfo, error)
	ReadFile(path Path) ([]byte, error)
	ReadDir(path Path) ([]fs.DirEntry, error)
}

type fsCache struct {
	mutex   sync.Mutex
	loader  fsLoader
	entries map[Path]*fsEntry
}

func (cache *fsCache) Get(path Path) *fsEntry {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	out := cache.entries[path]
	if out == nil {
		out = fsEntryNew(cache, path, path.Base())
		if cache.entries == nil {
			cache.entries = make(map[Path]*fsEntry)
		}
		cache.entries[path] = out
	}
	return out
}

func (cache *fsCache) ReadDir(path Path) (list []fs.DirEntry, err error) {
	// TODO: make this consider in-memory items
	return cache.loader.ReadDir(path)
}

type fsLoaderFromRootDir struct {
	rootPath string
}

func (loader fsLoaderFromRootDir) Stat(path Path) (fs.FileInfo, error) {
	fullPath := filepath.Join(loader.rootPath, string(path))
	return os.Stat(fullPath)
}

func (loader fsLoaderFromRootDir) ReadFile(path Path) ([]byte, error) {
	fullPath := filepath.Join(loader.rootPath, string(path))
	return os.ReadFile(fullPath)
}

func (loader fsLoaderFromRootDir) ReadDir(path Path) ([]fs.DirEntry, error) {
	fullPath := filepath.Join(loader.rootPath, string(path))
	return os.ReadDir(fullPath)
}
