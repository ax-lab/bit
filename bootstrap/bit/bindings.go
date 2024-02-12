package bit

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	queueCount = 0
)

type BindingMap struct {
	mutex sync.Mutex
	byKey map[Key]*segmentsByKey

	program *Program

	queue processQueue

	globalMutex   sync.Mutex
	globals       map[Key]Binding
	globalSources map[*Source]bool
}

func (segs *BindingMap) StepNext() bool {
	for {
		segments, nodes := segs.queue.Dequeue()
		if len(segments) == 0 {
			return false
		}

		if debugQueue {
			queueCount += 1
			header := fmt.Sprintf("Queue %d - ", queueCount)
			DebugNodes(header+" Nodes", nodes...)
			DebugSegments(header+" Segments", segments...)
		}

		for _, it := range segments {
			it.pending.Store(false)
		}

		if len(nodes) == 0 {
			continue
		}

		SortNodes(nodes)

		binding := segments[0].binding
		args := BindArgs{
			Program:  segs.program,
			Segments: segments,
			Nodes:    nodes,
		}
		binding.val.Process(&args)
		return true
	}
}

func (segs *BindingMap) AddNodes(key Key, nodes ...*Node) {
	cur := 0
	for cur < len(nodes) {
		sta := cur
		cur++

		src := nodes[sta].Span().Source()
		for cur < len(nodes) {
			if src != nodes[cur].Span().Source() {
				break
			}
			cur++
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
			segs.doBindSpan(key, src.Span(), binding, true)
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
		segs.doBindSpan(key, src.Span(), binding, true)
	}
}

func (segs *BindingMap) BindStatic(key Key, src *Source, binding Binding) {
	segs.Bind(key, src.Span(), binding)
}

func (segs *BindingMap) Bind(key Key, span Span, binding Binding) {
	segs.doBindSpan(key, span, binding, false)
}

func (segs *BindingMap) BindAt(key Key, sta, end int, src *Source, binding Binding) {
	segs.doBind(key, sta, end, src, binding, false)
}

func (segs *BindingMap) Dump() string {
	out := strings.Builder{}

	var keys []Key
	for k := range segs.byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(a, b int) bool {
		return keys[a].Repr(true) < keys[b].Repr(true)
	})

	for n, key := range keys {
		if n > 0 {
			out.WriteString("\n")
		}
		out.WriteString(fmt.Sprintf(">>> KEY = %s:\n", key.Repr(true)))

		byKey := segs.byKey[key]

		var sources []*Source
		for src := range byKey.bySource {
			sources = append(sources, src)
		}

		sort.Slice(sources, func(a, b int) bool {
			return sources[a].Compare(sources[b]) < 0
		})

		for _, src := range sources {
			out.WriteString(fmt.Sprintf("\n\t--> %s:\n", src.Name()))

			bySrc := byKey.bySource[src]

			out.WriteString("\n\t\tNodes {\n")
			for n, it := range bySrc.nodes {
				out.WriteString(fmt.Sprintf("\t\t\t[%03d] %s#%d  @%s", n, it.Value().Repr(true), it.Id(), it.Span().String()))
				if txt := it.Span().DisplayText(0); len(txt) > 0 {
					out.WriteString("  # ")
					out.WriteString(txt)
				}
				out.WriteString(fmt.Sprintf("  --  %d\n", it.Offset()))
			}
			out.WriteString("\t\t}\n")

			out.WriteString("\n\t\tSegments{\n")

			for n, seg := range bySrc.table.segs {
				var end string
				if seg.end == math.MaxInt {
					end = "MAX"
				} else {
					end = fmt.Sprint(seg.end)
				}
				out.WriteString(fmt.Sprintf("\t\t\t[%03d] %d..%s = ", n, seg.sta, end))
				binding := seg.binding
				if binding.global {
					out.WriteString("(GLOBAL) ")
				} else {
					out.WriteString(fmt.Sprintf("(%d..%d) ", binding.sta, binding.end))
				}
				out.WriteString(binding.val.String())
				if !seg.pending.Load() {
					out.WriteString(" [DONE]")
				}
				out.WriteString("\n")
			}
			out.WriteString("\t\t}\n")
		}
	}

	return out.String()
}

func (segs *BindingMap) doBindSpan(key Key, span Span, binding Binding, global bool) {
	sta, end, src := span.Sta(), span.End(), span.Source()
	segs.doBind(key, sta, end, src, binding, global)
}

