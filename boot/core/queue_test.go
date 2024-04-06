package core_test

import (
	"testing"

	"axlab.dev/bit/boot/core"
	"github.com/stretchr/testify/require"
)

func TestQueueEmpty(t *testing.T) {
	test := require.New(t)
	queue := core.Queue[string]{}

	test.Zero(queue.Len(), "empty queue len should be zero")
	_, peekOk := queue.Peek()
	test.False(peekOk, "peek empty queue should return false")

	_, shiftOk := queue.Shift()
	test.False(shiftOk, "shift empty queue should return false")
}

func TestQueueInit(t *testing.T) {
	test := require.New(t)

	q1 := core.QueueNew(1, 2, 3, 4)
	for i := 1; i <= 4; i++ {
		it, ok := q1.Shift()
		test.True(ok)
		test.Equal(i, it)
	}
	_, ok1 := q1.Shift()
	test.False(ok1)

	q2 := core.QueueNew(1, 2, 3)
	q2.Push(4, 5, 6)
	for i := 1; i <= 6; i++ {
		it, ok := q2.Shift()
		test.True(ok)
		test.Equal(i, it)
	}
	_, ok2 := q2.Shift()
	test.False(ok2)
}

func TestQueue(t *testing.T) {
	testQueueChunk(t, 1997, 2, 1)
	testQueueChunk(t, 1997, 4, 2)
	testQueueChunk(t, 1997, 10, 5)
	testQueueChunk(t, 1997, 13, 7)
	testQueueChunk(t, 1997, 100, 100)
}

func testQueueChunk(t *testing.T, elems int, chunkPush, chunkShift int) {
	test := require.New(t)

	queue := core.Queue[int]{}
	expected, next, expectedLen := 0, 0, 0

	drain := func(count int) {
		for count > 0 {
			count--

			test.Equal(expectedLen, queue.Len(), "expected queue len, before shift")
			if expectedLen == 0 {
				break
			}

			peek, ok := queue.Peek()
			test.True(ok, "expected peek to succeed")
			test.Equal(expected, peek, "unexpected peeked item")

			shift, ok := queue.Shift()
			test.True(ok, "expected shift to succeed")
			test.Equal(expected, shift, "unexpected shifted item")

			expected++
			expectedLen--
			test.Equal(expectedLen, queue.Len(), "expected queue len, after shift")
		}
	}

	for next < elems {
		test.Equal(expectedLen, queue.Len(), "expected queue len, before push")
		queue.Push(next)

		next++
		expectedLen++
		test.Equal(expectedLen, queue.Len(), "expected queue len, after push")

		if next%chunkPush == 0 {
			drain(chunkShift)
		}
	}

	if expectedLen > 0 {
		drain(expectedLen)
	}

	test.Zero(queue.Len(), "queue not empty after test")
}
