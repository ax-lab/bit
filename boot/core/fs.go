package core

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type FSLoader interface {
	Stat(path string) (fs.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]fs.DirEntry, error)
}

type FSRoot interface {
	Get(name string) (out File, valid bool)
	Glob(glob string, options ...GlobOption) (list []File, errs []error)
}

func FS(root string) (out FSRoot, err error) {
	root, err = filepath.EvalSymlinks(root)
	if err == nil {
		root, err = filepath.Abs(root)
	}

	if err == nil {
		out = &fsRoot{
			path: root,
			cache: &fsCache{
				loader: fsLoader{},
			},
		}
	}

	return out, err
}

type File struct {
	path  string
	root  *fsRoot
	entry *fsFile
}

func (file File) Valid() bool {
	return file.entry != nil && file.root != nil
}

func (file File) Path() string {
	return file.path
}

func (file File) Name() string {
	if file.path == "" || file.path == "." {
		return "."
	}
	return path.Base(file.path)
}

func (file File) NameWithoutExt() string {
	name := file.Name()
	if ext := path.Ext(name); ext != "" && len(ext) < len(name) {
		name = name[:len(name)-len(ext)]
	}
	return name
}

func (file File) Root() FSRoot {
	if file.root == nil {
		panic("File is invalid (at Root)")
	}
	return file.root
}

func (file File) Data() ([]byte, error) {
	if file.entry == nil {
		panic("File is invalid (at Data)")
	}
	return file.entry.Data()
}

func (file File) Info() (fs.FileInfo, error) {
	if file.entry == nil {
		panic("File is invalid (at Info)")
	}
	return file.entry.Stat()
}

func (file File) Exists() bool {
	_, err := file.Info()
	return err == nil || !errors.Is(err, fs.ErrNotExist)
}

func (file File) IsDir() bool {
	stat, _ := file.Info()
	if stat != nil {
		return stat.IsDir()
	}
	return false
}

func (file File) List(glob ...string) (list []File, err error) {
	info, err := file.Info()
	if err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, nil
	}

	var matchREs []*regexp.Regexp
	for _, it := range glob {
		matchREs = append(matchREs, regexp.MustCompile(GlobRegex(it)))
	}

	entries, err := file.entry.ListDir()
	for _, it := range entries {
		isMatch := len(matchREs) == 0
		for _, re := range matchREs {
			if re.MatchString(it.name) {
				isMatch = true
				break
			}
		}

		if !isMatch {
			continue
		}

		item := File{
			path:  it.name,
			root:  file.root,
			entry: it,
		}
		if file.path != "" && file.path != "." {
			name := make([]byte, 0, len(file.path)+len(item.path)+1)
			name = append(name, file.path...)
			name = append(name, fsPathSep)
			name = append(name, item.path...)
			item.path = string(name)
		}
		list = append(list, item)
	}
	return list, err
}

type fsRoot struct {
	path  string
	cache *fsCache
}

func (dir *fsRoot) Get(name string) (out File, valid bool) {
	fileName, filePath, valid := fsPathJoin(dir.path, name)
	if !valid {
		return out, false
	}

	out = File{
		path:  fileName,
		root:  dir,
		entry: dir.cache.Get(name, filePath),
	}
	if out.path == "" {
		if out.entry.path != dir.path {
			panic("FS: empty file name")
		}
		out.path = "."
	}
	return out, true
}

func (dir *fsRoot) Glob(glob string, options ...GlobOption) (list []File, errs []error) {
	root, _ := dir.Get("")
	return root.Glob(glob, options...)
}

func (file File) Glob(glob string, options ...GlobOption) (list []File, errs []error) {
	var (
		globRegex   *regexp.Regexp
		includeDirs bool
	)
	if glob != "" {
		globRegex = regexp.MustCompile(GlobRegex(glob))
	}

	for _, opt := range options {
		switch opt {
		case GlobIncludeDirs:
			includeDirs = true
		default:
			panic(fmt.Sprintf("Glob: invalid option `%s`", opt))
		}
	}

	queue := QueueNew(file)
	for queue.Len() > 0 {
		item, _ := queue.Shift()
		dirList, err := item.List()
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, fmt.Errorf("%s: stat failed -- %v", item.Name(), err))
			}
			continue
		}

		queue.Push(dirList...)

		if isRoot := item == file; isRoot {
			continue
		}

		if !includeDirs && item.IsDir() {
			continue
		}

		if globRegex != nil && !globRegex.MatchString(item.Name()) {
			continue
		}

		list = append(list, item)
	}

	return
}

type fsCache struct {
	files  sync.Map // map[string]*fsFile
	loader FSLoader
}

func (cache *fsCache) Get(name, path string) *fsFile {
	file, _ := cache.files.LoadOrStore(path, fsFileNew(cache, name, path))
	return file.(*fsFile)
}

type fsFile struct {
	cache *fsCache
	name  string
	path  string

	dirs     *fsDirs
	dirsSync sync.Mutex
	dirsDone atomic.Bool

	stat     *fsStat
	statSync sync.Mutex
	statDone atomic.Bool

	data     []byte
	dataErr  error
	dataSync sync.Mutex
	dataDone atomic.Bool
}

func fsFileNew(cache *fsCache, name, path string) *fsFile {
	return &fsFile{cache: cache, name: name, path: path}
}

