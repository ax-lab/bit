package core

import (
	"math"
	"runtime"
	"slices"
	"sync/atomic"
	"unsafe"
)

type dataSet struct {
	pages       atomic.Pointer[[]atomic.Pointer[dataSetPage]]
	editVersion atomic.Uint64
	lockVersion atomic.Uint64
}

func (set *dataSet) Clone() *dataSet {
	if set == nil {
		return &dataSet{}
	}

	curPages := *set.pages.Load()
	if curPages == nil {
		return &dataSet{}
	}

	newPages := make([]atomic.Pointer[dataSetPage], len(curPages))
	for i := range curPages {
		newPages[i].Store(curPages[i].Load())
	}

	// prevent further modifications to the cloned pages
	set.lockVersion.Add(1)

	out := &dataSet{}
	out.pages.Store(&newPages)
	return out
}

func (set *dataSet) index(pos uint64) (page, block, offset int) {
	if pos >= uint64(math.MaxInt) {
		panic("DataSet: position overflow")
	}

	const perPage uint64 = (dataSetPageSize * dataSetBlockSize)
	page = int(pos / perPage)
	block = int((pos % perPage) / dataSetBlockSize)
	offset = int(pos % dataSetBlockSize)
	return
}

func dataSetRead[T any](set *dataSet, pos uint64) *T {
	page, block, index := set.index(pos)
	return dataSetReadIndex[T](set, page, block, index)
}

func dataSetReadIndex[T any](set *dataSet, page, block, offset int) *T {
	if set == nil {
		return nil
	}

	pagesPtr := set.pages.Load()
	if pagesPtr == nil {
		return nil
	}

	pages := *pagesPtr
	if page < len(pages) {
		if pg := pages[page].Load(); pg != nil {
			ch := pg.blocks[block].Load()
			if ch != nil {
				return (*T)(atomic.LoadPointer(&ch.data[offset]))
			}
		}
	}
	return nil
}

func dataSetWrite[T any](set *dataSet, pos uint64, newValue *T) {
	if set == nil {
		panic("DataSet: writing to nil set")
	}
	pageIndex, blockIndex, itemIndex := set.index(pos)
	curValue := dataSetReadIndex[T](set, pageIndex, blockIndex, itemIndex)
	if curValue == newValue {
		return
	}

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
			var pageList []atomic.Pointer[dataSetPage]
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
			newPage = &dataSetPage{set: set, share: lockVersion}
		} else if isSharedPage := curPage.set != set || curPage.share != lockVersion; isSharedPage {
			newPage = &dataSetPage{set: set, share: lockVersion}
			for i := range curPage.blocks {
				newPage.blocks[i].Store(curPage.blocks[i].Load())
			}
		}

		// make sure the block is allocated and exclusive
		curBlock := newPage.blocks[blockIndex].Load()
		newBlock := curBlock
		if curBlock == nil {
			newBlock = &dataSetBlock{set: set, share: lockVersion}
		} else if curBlock.set != set || curBlock.share != lockVersion {
			newBlock = &dataSetBlock{set: set, share: lockVersion}
			for i := range curBlock.data {
				value := atomic.LoadPointer(&curBlock.data[i])
				atomic.StorePointer(&newBlock.data[i], value)
			}
		}

		if set.editVersion.CompareAndSwap(editVersion, editVersion+1) {
			if !atomic.CompareAndSwapPointer(&newBlock.data[itemIndex], unsafe.Pointer(curValue), unsafe.Pointer(newValue)) {
				panic("DataSet: concurrent write detected")
			}

			if !newPage.blocks[blockIndex].CompareAndSwap(curBlock, newBlock) {
				panic("DataSet: block commit failed")
			}

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
	dataSetPageSize  = 1024
	dataSetBlockSize = 1024
)

type dataSetPage struct {
	set    *dataSet
	share  uint64
	blocks [dataSetPageSize]atomic.Pointer[dataSetBlock]
}

type dataSetBlock struct {
	set   *dataSet
	share uint64
	data  [dataSetBlockSize]unsafe.Pointer
}
