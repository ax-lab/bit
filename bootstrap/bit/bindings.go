package bit

import (
	"cmp"
	"math"
	"slices"
	"sync"
	"sync/atomic"
)

type BindingMap struct {
	mutex sync.Mutex
	byKey map[Key]*segmentsByKey

	queue processQueue

	globalMutex   sync.Mutex
	globals       map[Key]Binding
	globalSources map[*Source]bool
}

func (segs *BindingMap) StepNext() bool {
	segments, nodes, requeue := segs.queue.Dequeue()
	if len(segments) == 0 {
		return false
	}

	for _, it := range segments {
		it.process.Store(false)
	}

	binding := segments[0].binding
	toQueue := binding.val.Process(binding, segments, nodes)
	requeue(toQueue)

	return true
}

func (segs *BindingMap) AddNodes(nodes ...*Node) {
	cur := 0
	for cur < len(nodes) {
		sta := cur
		cur++

		src := nodes[sta].Span().Source()
		key := nodes[sta].Key()
		if key == nil {
			continue
		}

		for cur < len(nodes) {
			if src != nodes[cur].Span().Source() {
				break
			}
			next := nodes[cur].Key()
			if next != nil && next.IsEqual(key) {
				cur++
			} else {
				break
			}
		}

		bySource := segs.getByKey(key).getBySource(src)
		bySource.addNodes(nodes[sta:cur]...)
	}
}

func (segs *BindingMap) InitSource(src *Source) {
	segs.globalMutex.Lock()
	defer segs.globalMutex.Unlock()

	if segs.globalSources == nil {
		segs.globalSources = make(map[*Source]bool)
	}
	if !segs.globalSources[src] {
		segs.globalSources[src] = true
		for key, binding := range segs.globals {
			segs.doBind(key, src.Span(), binding, true)
		}
	}
}

func (segs *BindingMap) BindGlobal(key Key, binding Binding) {
	segs.globalMutex.Lock()
	defer segs.globalMutex.Unlock()
	if segs.globals == nil {
		segs.globals = make(map[Key]Binding)
	}
	segs.globals[key] = binding

	for src := range segs.globalSources {
		segs.doBind(key, src.Span(), binding, true)
	}
}

func (segs *BindingMap) BindStatic(key Key, src *Source, binding Binding) {
	segs.Bind(key, src.Span(), binding)
}

func (segs *BindingMap) Bind(key Key, span Span, binding Binding) {
	segs.doBind(key, span, binding, false)
}

func (segs *BindingMap) doBind(key Key, span Span, binding Binding, global bool) {
	segments := segs.getByKey(key).getBySource(span.Source())
	sta, end := span.Sta(), span.End()
	if global {
		sta = 0
		end = math.MaxInt
	}
	segments.table.bind(&BindingValue{
		from:   segs,
		parent: segments,
		global: global,
		val:    binding,
		key:    key,
		src:    span.Source(),
		sta:    sta,
		end:    end,
	})
}

func (segs *BindingMap) getByKey(key Key) *segmentsByKey {
	segs.mutex.Lock()
	defer segs.mutex.Unlock()
	if segs.byKey == nil {
		segs.byKey = make(map[Key]*segmentsByKey)
	}

	data := segs.byKey[key]
	if data == nil {
		data = &segmentsByKey{
			parent: segs,
			key:    key,
		}
		segs.byKey[key] = data
	}

	return data
}

type Segment struct {
	binding  *BindingValue
	process  atomic.Bool
	queued   atomic.Bool
	queuePos atomic.Int64
	sta      int
	end      int
}

func (seg *Segment) IsDone() bool {
	return !seg.process.Load()
}

func (seg *Segment) Compare(other *Segment) int {
	if ord := cmp.Compare(seg.Precedence(), other.Precedence()); ord != 0 {
		return ord
	}

	if seg.IsGlobal() != other.IsGlobal() {
		if other.IsGlobal() {
			return -1
		} else {
			return +1
		}
	}

	if ord := seg.binding.src.Compare(other.binding.src); ord != 0 {
		return ord
	}

	if ord := cmp.Compare(seg.sta, other.sta); ord != 0 {
		return ord
	}

	return cmp.Compare(seg.end, other.end)
}

func (seg *Segment) IsGlobal() bool {
	return seg.binding.global
}

func (seg *Segment) Precedence() Precedence {
	return seg.binding.val.Precedence()
}

func (seg *Segment) skip() {
	seg.process.Store(false)
}

func (seg *Segment) enqueue() {
	seg.process.Store(true)
	seg.binding.from.queue.Queue(seg)
}

type BindingValue struct {
	from   *BindingMap
	parent *segmentsBySource
	global bool
	sta    int
	end    int
	val    Binding
	key    Key
	src    *Source
}

