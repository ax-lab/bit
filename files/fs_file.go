package files

import (
	"fmt"
	"io"
	"io/fs"
	"sync/atomic"
)

type fsFile struct {
	entry *fsEntry
	pos   atomic.Int64
	dir   atomic.Int64
}

func fsFileNew(entry *fsEntry) *fsFile {
	return &fsFile{entry: entry}
}

func (file *fsFile) Path() Path {
	return file.entry.path
}

func (file *fsFile) ListDir() ([]File, error) {
	return file.entry.cache.ListDir(file.entry.path)
}

func (file *fsFile) Stat() (fs.FileInfo, error) {
	return file.entry.Info()
}

func (file *fsFile) Close() error {
	return nil
}

func (file *fsFile) ReadAt(out []byte, pos int64) (n int, err error) {
	fullData, err := file.entry.ReadFile()
	if err != nil {
		return 0, err
	}

	if pos < 0 || pos > int64(len(fullData)) {
		return 0, fmt.Errorf("invalid file offset")
	}

	data := fullData[pos:]
	size := min(len(out), len(data))
	if size == len(data) {
		err = io.EOF
	}

	if size > 0 {
		n = copy(out, data)
	}

	return
}

func (file *fsFile) Read(out []byte) (int, error) {
	fullData, err := file.entry.ReadFile()
	if err != nil {
		return 0, err
	}

	for {
		pos := file.pos.Load()
		idx := min(int(pos), len(fullData))
		data := fullData[idx:]
		if len(data) == 0 {
			return 0, io.EOF
		}

		readLen := min(len(data), len(out))
		if readLen > 0 {
			if !file.pos.CompareAndSwap(pos, pos+int64(readLen)) {
				continue
			}
			copy(out, data)
		}
		return readLen, nil
	}
}

func (file *fsFile) Seek(offset int64, whence int) (int64, error) {
	fullData, err := file.entry.ReadFile()
	if err != nil {
		return 0, err
	}

	size := int64(len(fullData))
	for {
		var offsetNew int64
		offsetCur := file.pos.Load()
		switch whence {
		case io.SeekStart:
			offsetNew = 0 + offset
		case io.SeekEnd:
			offsetNew = size - offset
		case io.SeekCurrent:
			offsetNew = offsetCur + offset
		default:
			panic("seek: invalid whence value")
		}

		offsetNew = min(offsetNew, size)
		if offsetNew < 0 {
			return offsetCur, fmt.Errorf("seek with negative file offset is invalid")
		}

		if file.pos.CompareAndSwap(offsetCur, offsetNew) {
			return offsetNew, nil
		}
	}
}

// Implements fs.ReadDirFile.
func (file *fsFile) ReadDir(n int) (out []fs.DirEntry, err error) {
	dirEntries, err := file.entry.cache.ReadDir(file.entry.path)
	if err != nil {
		return nil, err
	}

	for {
		pos := file.dir.Load()
		idx := min(len(dirEntries), int(pos))
		list := dirEntries[idx:]
		if len(list) == 0 {
			return nil, io.EOF
		}

		listLen := min(len(list), n)
		if listLen <= 0 {
			listLen = len(list)
		}

		if listLen > 0 {
			if !file.dir.CompareAndSwap(pos, pos+int64(listLen)) {
				continue
			}
			out = append(out, list[:listLen]...)
		}
		return out, nil
	}
}
