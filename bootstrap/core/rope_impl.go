package core

import "slices"

type ropeBuffer[T any] struct {
	Blocks []ropeBlock[T]
}

func (rope *ropeBuffer[T]) Splice(sta, end int, pageSize [2]int, data RopeIter[T]) {
	if sta > end {
		panic("rope: invalid range")
	} else if sta == end && data.Count(1) == 0 {
		return
	}

	idxSta, validSta := rope.PageIndex(sta)
	idxEnd, validEnd := idxSta, validSta
	if end != sta {
		if !validSta {
			panic("rope: invalid start index")
		}
		idxEnd, validEnd = rope.PageIndex(end)
	}

	blkSta := idxSta[0]
	blkEnd := idxEnd[0]

	pageSta := idxSta[1]
	pageEnd := idxEnd[1]

	posSta := idxSta[2]
	posEnd := idxEnd[2]

	head := ropeBlock[T]{}
	if blkSta < len(rope.Blocks) {
		head = rope.Blocks[blkSta]
	}
	head.Pages = append(([][]T)(nil), head.Pages[:pageSta]...)

	var (
		pagePrefix  RopeIter[T]
		pageSuffix  RopeIter[T]
		blockSuffix RopeIter[T]
	)
	if posSta > 0 {
		data := rope.Blocks[blkSta].Pages[pageSta][:posSta]
		pagePrefix = RopeIterChunks(data)
	}

	if validEnd {
		if posEnd > 0 {
			sizeEnd := len(rope.Blocks[blkEnd].Pages[pageEnd])
			if posEnd < sizeEnd {
				data := rope.Blocks[blkEnd].Pages[pageEnd][posEnd:]
				pageSuffix = RopeIterChunks(data)
			}
			pageEnd, posEnd = pageEnd+1, 0
		}

		if pageEnd < len(rope.Blocks[blkEnd].Pages) {
			pages := rope.Blocks[blkEnd].Pages[pageEnd:]
			blockSuffix = RopeIterChunks(pages...)
		}
	}

	data = RopeIterChain(pagePrefix, data, pageSuffix, blockSuffix)
	tail := head.Push(false, pageSize, data)

	pre := rope.Blocks[:blkSta]
	pos := rope.Blocks[blkEnd:]

	rope.Blocks = make([]ropeBlock[T], 0, len(pre)+1+len(tail)+len(pos))
	rope.Blocks = append(rope.Blocks, pre...)
	rope.Blocks = append(rope.Blocks, head)
	rope.Blocks = append(rope.Blocks, tail...)
	rope.Blocks = append(rope.Blocks, pos...)

	prevEnd := 0
	for i := range rope.Blocks {
		it := &rope.Blocks[i]
		it.offset = prevEnd
		prevEnd = it.End()
	}
}

func (rope *ropeBuffer[T]) PageIndex(n int) (index [3]int, valid bool) {
	if n == 0 {
		return [3]int{0, 0, 0}, len(rope.Blocks) > 0
	}
	sta, end := 0, len(rope.Blocks)
	for sta < end {
		mid := (end-sta)/2 + sta
		block := rope.Blocks[mid]
		if n < block.Sta() {
			end = mid
		} else if blockEnd := block.End(); n > blockEnd {
			sta = mid + 1
		} else if n == blockEnd {
			valid = mid < len(rope.Blocks)-1
			if valid && len(rope.Blocks[mid+1].Pages[0]) == 0 {
				panic("Rope: block has empty page")
			}
			return [3]int{mid + 1, 0, 0}, valid
		} else {
			index, valid := block.PageIndex(n)
			return [3]int{mid, index[0], index[1]}, valid
		}
	}

	panic("rope: invalid index")
}

type ropeBlock[T any] struct {
	offset int
	Pages  [][]T
	Count  []int
}

func (rope *ropeBlock[T]) Sta() int {
	return rope.offset
}

func (rope *ropeBlock[T]) End() int {
	if idxLast := len(rope.Pages) - 1; idxLast >= 0 {
		return rope.offset + rope.Count[idxLast]
	} else {
		return rope.offset
	}
}

func (rope *ropeBlock[T]) PageIndex(n int) (index [2]int, valid bool) {
	if n < rope.Sta() || rope.End() < n {
		panic("rope block: index out of bounds")
	}

	pos := n - rope.offset
	sta, end := 0, len(rope.Pages)
	for sta < end {
		mid := (end-sta)/2 + sta
		page := rope.Pages[mid]
		pageEnd := rope.Count[mid]
		if pos == pageEnd {
			valid = mid < len(rope.Pages)-1
			return [2]int{mid + 1, 0}, valid
		} else if pos > pageEnd {
			sta = mid + 1
		} else if pos < pageEnd {
			pageSta := pageEnd - len(page)
			if pos < pageSta {
				end = mid
			} else {
				return [2]int{mid, pos - pageSta}, true
			}
		}
	}

	panic("rope block: index is valid but was not found")
}

