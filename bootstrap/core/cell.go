package core

import (
	"math"
	"sync/atomic"
)

type Cell[T any] struct {
	data atomic.Pointer[cellData[T]]
}

func (cell *Cell[T]) Get(version uint64) (out T) {
	data := cell.data.Load()
	if data != nil {
		ptr := (*data).Get(version)
		if ptr != nil {
			out = *ptr
		}
	}
	return
}

func (cell *Cell[T]) Set(version uint64, value T) {
	for {
		curData := cell.data.Load()

		var newData cellData[T]
		if curData == nil {
			newData = &cell1[T]{version, value}
		} else {
			newData = (*curData).Set(version, value)
		}

		if cell.data.CompareAndSwap(curData, &newData) {
			return
		}
	}
}

const cellVersionEdit uint64 = math.MaxUint64

type cellData[T any] interface {
	Get(version uint64) *T
	Set(version uint64, data T) cellData[T]
}

type cell1[T any] struct {
	version uint64
	data    T
}

func (cell *cell1[T]) Get(version uint64) *T {
	if cell.version == version {
		return &cell.data
	}
	return nil
}

func (cell *cell1[T]) Set(version uint64, data T) cellData[T] {
	if cellSet(&cell.version, &cell.data, version, &data) {
		return cell
	}

	out := &cell2[T]{
		version: [2]uint64{cell.version, version},
		data:    [2]T{cell.data, data},
	}
	return out
}

func cellSet[T any](version *uint64, data *T, newVersion uint64, newData *T) bool {
	if *version == newVersion || atomic.CompareAndSwapUint64(version, 0, cellVersionEdit) {
		*data = *newData
		atomic.StoreUint64(version, newVersion)
		return true
	}
	return false
}

type cell2[T any] struct {
	version [2]uint64
	data    [2]T
}

func (cell *cell2[T]) Get(version uint64) *T {
	if cell.version[0] == version {
		return &cell.data[0]
	}
	if cell.version[1] == version {
		return &cell.data[1]
	}
	return nil
}

func (cell *cell2[T]) Set(version uint64, data T) cellData[T] {
	if cellSet(&cell.version[0], &cell.data[0], version, &data) ||
		cellSet(&cell.version[1], &cell.data[1], version, &data) {
		return cell
	}

	out := &cellArr[T]{
		version: [cellArrLen]uint64{cell.version[0], cell.version[1], version},
		data:    [cellArrLen]T{cell.data[0], cell.data[1], data},
	}
	return out
}

const cellArrLen = 4

type cellArr[T any] struct {
	version [cellArrLen]uint64
	data    [cellArrLen]T
}

func (cell *cellArr[T]) Get(version uint64) *T {
	for n := range cell.version {
		if cell.version[n] == version {
			return &cell.data[n]
		}
	}
	return nil
}

func (cell *cellArr[T]) Set(version uint64, data T) (out cellData[T]) {
	for n := range cell.version {
		if cellSet(&cell.version[n], &cell.data[n], version, &data) {
			return cell
		}
	}

	out = cellTableNew[T](cellArrLen * 2)
	for n := range cell.version {
		version := atomic.LoadUint64(&cell.version[n])
		if version != 0 && version != cellVersionEdit {
			out = out.Set(version, cell.data[n])
		}
	}
	return out.Set(version, data)
}

const cellTableMod = 17

type cellTable[T any] struct {
	count   atomic.Uint64
	version []uint64
	data    []T
}

func cellTableNew[T any](size int) *cellTable[T] {
	size = max(size, 8) + size%2
	return &cellTable[T]{
		version: make([]uint64, size),
		data:    make([]T, size),
	}
}

func (cell *cellTable[T]) Get(version uint64) *T {
	inc := int(version%cellTableMod | 1)
	pos := int(version % uint64(len(cell.version)))
	for i := 0; i < len(cell.data); i++ {
		if v := cell.version[pos]; v == version {
			return &cell.data[pos]
		}
		pos = (pos + inc) % len(cell.version)
	}
	return nil
}

func (cell *cellTable[T]) Set(version uint64, data T) cellData[T] {
	for {
		cnt := cell.count.Load()
		if cnt*2 > uint64(len(cell.version)) {
			new := cellTableNew[T](len(cell.version) * 2)
			for i := 0; i < len(cell.data); i++ {
				if v := atomic.LoadUint64(&cell.version[i]); v != 0 && v != cellVersionEdit {
					new.Set(v, cell.data[i])
				}
			}
			return new.Set(version, data)
		}

		if cell.count.CompareAndSwap(cnt, cnt+1) {
			break
		}
	}

	inc := int(version%cellTableMod | 1)
	pos := int(version % uint64(len(cell.version)))
	for i := 0; i < len(cell.data); i++ {
		if cellSet(&cell.version[pos], &cell.data[pos], version, &data) {
			return cell
		}
		pos = (pos + inc) % len(cell.version)
	}

	panic("CellTable: failed to set")
}