func (segs *BindingMap) doBind(key Key, sta, end int, src *Source, binding Binding, global bool) {
	segments := segs.getByKey(key).getBySource(src)
	if global {
		sta = 0
		end = math.MaxInt
	}
	segments.table.bind(&BindingValue{
		from:   segs,
		parent: segments,
		global: global,
		val:    binding,
		repr:   binding.String(),
		key:    key,
		src:    src,
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
	pending  atomic.Bool
	queued   atomic.Bool
	queuePos atomic.Int64
	sta      int
	end      int
}

func (seg *Segment) IsDone() bool {
	pending := seg.pending.Load()
	return !pending
}

func (seg *Segment) Compare(other *Segment) int {
	// precedence has the precedence
	if ord := cmp.Compare(seg.Precedence(), other.Precedence()); ord != 0 {
		return ord
	}

	// use the string repr for the binding value as an easy way to sort and
	// keep segments for the same binding operator together
	if ord := cmp.Compare(seg.binding.repr, other.binding.repr); ord != 0 {
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

func (seg *Segment) takeNodes() (out []*Node) {
	src := seg.binding.parent
	src.nodesMutex.Lock()
	defer src.nodesMutex.Unlock()
	if !src.nodesSorted {
		SortNodes(src.nodes)
		src.nodesSorted = true
	}

	sta, _ := findNodeAt(seg.sta, src.nodes)
	end, _ := findNodeAt(seg.end, src.nodes[sta:])
	if end > 0 {
		end += sta
		all := src.nodes
		src.nodes = nil
		src.nodes = append(src.nodes, all[:sta]...)

		for _, it := range all[sta:end] {
			if it.done.CompareAndSwap(false, true) {
				out = append(out, it)
			}
		}

		src.nodes = append(src.nodes, all[end:]...)
	}
	return out
}

func (seg *Segment) skip() {
	seg.pending.Store(false)
}

func (seg *Segment) enqueue() {
	seg.pending.Store(true)
	seg.binding.from.queue.Queue(seg)
}

type BindingValue struct {
	from   *BindingMap
	parent *segmentsBySource
	global bool
	sta    int
	end    int
	val    Binding
	repr   string
	key    Key
	src    *Source
}

func (bind *BindingValue) overrides(other *BindingValue) bool {
	return other.sta <= bind.sta && bind.end <= other.end
}

func (bind *BindingValue) String() string {
	out := strings.Builder{}
	out.WriteString("Binding(")
	if bind.global {
		out.WriteString("GLOBAL")
	} else {
		out.WriteString(fmt.Sprintf("%d..%d", bind.sta, bind.end))
	}

	out.WriteString(" / ")
	out.WriteString(bind.key.Repr(true))
	out.WriteString(" = ")
	out.WriteString(bind.val.String())
	out.WriteString(" @")
	out.WriteString(bind.src.Name())
	out.WriteString(")")
	return out.String()
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

	SortNodes(nodes)

	segs.nodesMutex.Lock()
	defer segs.table.revive(nodes)
	defer segs.nodesMutex.Unlock()

	sorted := len(segs.nodes) == 0 || (segs.nodesSorted && nodes[0].Offset() >= segs.nodes[len(segs.nodes)-1].Offset())

	segs.nodes = append(segs.nodes, nodes...)
	segs.nodesSorted = sorted
}

type segmentTable struct {
	mutex sync.RWMutex
	segs  []*Segment
}

func (table *segmentTable) revive(nodes []*Node) {
	curIdx, curOff := 0, 0
	for _, it := range nodes {
		off := it.Offset()
		if off < curOff {
			continue
		}

		if idx, ok := findSegmentAt(off, table.segs[curIdx:]); ok {
			idx += curIdx
			seg := table.segs[idx]
			seg.enqueue()
			curIdx = idx + 1
			curOff = seg.end
		}
	}
}

func (table *segmentTable) bind(binding *BindingValue) {
	table.mutex.Lock()
	defer table.mutex.Unlock()

	sta, end := binding.sta, binding.end

	replaceSta, _ := findSegmentAt(sta, table.segs)
	replaceEnd := replaceSta
	if is_append := replaceSta >= len(table.segs); is_append {
		new_seg := &Segment{
			binding: binding,
			sta:     binding.sta,
			end:     binding.end,
		}
		table.segs = append(table.segs, new_seg)
		new_seg.enqueue()
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

func DebugSegments(msg string, segments ...*Segment) {
	out := strings.Builder{}
	out.WriteString(msg)
	for _, it := range segments {
		var end string
		if it.end == math.MaxInt {
			end = "MAX"
		} else {
			end = fmt.Sprint(it.end)
		}
		out.WriteString(fmt.Sprintf("\n\n    @ %d..%s\n    = %s\n", it.sta, end, it.binding.String()))
	}

	if len(segments) == 0 {
		out.WriteString("  (no segments)\n")
	}
	fmt.Println(out.String())
}
