package core

type RopeIter[T any] interface {
	Count(upTo int) int
	Next() []T
	Skip(count int)
}

func RopeIterChain[T any](iters ...RopeIter[T]) RopeIter[T] {
	return &ropeIterChain[T]{inner: iters}
}

type ropeIterChain[T any] struct {
	inner     []RopeIter[T]
	lastCount int
}

func (iter *ropeIterChain[T]) Count(upTo int) int {
	if upTo <= 0 {
		panic("RopeIter.Count: invalid upTo")
	}

	if iter.lastCount >= upTo {
		return iter.lastCount
	}

	iter.lastCount = 0
	for _, it := range iter.inner {
		iter.lastCount += it.Count(upTo - iter.lastCount)
		if iter.lastCount >= upTo {
			break
		}
	}

	return iter.lastCount
}

func (iter *ropeIterChain[T]) Next() []T {
	for len(iter.inner) > 0 {
		if iter.inner[0] == nil {
			iter.inner = iter.inner[1:]
			continue
		}
		next := iter.inner[0].Next()
		if len(next) == 0 {
			iter.inner = iter.inner[1:]
		} else {
			return next
		}
	}
	return nil
}

func (iter *ropeIterChain[T]) Skip(count int) {
	iter.lastCount = max(0, iter.lastCount-count)
	for count > 0 {
		next := iter.Next()
		if len(next) == 0 {
			panic("RopeIter.Skip: inner chained iterator returned empty")
		}
		skip := min(count, len(next))
		iter.inner[0].Skip(skip)
		count -= skip
	}
}

func RopeIterChunks[T any](chunks ...[]T) RopeIter[T] {
	return &ropeIterChunks[T]{chunks: chunks}
}

type ropeIterChunks[T any] struct {
	chunks    [][]T
	lastCount int
}

func (iter *ropeIterChunks[T]) Count(upTo int) int {
	if upTo <= 0 {
		panic("RopeIter.Count: invalid upTo")
	}

	if iter.lastCount >= upTo {
		return iter.lastCount
	}

	iter.lastCount = 0
	for _, it := range iter.chunks {
		iter.lastCount += len(it)
		if iter.lastCount >= upTo {
			break
		}
	}

	return iter.lastCount
}

func (iter *ropeIterChunks[T]) Next() []T {
	for len(iter.chunks) > 0 {
		if len(iter.chunks[0]) == 0 {
			iter.chunks = iter.chunks[1:]
		} else {
			return iter.chunks[0]
		}
	}
	return nil
}

func (iter *ropeIterChunks[T]) Skip(count int) {
	iter.lastCount = max(0, iter.lastCount-count)
	for count > 0 {
		if next := len(iter.chunks[0]); next <= count {
			iter.chunks = iter.chunks[1:]
			count -= next
		} else {
			iter.chunks[0] = iter.chunks[0][count:]
			count = 0
		}
	}
}