func (file *fsFile) Reload() {
	file.dirsSync.Lock()
	file.dirsDone.Store(false)
	file.dirs = nil
	file.dirsSync.Unlock()

	file.statSync.Lock()
	file.statDone.Store(false)
	file.stat = nil
	file.statSync.Unlock()

	file.dataSync.Lock()
	file.dataDone.Store(false)
	file.data, file.dataErr = nil, nil
	file.dataSync.Unlock()
}

func (file *fsFile) Data() ([]byte, error) {
	if file.dataDone.Load() {
		return file.data, file.dataErr
	}

	file.dataSync.Lock()
	defer file.dataSync.Unlock()

	if file.dataDone.Load() {
		return file.data, file.dataErr
	}
	defer file.dataDone.Store(true)

	file.data, file.dataErr = file.cache.loader.ReadFile(file.path)
	return file.data, file.dataErr
}

func (file *fsFile) Stat() (fs.FileInfo, error) {
	stat := file.statLoadOrStore(nil)
	if stat.err != nil {
		return nil, stat.err
	}
	return stat, nil
}

func (file *fsFile) ListDir() ([]*fsFile, error) {
	if file.dirsDone.Load() {
		return file.dirs.list, file.dirs.err
	}

	file.dirsSync.Lock()
	defer file.dirsSync.Unlock()

	if file.dirsDone.Load() {
		return file.dirs.list, file.dirs.err
	}
	defer file.dirsDone.Store(true)

	list, err := file.cache.loader.ReadDir(file.path)
	dirs := &fsDirs{err: err}
	for _, dirEntry := range list {
		itemName := dirEntry.Name()
		itemPath := path.Join(file.path, itemName)
		item := file.cache.Get(itemName, itemPath)
		item.statLoadOrStore(dirEntry)
		dirs.list = append(dirs.list, item)
	}
	file.dirs = dirs

	return file.dirs.list, file.dirs.err
}

func (file *fsFile) statLoadOrStore(dirEntry fs.DirEntry) *fsStat {
	if file.statDone.Load() {
		return file.stat
	}

	file.statSync.Lock()
	defer file.statSync.Unlock()

	if file.statDone.Load() {
		return file.stat
	}
	defer file.statDone.Store(true)

	var (
		info fs.FileInfo
		err  error
	)
	if dirEntry != nil {
		info, err = dirEntry.Info()
	} else {
		info, err = file.cache.loader.Stat(file.path)
	}

	file.stat = &fsStat{err: err}
	if info != nil {
		file.stat.mode = info.Mode()
		file.stat.name = info.Name()
		file.stat.size = info.Size()
		file.stat.time = info.ModTime()
	}

	return file.stat
}

type fsDirs struct {
	list []*fsFile
	err  error
}

type fsStat struct {
	err  error
	name string
	size int64
	time time.Time
	mode fs.FileMode
}

// fs.FileInfo
func (stat *fsStat) Name() string       { return stat.name }
func (stat *fsStat) Size() int64        { return stat.size }
func (stat *fsStat) Mode() fs.FileMode  { return stat.mode }
func (stat *fsStat) ModTime() time.Time { return stat.time }
func (stat *fsStat) IsDir() bool        { return stat.mode.IsDir() }
func (stat *fsStat) Sys() any           { return nil }

type fsLoader struct{}

// FileLoader
func (fsLoader) Stat(path string) (fs.FileInfo, error)      { return os.Stat(path) }
func (fsLoader) ReadFile(path string) ([]byte, error)       { return os.ReadFile(path) }
func (fsLoader) ReadDir(path string) ([]fs.DirEntry, error) { return os.ReadDir(path) }

const fsPathSep = '/'

func fsPathJoin(root string, items ...string) (fileName, filePath string, valid bool) {
	name := fsPathBuilder{sep: fsPathSep}
	path := fsPathBuilder{sep: filepath.Separator, buf: []byte(root)}

	valid = true
	for _, it := range items {
		cur := []byte(it)
		sta := 0
		for idx := 0; valid && idx < len(cur); idx++ {
			if cur[idx] == '/' || cur[idx] == '\\' || cur[idx] == filepath.Separator {
				if idx > sta {
					if name.Push(cur[sta:idx]) {
						path.Push(cur[sta:idx])
					} else {
						valid = false
					}
				}
				sta = idx + 1
			}
		}
		if sta < len(cur) {
			if name.Push(cur[sta:]) {
				path.Push(cur[sta:])
			} else {
				valid = false
			}
		}
	}

	if valid {
		fileName = string(name.buf)
		filePath = string(path.buf)
	}
	return
}

type fsPathBuilder struct {
	sep byte
	buf []byte
}

func (pb *fsPathBuilder) Push(elem []byte) bool {
	if len(elem) == 0 || (len(elem) == 1 && elem[0] == '.') {
		return false
	}

	if len(elem) == 2 && elem[0] == '.' && elem[1] == '.' {
		if len(pb.buf) == 0 {
			return false
		}
		idx := bytes.LastIndexByte(pb.buf, pb.sep)
		if idx >= 0 {
			pb.buf = pb.buf[:idx]
		} else {
			pb.buf = pb.buf[:0]
		}
		return true
	}

	if len(pb.buf) == 0 {
		pb.buf = append(pb.buf, elem...)
	} else {
		pb.buf = slices.Grow(pb.buf, 1+len(elem))
		pb.buf = append(pb.buf, pb.sep)
		pb.buf = append(pb.buf, elem...)
	}
	return true
}
