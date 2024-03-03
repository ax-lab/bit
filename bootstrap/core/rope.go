package core

const (
	ropeSizePage  = 256
	ropeSizeBlock = 256
)

type Rope[T any] struct {
	buffer *ropeBuffer[T]
}

func (rope Rope[T]) Get(n int) T {
	if rope.buffer != nil {
		index, valid := rope.buffer.PageIndex(n)
		if valid {
			return rope.buffer.Blocks[index[0]].Pages[index[1]][index[2]]
		}
	}
	panic("rope: invalid index for get")
}

func (rope Rope[T]) Slice(sta, end int) Rope[T] {
	panic("TODO")
}

func (rope Rope[T]) Splice(sta, end int, blocks ...[]T) Rope[T] {
	if rope.buffer == nil {
		rope.buffer = &ropeBuffer[T]{}
	} else {
		newBuffer := &ropeBuffer[T]{}
		*newBuffer = *rope.buffer
	}
	rope.buffer.Splice(sta, end, [2]int{ropeSizeBlock, ropeSizePage}, RopeIterChunks(blocks...))
	return rope
}

func (rope *Rope[T]) Set(n int, v T) Rope[T] {
	return rope.Splice(n, n, []T{v})
}

func (rope *Rope[T]) Remove(sta, end int) Rope[T] {
	return rope.Splice(sta, end)
}

func (rope *Rope[T]) Insert(at int, values ...T) Rope[T] {
	return rope.Splice(at, at, values)
}