func (src *BindingValue) overrides(other *BindingValue) bool {
	return other.sta <= src.sta && src.end <= other.end
}

type segmentsByKey struct {
	mapMutex sync.Mutex
	parent   *BindingMap
	key      Key
	bySource map[*Source]*segmentsBySource
}

func (segs *segmentsByKey) getBySource(source *Source) *segmentsBySource {
	segs.mapMutex.Lock()
	defer segs.mapMutex.Unlock()
	if segs.bySource == nil {
		segs.bySource = make(map[*Source]*segmentsBySource)
	}

	data := segs.bySource[source]
	if data == nil {
		data = &segmentsBySource{
			parent: segs,
			source: source,
		}
		segs.bySource[source] = data
	}
	return data
}

type segmentsBySource struct {
	parent *segmentsByKey
	source *Source
	table  segmentTable

	nodesMutex  sync.Mutex
	nodesSorted bool
	nodes       []*Node
}

func (segs *segmentsBySource) addNodes(nodes ...*Node) {
	if len(nodes) == 0 {
		return
	}

	segs.nodesMutex.Lock()
	defer segs.nodesMutex.Unlock()

	segs.nodes = append(segs.nodes, nodes...)
	segs.nodesSorted = false
}

type segmentTable struct {
	mutex sync.RWMutex
	segs  []*Segment
}

func (table *segmentTable) bind(binding *BindingValue) {
	table.mutex.Lock()
	defer table.mutex.Unlock()

	sta, end := binding.sta, binding.end

	replaceSta, _ := findSegmentAt(sta, table.segs)
	replaceEnd := replaceSta
	if is_append := replaceSta >= len(table.segs); is_append {
		table.segs = append(table.segs, &Segment{
			binding: binding,
			sta:     binding.sta,
			end:     binding.end,
		})
		return
	}

	var new_segs []*Segment
	push_existing := func(n int, it *Segment) {
		replaceEnd = n + 1
		if len(new_segs) == 0 && replaceSta == n {
			replaceSta = replaceEnd
		} else {
			new_segs = append(new_segs, it)
		}
	}

	for n, it := range table.segs[replaceSta:] {
		cur := n + replaceSta
		if replaceEnd != cur {
			panic("invariant `replaceEnd == cur` failed")
		}

		if sta >= end || it.sta >= end {
			break
		} else if !binding.overrides(it.binding) {
			if has_gap_before := sta < it.sta; has_gap_before {
				new_segs = append(new_segs, &Segment{
					binding: binding,
					sta:     sta,
					end:     it.sta,
				})
			}
			sta = it.end
			push_existing(cur, it)
		} else {
			if overwrite := sta <= it.sta && it.end <= end; overwrite {
				replaceEnd = cur + 1
				continue
			}

			if keep_prefix := it.sta < sta; keep_prefix {
				it.end = sta
				push_existing(cur, it)
			} else if keep_suffix := end < it.end; keep_suffix {
				new_segs = append(new_segs, &Segment{
					binding: binding,
					sta:     sta,
					end:     end,
				})
				it.sta = end
				sta = end
				break
			} else {
				// new binding is right in the middle of an existing one
				push_existing(cur, it)
				new_segs = append(new_segs,
					&Segment{
						binding: binding,
						sta:     sta,
						end:     end,
					},
					&Segment{
						binding: it.binding,
						sta:     end,
						end:     it.end,
					},
				)
				it.end = sta
				sta = end
				break
			}
		}
	}

	if sta < end {
		new_segs = append(new_segs, &Segment{
			binding: binding,
			sta:     sta,
			end:     end,
		})
	}

	if len(new_segs) > 0 {
		var segs []*Segment
		segs = append(segs, table.segs[:replaceSta]...)
		segs = append(segs, new_segs...)
		segs = append(segs, table.segs[replaceEnd:]...)

		for _, it := range table.segs[replaceSta:replaceEnd] {
			it.skip()
		}

		for _, it := range new_segs {
			it.enqueue()
		}

		table.segs = segs
	}
}

func findNodeAt(offset int, nodes []*Node) (index int, found bool) {
	return slices.BinarySearchFunc(nodes, offset, func(node *Node, target int) int {
		if offset := node.Offset(); target == offset {
			return 0
		} else if target < offset {
			return +1
		} else {
			return -1
		}
	})
}

func findSegmentAt(offset int, segs []*Segment) (index int, found bool) {
	return slices.BinarySearchFunc(segs, offset, func(segment *Segment, target int) int {
		if segment.sta <= target && target < segment.end {
			return 0
		} else if target < segment.sta {
			return +1
		} else {
			return -1
		}
	})
}
