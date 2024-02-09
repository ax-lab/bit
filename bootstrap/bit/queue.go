package bit

import (
	"container/heap"
	"sort"
)

type BindArgs struct {
	Program  *Program
	Segments []*Segment
	Nodes    []*Node
	Requeue  []*Node
}

func (args *BindArgs) NodesByParent() (out [][]*Node) {
	set := make(map[*Node][]*Node)
	for _, it := range args.Nodes {
		par := it.Parent()
		if par == nil {
			continue
		}
		set[par] = append(set[par], it)
	}

	for _, v := range set {
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i][0].Parent().Compare(out[j][0].Parent()) < 0
	})

	return
}

func (args *BindArgs) ParentNodes() (out []*Node) {
	set := make(map[*Node][]*Node)
	for _, it := range args.Nodes {
		par := it.Parent()
		set[par] = append(set[par], it)
	}

	out = make([]*Node, 0, len(set))
	for it := range set {
		out = append(out, it)
	}

	SortNodes(out)
	return
}

func (args *BindArgs) RequeueNodes() {
	for _, it := range args.Nodes {
		it.Undo()
	}
}

type Binding interface {
	IsSame(other Binding) bool
	Precedence() Precedence
	Process(args *BindArgs)
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

func (h *processQueue) Dequeue() (segs []*Segment, nodes []*Node) {
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
		same := next.binding.src == head.binding.src && next.binding.val.IsSame(head.binding.val)
		if same {
			heap.Pop(h)
			segs = append(segs, next)
			h.skipDone()
		} else {
			break
		}
	}

	for _, it := range segs {
		add := it.takeNodes()
		nodes = append(nodes, add...)
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
