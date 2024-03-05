package core

import (
	"sync/atomic"
)

const (
	// Any page smaller than this will be merged
	pageSizeMin = 256

	// Desired page size, when allocating new pages
	pageSizeHint = 1024
)

/*
	Defragmentation strategy
	========================

	When inserting new pages, any page smaller than `SizeHint` will check for
	a merge possibility.

	Pages smaller than `SizeMin` will always try to merge with neighbor pages:

	- Merge with both neighbors if either:
		- Both are smaller than `SizeMin`;
		- The merged size would be less or equal to `SizeHint`.
	- Otherwise pick the smaller neighbor with size less than `SizeHint`.

	When allocating new pages, sizes up to `SizeHint` are rounded to the next
	power-of-two, and further doubled for a last page (up to `SizeHint`).

	New pages are flagged as private.
*/

type Cord[T any] struct {
	page   []Page[T]
	size   []int
	cached atomic.Uint64
}

func (cord *Cord[T]) Len() int {
	if last := len(cord.size) - 1; last >= 0 {
		return cord.size[last]
	}
	return 0
}

func (cord *Cord[T]) Range(sta, end int) (pre Page[T], mid []Page[T], pos Page[T]) {
	if end < sta {
		panic("Page: invalid range")
	}

	idxSta, posSta := cord.pageIndex(sta)
	idxEnd, posEnd := idxSta, posSta
	if end > sta {
		idxEnd, posEnd = cord.pageIndex(end)
	}

	if idxSta < 0 || idxEnd < 0 {
		panic("Page: range out of bounds")
	} else if idxSta > idxEnd || posSta < 0 || posEnd < 0 {
		panic("Page: invalid range calculation [BUG]")
	}

	if isEnd := idxSta == len(cord.page); isEnd {
		if idxEnd != idxSta || posSta != 0 || posEnd != 0 {
			panic("Page: invalid range calculation at end [BUG]")
		}
		return
	}

	if posEnd > cord.page[idxEnd].Len() {
		panic("Page: invalid range calculation - posEnd out of bounds [BUG]")
	}

	pre = cord.page[idxSta]
	if idxEnd == idxSta {
		len := pre.Len()
		pre.sta = pre.sta + posSta
		pre.end = pre.end - (len - posEnd)
	} else {
		pre.sta += posSta
	}

	if pre.sta > pre.end || pre.sta == pre.end && end > sta {
		panic("Page: range calculation generated invalid prefix [BUG]")
	}

	if idxEnd != idxSta {
		mid = cord.page[idxSta+1 : idxEnd]
		pos = cord.page[idxEnd]
		pos.end = pos.end - (pos.Len() - posEnd)
	}

	return
}

func (cord *Cord[T]) Splice(sta, end int, data ...Page[T]) {
	panic("TODO")
}

func (cord *Cord[T]) pageIndex(pos int) (index int, offset int) {
	sta, end := 0, len(cord.page)

	// common special cases
	if pos == 0 {
		return 0, 0
	}
	if last := end - 1; last >= 0 && pos == cord.size[last] {
		return end, 0
	}

	// try the last accessed page
	if idx := int(cord.cached.Load()); idx < len(cord.size) {
		cnt := cord.size[idx]
		if pos == cnt {
			return idx + 1, 0
		} else if pos < cnt {
			sta := cnt - cord.page[idx].Len()
			if pos >= sta {
				return idx, pos - sta
			}
		}
	}

	index = -1

	for sta < end {
		mid := (sta + end) / 2
		cnt := cord.size[mid]
		if pos > cnt {
			sta = mid + 1
		} else if pos == cnt {
			// this might not exist yet, but we don't care here
			index = mid + 1
			offset = 0
			break
		} else if pageSta := cnt - cord.page[mid].Len(); pos < pageSta {
			end = mid
		} else {
			index = mid
			offset = pos - pageSta
			break
		}
	}

	if index >= 0 {
		cord.cached.Store(uint64(index))
	}

	return
}

type Page[T any] struct {
	sta    int
	end    int
	buffer *pageBuffer[T]
}

func (page *Page[T]) Len() int {
	return page.end - page.sta
}

func (page *Page[T]) Cap() int {
	if page.buffer == nil {
		return 0
	}
	return page.buffer.Cap(page.sta)
}

func (page *Page[T]) Data() []T {
	if page.buffer == nil {
		return nil
	}
	return page.buffer.data[page.sta:page.end]
}

const (
	pageSizeMask   = ^uint64(0) >> 1
	pageSharedFlag = ^pageSizeMask
)

type pageBuffer[T any] struct {
	size atomic.Uint64
	data []T
}

func newPageBufferStatic[T any](data ...T) *pageBuffer[T] {
	page := &pageBuffer[T]{data: data}
	page.size.Store(uint64(len(data)))
	return page
}

func newPageBufferWithCapacity[T any](capacity int, data ...T) *pageBuffer[T] {
	if capacity <= 0 || capacity < len(data) {
		panic("Page: invalid buffer capacity")
	}

	page := &pageBuffer[T]{
		data: make([]T, capacity),
	}
	page.size.Store(uint64(len(data)))
	copy(page.data, data)

	return page
}

func (page *pageBuffer[T]) Cap(at int) int {
	size := int(page.size.Load() & pageSizeMask)
	if at == size {
		return len(page.data) - size
	}
	return 0
}

// Reserve the unused space in buffer.
//
// This is safe even if the buffer is shared, including between threads.
func (page *pageBuffer[T]) Reserve(index int, size int) bool {
	if index < 0 || size <= 0 {
		panic("Page: invalid push index or size")
	}

	if index+size > len(page.data) {
		return false
	}

	rangeSta := uint64(index)
	rangeEnd := rangeSta + uint64(size)
	if rangeEnd > pageSizeMask {
		return false
	}

	curFlag := page.size.Load() & pageSharedFlag
	curSize := curFlag + rangeSta
	newSize := curFlag + rangeEnd
	return page.size.CompareAndSwap(curSize, newSize)
}