func (rope *ropeBlock[T]) Push(ownsLast bool, pageSize [2]int, data RopeIter[T]) (overflow []ropeBlock[T]) {
	builder := ropeBuilder[T]{
		OwnsLast: ownsLast,
		PageSize: pageSize[1],
		Pages:    rope.Pages,
		Count:    rope.Count,
	}
	builder.PushData(data)

	maxSize := pageSize[0]
	curSize := min(maxSize, len(builder.Pages))
	rope.Pages = builder.Pages[:curSize]
	rope.Count = builder.Count[:curSize]

	for len(builder.Pages) > 0 {
		builder.Pages = builder.Pages[curSize:]
		builder.Count = builder.Count[curSize:]

		curSize = min(maxSize, len(builder.Pages))
		if curSize > 0 {
			next := ropeBlock[T]{
				Pages: builder.Pages[:curSize],
				Count: builder.Count[:curSize],
			}
			overflow = append(overflow, next)
		}
	}

	return
}

type ropeBuilder[T any] struct {
	OwnsLast bool
	PageSize int
	Pages    [][]T
	Count    []int
}

func (rope *ropeBuilder[T]) MergePages(pages [][]T, pagesOwned bool) {
	if len(rope.Pages) > 0 {
		totalLen := len(rope.Pages[0])
		mergeCnt := 0
		for mergeCnt < len(pages) && totalLen+len(pages[mergeCnt]) <= rope.PageSize {
			totalLen += len(pages[mergeCnt])
			mergeCnt++
		}

		if mergeCnt > 0 {
			lastPage, lastIdx := rope.getOwnedLastPage(mergeCnt)
			for _, it := range pages[:mergeCnt] {
				if len(lastPage)+len(it) > cap(lastPage) {
					panic("rope: last page has insufficient capacity for merge")
				}
				lastPage = append(lastPage, it...)
			}
			rope.Pages[lastIdx] = lastPage
			pages = pages[mergeCnt:]
		}
	}

	rope.Count = slices.Grow(rope.Count, len(pages))
	rope.Pages = slices.Grow(rope.Pages, len(pages))
	for _, it := range rope.Pages {
		rope.pushPage(it, pagesOwned)
	}
}

func (rope *ropeBuilder[T]) PushData(data RopeIter[T]) {
	pageSize := rope.PageSize
	if pageSize <= 0 {
		panic("rope.PushData: invalid page size")
	}

	if len(rope.Pages) != len(rope.Count) {
		panic("rope.PushData: invalid index")
	}

	canAppendToLast := len(rope.Pages) > 0 && len(rope.Pages[len(rope.Pages)-1]) < pageSize

	for {
		chunk := data.Next()
		if len(chunk) == 0 {
			break
		}

		var pushCount int
		if len(chunk) == pageSize {
			// when incoming chunks are exact full pages, trying to merge
			// could cause a runaway cascades of merges
			pushCount = pageSize
			rope.pushPage(chunk, false)
			canAppendToLast = false
		} else if !canAppendToLast {
			pushCount = min(len(chunk), pageSize)
			rope.pushPage(chunk[:pushCount], false)
			canAppendToLast = pushCount < pageSize
		} else {
			remainingLen := data.Count(pageSize)
			lastPage, lastIdx := rope.getOwnedLastPage(remainingLen)
			pushCount = min(len(chunk), pageSize-len(lastPage))
			if pushCount <= 0 || cap(lastPage) < len(lastPage)+pushCount {
				panic("rope.PushData: last page has insufficient capacity")
			}

			rope.Pages[lastIdx] = append(lastPage, chunk[:pushCount]...)
			canAppendToLast = len(rope.Pages[lastIdx]) < pageSize
		}

		if pushCount <= 0 {
			panic("rope.PushData: push count was invalid")
		}
		data.Skip(pushCount)
	}
}

func (rope *ropeBuilder[T]) getOwnedLastPage(remainingLen int) (page []T, lastPos int) {
	lastPos = len(rope.Pages) - 1
	if lastPos < 0 {
		panic("rope: there is no last page")
	}

	lastPage := rope.Pages[lastPos]
	if rope.OwnsLast {
		return lastPage, lastPos
	}

	lastLen := len(lastPage)
	rope.OwnsLast = true
	newSize := min(remainingLen+lastLen, rope.PageSize)
	newPage := make([]T, lastLen, newSize)
	copy(newPage, lastPage)
	rope.Pages[lastPos] = newPage
	return newPage, lastPos
}

func (rope *ropeBuilder[T]) pushPage(data []T, owned bool) {
	if len(data) == 0 {
		panic("rope: pushing empty page")
	}
	rope.OwnsLast = owned

	count := 0
	if lastPos := len(rope.Count) - 1; lastPos >= 0 {
		count = rope.Count[lastPos]
	}

	rope.Pages = append(rope.Pages, data)
	rope.Count = append(rope.Count, count+len(data))
}
