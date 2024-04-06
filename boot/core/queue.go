package core

import "fmt"

type Queue[T any] struct {
	list []T
	next int
	size int
}

func QueueNew[T any](list ...T) Queue[T] {
	return Queue[T]{list, 0, len(list)}
}

func (queue *Queue[T]) Len() int {
	return queue.size
}

func (queue *Queue[T]) Peek() (out T, ok bool) {
	if queue.size > 0 {
		return queue.list[queue.next], true
	}
	return out, false
}

func (queue *Queue[T]) Shift() (out T, ok bool) {
	if queue.size > 0 {
		var zero T
		out, ok = queue.list[queue.next], true
		queue.list[queue.next] = zero
		queue.next = (queue.next + 1) % len(queue.list)
		queue.size--
	}
	return
}

func (queue *Queue[T]) Push(elems ...T) {
	new := len(elems)
	if new == 0 {
		return
	}

	assertFail := func(msg string, copied int) {
		panic(fmt.Sprintf("Queue: %s (next=%d, size=%d, elems=%d, cap=%d, copied=%d)",
			msg, queue.next, queue.size, len(elems), len(queue.list), copied))
	}

	queueCap := len(queue.list)
	queueLen := queue.size
	if cap := queueCap - queueLen; cap >= new {
		var copied int

		sta := 0
		end := queue.next + queueLen
		if end < queueCap {
			copied += copy(queue.list[end:], elems)
		} else {
			sta = end - queueCap
		}

		copied += copy(queue.list[sta:queue.next], elems[copied:])
		if copied < len(elems) {
			assertFail("failed to push elems, not enough capacity", copied)
		}

		queue.size += len(elems)
		return
	}

	newLen := queueLen + len(elems)
	newCap := max(len(queue.list)*2, 8, newLen)
	newBuf := make([]T, newCap)

	var copied int
	if queueLen > 0 {
		end := queue.next + queue.size
		copied += copy(newBuf, queue.list[queue.next:min(queueCap, end)])

		sta := 0
		if end > queueCap {
			sta = end - queueCap
		}
		copied += copy(newBuf[copied:], queue.list[0:sta])
	}

	if copied != queueLen {
		assertFail("resize failed to copy existing elements", copied)
	}

	copied += copy(newBuf[queueLen:], elems)
	if copied != newLen {
		assertFail("resize failed to copy all elements", copied)
	}

	queue.next = 0
	queue.size = newLen
	queue.list = newBuf
}
