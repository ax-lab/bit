package core

import (
	"math"
	"runtime"
	"slices"
	"sync/atomic"
)

type DataSet[T any] struct {
	inner *dataSetOf[T]
}

func (set DataSet[T]) Clone() DataSet[T] {

	return DataSet[T]{set.inner.Clone()}
}

func (set *DataSet[T]) Set(id uint64, value T) {
	if set.inner == nil {
		set.inner = &dataSetOf[T]{}
	}
	dataSetOfWrite(set.inner, id, value)
}

func (set DataSet[T]) Get(id uint64) (out T) {
	return dataSetOfRead[T](set.inner, id)
}

type dataSetOf[T any] struct {
	pages       atomic.Pointer[[]atomic.Pointer[dataSetOfPage[T]]]
	editVersion atomic.Uint64
	lockVersion atomic.Uint64
}

func (set *dataSetOf[T]) Clone() *dataSetOf[T] {
	if set == nil {
		return nil
	}

	ptrPages := set.pages.Load()
	if ptrPages == nil {
		return &dataSetOf[T]{}
	}

	curPages := *ptrPages
	newPages := make([]atomic.Pointer[dataSetOfPage[T]], len(curPages))
	for n := range curPages {
		newPages[n].Store(curPages[n].Load())
	}

	// prevent further modifications to the cloned pages
	set.lockVersion.Add(1)

	out := &dataSetOf[T]{}
	out.pages.Store(&newPages)
	return out
}

func (set *dataSetOf[T]) index(pos uint64) (page, index int) {
	if pos >= uint64(math.MaxInt) {
		panic("DataSet: position overflow")
	}

	page = int(pos / dataSetOfPageSize)
	index = int(pos % dataSetOfPageSize)
	return
}

func dataSetOfRead[T any](set *dataSetOf[T], pos uint64) T {
	page, index := set.index(pos)
	return dataSetOfReadIndex[T](set, page, index)
}

func dataSetOfReadIndex[T any](set *dataSetOf[T], page, offset int) (out T) {
	if set == nil {
		return
	}

	ptrPages := set.pages.Load()
	if ptrPages == nil {
		return
	}

	pages := *ptrPages
	if page := pages[page].Load(); page != nil {
		return page.data[offset]
	}

	return
}

func dataSetOfWrite[T any](set *dataSetOf[T], pos uint64, newValue T) {
	if set == nil {
		panic("DataSet: writing to nil set")
	}
	pageIndex, itemIndex := set.index(pos)

	for {
		lockVersion := set.lockVersion.Load()
		editVersion := set.editVersion.Load()
		if editVersion%2 != 0 {
			runtime.Gosched()
		}

		// extend the page list if necessary
		curPageList := set.pages.Load()
		newPageList := curPageList
		if newPageList == nil || len(*newPageList) <= pageIndex {
			var pageList []atomic.Pointer[dataSetOfPage[T]]
			if newPageList != nil {
				pageList = *newPageList
			}
			pageList = slices.Grow(pageList, pageIndex+1-len(pageList))
			pageList = pageList[:cap(pageList)]
			newPageList = &pageList
		}

		pageList := *newPageList

		// make sure the page is allocated and exclusive
		curPage := pageList[pageIndex].Load()
		newPage := curPage
		if curPage == nil {
			newPage = &dataSetOfPage[T]{set: set, share: lockVersion}
		} else if isSharedPage := curPage.set != set || curPage.share != lockVersion; isSharedPage {
			newPage = &dataSetOfPage[T]{set: set, share: lockVersion}
			copy(newPage.data[:], curPage.data[:])
		}

		if set.editVersion.CompareAndSwap(editVersion, editVersion+1) {
			newPage.data[itemIndex] = newValue

			if !pageList[pageIndex].CompareAndSwap(curPage, newPage) {
				panic("DataSet: page commit failed")
			}

			if !set.pages.CompareAndSwap(curPageList, newPageList) {
				panic("DataSet: page list commit failed")
			}

			if !set.editVersion.CompareAndSwap(editVersion+1, editVersion+2) {
				panic("DataSet: version commit failed")
			}
		} else {
			continue
		}

		// write was successful
		break
	}
}

const (
	dataSetOfPageSize = 256
)

type dataSetOfPage[T any] struct {
	set   *dataSetOf[T]
	share uint64
	data  [dataSetOfPageSize]T
}
