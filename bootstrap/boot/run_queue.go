package boot

import "container/heap"

type bindingQueue struct {
	heap []*bindingSegment
}

func (h *bindingQueue) Queue(segment *bindingSegment) {
	if !segment.queued {
		heap.Push(h, segment)
	} else {
		heap.Fix(h, segment.queuePos)
	}
}

func (h *bindingQueue) Dequeue() (out []*bindingSegment) {
	h.purgeSkipped()
	if len(h.heap) == 0 {
		return
	}

	head := h.heap[0]
	heap.Pop(h)
	out = append(out, head)
	h.purgeSkipped()
	for len(h.heap) > 0 {
		next := h.heap[0]
		same := next.IsSameGroup(head)
		if same {
			if out[len(out)-1].Cmp(next) == 0 {
				panic("duplicated segment in processing queue")
			}
			heap.Pop(h)
			out = append(out, next)
			h.purgeSkipped()
		} else {
			break
		}
	}

	return
}

func (h *bindingQueue) purgeSkipped() {
	for len(h.heap) > 0 && h.heap[0].skip {
		heap.Pop(h)
	}
}

func (h *bindingQueue) Len() int {
	return len(h.heap)
}

func (h *bindingQueue) Less(i, j int) bool {
	return h.heap[i].Cmp(h.heap[j]) < 0
}

func (h *bindingQueue) Swap(i, j int) {
	ls := h.heap
	ls[i], ls[j] = ls[j], ls[i]
	ls[i].queuePos = i
	ls[j].queuePos = j
}

func (h *bindingQueue) Push(x any) {
	segment := x.(*bindingSegment)
	segment.queuePos = len(h.heap)
	segment.queued = true
	h.heap = append(h.heap, segment)
}

func (h *bindingQueue) Pop() any {
	index := len(h.heap) - 1
	last := h.heap[index]
	last.queued = false
	last.queuePos = -1
	h.heap = h.heap[:index]
	return last
}
