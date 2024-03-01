package core

import (
	"fmt"
	"slices"
	"sync/atomic"
	"unsafe"
)

const (
	tablePageSize uint64 = 256
)

type tableBuffer[T any] struct {
	data atomic.Pointer[tablePage[T]]
}

func (tb *tableBuffer[T]) ToList() []T {
	size := 0
	return tb.data.Load().AppendToList(nil, &size)
}

func (tb *tableBuffer[T]) Get(n uint64) *T {
	cur := tb.data.Load()
	off := uint64(0)
	for cur != nil {
		if cur.itemSize == 1 {
			return cur.GetAsItem(n - off)
		} else {
			next := (n - off) / cur.itemSize
			off = next * cur.itemSize
			cur = cur.GetAsPage(next)
		}
	}
	return nil
}

func (tb *tableBuffer[T]) Set(n uint64, value *T) {
	// reserve space for the new item
	for {
		curData := tb.data.Load()
		if curData != nil && curData.itemSize*tablePageSize > n {
			break
		}

		newData := &tablePage[T]{
			pageOwner: tb,
			itemSize:  1,
		}
		if curData != nil {
			if curData.itemSize == 0 {
				panic("curData.itemSize is zero")
			}
			newData.itemSize = curData.itemSize * tablePageSize
			newData.InitPage(0, curData)
		}
		tb.data.CompareAndSwap(curData, newData) // drop the new page if this fails
	}

	pages := newTablePageStack(tb)
	pages.SetItem(n, value)
}

// Fixed size list of either pages or items, depending on `blockSize`.
type tablePage[T any] struct {
	pageOwner *tableBuffer[T]
	itemSize  uint64
	itemList  [tablePageSize]unsafe.Pointer
}

func newTablePage[T any](pageOwner *tableBuffer[T], itemSize uint64) *tablePage[T] {
	if itemSize == 0 || (itemSize > 1 && itemSize%tablePageSize != 0) {
		panic(fmt.Sprintf("creating table page with invalid itemSize of %d", itemSize))
	}
	out := &tablePage[T]{
		pageOwner: pageOwner,
		itemSize:  itemSize,
	}
	return out
}

// Used to copy an immutable page to a new owner so it can be modified.
func (page *tablePage[T]) NewCopy(newPageOwner *tableBuffer[T]) *tablePage[T] {
	out := newTablePage[T](newPageOwner, page.itemSize)
	for i := range out.itemList {
		atomic.StorePointer(&out.itemList[i], atomic.LoadPointer(&page.itemList[i]))
	}
	return out
}

func (page *tablePage[T]) AppendToList(ls []T, curLen *int) []T {
	if page == nil {
		return ls
	}

	growSlice := func(ls []T, newLen int, extra int) []T {
		cur := len(ls)
		ls = slices.Grow(ls, newLen+extra-cur)
		return ls[:newLen]
	}

	size := *curLen
	if page.itemSize > 1 {
		for i := range page.itemList {
			it := page.GetAsPage(uint64(i))
			if it != nil {
				ls = it.AppendToList(ls, &size)
			} else {
				size += int(tablePageSize)
			}
		}
	} else {
		for i := range page.itemList {
			if ptr := page.GetAsItem(uint64(i)); ptr != nil {
				ls = growSlice(ls, size, 1)
				ls = append(ls, *ptr)
			}
			size += 1
		}
	}

	*curLen = size
	return ls
}

func (page *tablePage[T]) GetAsItem(n uint64) *T {
	return (*T)(atomic.LoadPointer(&page.itemList[n]))
}

func (page *tablePage[T]) GetAsPage(n uint64) *tablePage[T] {
	return (*tablePage[T])(atomic.LoadPointer(&page.itemList[n]))
}

func (page *tablePage[T]) InitPage(pos uint64, item *tablePage[T]) {
	if !page.CompareAndSwapPage(pos, nil, item) {
		panic("Table: page init failed: already init")
	}
}

func (page *tablePage[T]) CompareAndSwapPage(pos uint64, old *tablePage[T], new *tablePage[T]) bool {
	return atomic.CompareAndSwapPointer(&page.itemList[pos], unsafe.Pointer(old), unsafe.Pointer(new))
}

func (page *tablePage[T]) CompareAndSwapItem(pos uint64, old *T, new *T) bool {
	return atomic.CompareAndSwapPointer(&page.itemList[pos], unsafe.Pointer(old), unsafe.Pointer(new))
}

// Stack of pages for a recursive write operation. This is used to copy parent
// pages on demand when a leaf page is copied on write.
type tablePageStack[T any] struct {
	buffer  *tableBuffer[T]
	page    *tablePage[T]
	pageIdx uint64
	prev    *tablePageStack[T]
}

func newTablePageStack[T any](buffer *tableBuffer[T]) tablePageStack[T] {
	out := tablePageStack[T]{
		buffer:  buffer,
		page:    buffer.data.Load(),
		prev:    nil,
		pageIdx: 0,
	}
	return out
}

func (ref *tablePageStack[T]) MakePageOwnedForWrite() {
	if needToCopy := ref.page.pageOwner != ref.buffer; needToCopy {
		newPage := ref.page.NewCopy(ref.buffer)
		if hasPrevious := ref.prev != nil; hasPrevious {
			ref.prev.MakePageOwnedForWrite()
			parent := ref.prev.page
			if parent.CompareAndSwapPage(ref.pageIdx, ref.page, newPage) {
				ref.page = newPage
			} else {
				ref.page = parent.GetAsPage(ref.pageIdx)
			}
		} else if ref.buffer.data.CompareAndSwap(ref.page, newPage) {
			ref.page = newPage
		} else {
			ref.page = ref.buffer.data.Load()
		}

		if ref.page.pageOwner != ref.buffer {
			panic("Table: swapped page has wrong ownership (concurrent write?)")
		}
	}
}

func (ref *tablePageStack[T]) SetItem(pos uint64, val *T) {
	if ref.page == nil {
		panic("SetItem: tablePageStack.page is nil")
	}

	if isPageGroup := ref.page.itemSize > 1; isPageGroup {
		nextIndex := pos / ref.page.itemSize
		nextPage := ref.page.GetAsPage(nextIndex)
		if nextPage == nil {
			ref.MakePageOwnedForWrite()
			nextPage = ref.page.GetAsPage(nextIndex)
			if nextPage == nil {
				if ref.page.itemSize < tablePageSize || ref.page.itemSize%tablePageSize != 0 {
					panic(fmt.Sprintf("ref.page.itemSize is invalid = %d", ref.page.itemSize))
				}
				nextPage = newTablePage(ref.buffer, ref.page.itemSize/tablePageSize)
				ref.page.InitPage(nextIndex, nextPage)
			}
		}

		pageOffset := nextIndex * ref.page.itemSize
		nextRef := tablePageStack[T]{
			buffer:  ref.buffer,
			page:    nextPage,
			pageIdx: nextIndex,
			prev:    ref,
		}
		nextRef.SetItem(pos-pageOffset, val)
		return
	}

	cur := ref.page.GetAsItem(pos)
	if cur == val {
		return // ignore no-op writes
	}

	ref.MakePageOwnedForWrite()
	if !ref.page.CompareAndSwapItem(pos, cur, val) {
		panic("Table: set in leaf page failed due to concurrent write")
	}
}
