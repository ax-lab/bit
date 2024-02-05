package bit

import "container/heap"

const (
	PrecFirst Precedence = iota
	PrecLast
)

type Precedence int

type Binding interface {
	IsSame(other Binding) bool
	Precedence() Precedence
	Process(binding *BindingValue, segments []*Segment, nodes []*Node) (requeue []*Node)
	String() string
}

type processQueue struct {
	list []*Segment
}

func (h *processQueue) Queue(segment *Segment) {
	if segment.queued.CompareAndSwap(false, true) {
		heap.Push(h, segment)
	} else {
		heap.Fix(h, int(segment.queuePos.Load()))
	}
}

func (h *processQueue) Dequeue() (segs []*Segment, nodes []*Node, requeue func(nodes []*Node)) {
	h.skipDone()
	if len(h.list) == 0 {
		return
	}

	head := h.list[0]
	heap.Pop(h)
	segs = append(segs, head)
	h.skipDone()
	for len(h.list) > 0 {
		next := h.list[0]
		same := next.binding.src == head.binding.src && next.binding.parent == head.binding.parent && next.binding.val.IsSame(head.binding.val)
		if same {
			heap.Pop(h)
			segs = append(segs, next)
			h.skipDone()
		} else {
			break
		}
	}

	src := head.binding.parent
	src.nodesMutex.Lock()
	defer src.nodesMutex.Unlock()
	if !src.nodesSorted {
		SortNodes(src.nodes)
		src.nodesSorted = true
	}

	all := src.nodes
	src.nodes = nil

	for _, it := range segs {
		sta, _ := findNodeAt(it.sta, all)
		end, _ := findNodeAt(it.end, all[sta:])
		end += sta
		if end > sta {
			src.nodes = append(src.nodes, all[:sta]...)

			add := all[sta:end]
			nodes = append(nodes, add...)
			for _, it := range add {
				it.SetDone(true)
			}

			all = all[end:]
		}
	}

	src.nodes = append(src.nodes, all...)
	requeue = func(toRequeue []*Node) {
		cur := 0
		for _, it := range toRequeue {
			if !it.Done() {
				toRequeue[cur] = it
				cur++
			}
		}
		toRequeue = toRequeue[:cur]
		if len(toRequeue) == 0 {
			return
		}

		src.nodesMutex.Lock()
		defer src.nodesMutex.Unlock()
		src.nodesSorted = false
		src.nodes = append(src.nodes, toRequeue...)
	}

	return
}

func (h *processQueue) skipDone() {
	for len(h.list) > 0 && h.list[0].IsDone() {
		heap.Pop(h)
	}
}

func (h *processQueue) Len() int {
	return len(h.list)
}

func (h *processQueue) Less(i, j int) bool {
	return h.list[i].Compare(h.list[j]) < 0
}

func (h *processQueue) Swap(i, j int) {
	ls := h.list
	ls[i], ls[j] = ls[j], ls[i]
	ls[i].queuePos.Store(int64(i))
	ls[j].queuePos.Store(int64(j))
}

func (h *processQueue) Push(x any) {
	segment := x.(*Segment)
	segment.queuePos.Store(int64(len(h.list)))
	h.list = append(h.list, segment)
}

func (h *processQueue) Pop() any {
	index := len(h.list) - 1
	last := h.list[index]
	last.queuePos.Store(0)
	last.queued.Store(false)
	h.list = h.list[:index]
	return last
}
