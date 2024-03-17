package boot

import (
	"cmp"
	"math"
	"slices"
	"sync"
	"unsafe"
)

type BindOrder int

type Bind struct {
	inner *bindInner
}

func (bind Bind) Cmp(other Bind) int {
	va, vb := bind.inner, other.inner
	if va == vb {
		return 0
	}

	if res := cmp.Compare(va.order, vb.order); res != 0 {
		return res
	}

	pa := uintptr(unsafe.Pointer(va))
	pb := uintptr(unsafe.Pointer(vb))
	return cmp.Compare(pa, pb)
}

type BindArgs struct {
	Type Type
	Key  Key
	List []BindItem
}

type BindItem struct {
	src  *Source
	list [][2]int
}

func (br BindItem) Src() *Source {
	return br.src
}

func (br BindItem) Len() int {
	return len(br.list)
}

func (br BindItem) Get(index int) Span {
	sta := br.list[index][0]
	end := br.list[index][1]
	return Span{br.src, sta, end}
}

type bindInner struct {
	order BindOrder
	eval  func(BindArgs)
}

func BindFunc(order BindOrder, eval func(args BindArgs)) Bind {
	inner := &bindInner{order, eval}
	return Bind{inner}
}

func (val Bind) Eval(st *State, args BindArgs) {
	val.inner.eval(args)
}

type bindingMap struct {
	mutex   sync.Mutex
	sources bindingMapBySource
	globals map[bindingGlobal]bool
	queue   bindingQueue
}

func (st *State) Evaluate() {
	for {
		next := st.queue.Dequeue()
		if len(next) == 0 {
			break
		}

		info := next[0].info
		args := BindArgs{
			Type: info.typ,
			Key:  info.key,
		}

		for _, it := range next {
			sta := it.info.sta
			end := it.info.end
			src := it.info.src
			if last := len(args.List) - 1; last >= 0 && args.List[last].src == src {
				item := &args.List[last]
				item.list = append(item.list, [2]int{sta, end})
			} else {
				item := BindItem{src: src, list: [][2]int{{sta, end}}}
				args.List = append(args.List, item)
			}
		}

		info.val.Eval(st, args)
	}
}

// Compare the relative processing order of two segments.
func (seg *bindingSegment) Cmp(other *bindingSegment) int {
	sa, sb := seg, other
	va, vb := sa.info, sb.info

	if res := va.val.Cmp(vb.val); res != 0 {
		return res
	}

	if res := va.typ.Cmp(vb.typ); res != 0 {
		return res
	}

	if res := va.key.Cmp(vb.key); res != 0 {
		return res
	}

	if res := va.src.Cmp(vb.src); res != 0 {
		return res
	}

	if res := cmp.Compare(sa.sta, sb.sta); res != 0 {
		return res
	}

	if res := cmp.Compare(sa.end, sb.end); res != 0 {
		return res
	}

	return 0
}

func (seg *bindingSegment) IsSameGroup(other *bindingSegment) bool {
	a, b := seg.info, other.info

	return a.val.Cmp(b.val) == 0 &&
		a.typ.Cmp(b.typ) == 0 &&
		a.key.Cmp(b.key) == 0
}

type bindingGlobal struct {
	Type Type
	Key  Key
	Val  Bind
}

func (binds *bindingMap) BindSource(src *Source) {
	binds.mutex.Lock()
	defer binds.mutex.Unlock()
	if binds.sources == nil {
		binds.sources = make(bindingMapBySource)
	}

	if _, has := binds.sources[src]; !has {
		binds.sources[src] = make(bindingMapByType)
		for global := range binds.globals {
			binds.doDefine(src, global.Type, global.Key, 0, math.MaxInt, global.Val)
		}
	}
}

func (binds *bindingMap) Define(typ Type, key Key, val Bind) {
	binds.mutex.Lock()
	defer binds.mutex.Unlock()
	if binds.globals == nil {
		binds.globals = make(map[bindingGlobal]bool)
	}

	global := bindingGlobal{typ, key, val}
	if !binds.globals[global] {
		binds.globals[global] = true
		for src := range binds.sources {
			binds.doDefine(src, typ, key, 0, math.MaxInt, val)
		}
	}
}

func (binds *bindingMap) DefineAt(span Span, typ Type, key Key, val Bind) {
	binds.mutex.Lock()
	defer binds.mutex.Unlock()

	src, sta, end := span.Src(), span.Sta(), span.End()
	binds.doDefine(src, typ, key, sta, end, val)
}

func (binds *bindingMap) doDefine(src *Source, typ Type, key Key, sta, end int, val Bind) {
	if src == nil || end < sta {
		panic("BindingMap: invalid range")
	} else if end == sta {
		return
	}
	bind := bindingInfo{src: src, typ: typ, key: key, sta: sta, end: end, val: val}
	if binds.sources == nil {
		binds.sources = make(bindingMapBySource)
	}
	binds.sources.BindSource(binds, bind)
}

type bindingInfo struct {
	src *Source
	typ Type
	key Key
	sta int
	end int
	val Bind
}

type bindingMapBySource map[*Source]bindingMapByType

func (mSrc bindingMapBySource) BindSource(parent *bindingMap, bind bindingInfo) {
	mTyp := mSrc[bind.src]
	if mTyp == nil {
		mTyp = make(bindingMapByType)
		mSrc[bind.src] = mTyp
	}
	mTyp.BindType(parent, bind)
}

type bindingMapByType map[Type]bindingMapByKey

func (mTyp bindingMapByType) BindType(parent *bindingMap, bind bindingInfo) {
	mKey := mTyp[bind.typ]
	if mKey == nil {
		mKey = make(bindingMapByKey)
		mTyp[bind.typ] = mKey
	}
	mKey.BindKey(parent, bind)
}

type bindingMapByKey map[Key]*bindingMapSegments

func (mKey bindingMapByKey) BindKey(parent *bindingMap, bind bindingInfo) {
	segs := mKey[bind.key]
	if segs == nil {
		segs = &bindingMapSegments{}
		mKey[bind.key] = segs
	}
	segs.BindSegment(parent, bind)
}

type bindingMapSegments struct {
	list []*bindingSegment
}

type bindingSegment struct {
	info     bindingInfo
	skip     bool
	queuePos int
	queued   bool
	sta      int
	end      int
}

func (seg *bindingSegment) ShouldOverride(bind *bindingInfo) bool {
	return seg.info.sta <= bind.sta && bind.end <= seg.info.end
}

func (segs *bindingMapSegments) BindSegment(parent *bindingMap, bind bindingInfo) {
	idxSta := SliceSkipParted(segs.list, func(it *bindingSegment) bool {
		return it.end <= bind.sta
	})
	idxEnd := idxSta + SliceSkipParted(segs.list[idxSta:], func(it *bindingSegment) bool {
		return it.sta < bind.end
	})

	sta := bind.sta
	end := bind.end

	if idxSta == idxEnd {
		it := &bindingSegment{
			info: bind,
			sta:  sta,
			end:  end,
		}
		segs.list = slices.Insert(segs.list, idxSta, it)
		parent.queue.Queue(it)
		return
	}

	for _, it := range segs.list[idxSta:idxEnd] {
		it.skip = true
	}

	var newList, toQueue []*bindingSegment

	push := func(toEnd int) {
		if sta >= toEnd {
			return
		}

		if len(newList) > 0 {
			last := newList[len(newList)-1]
			if last.end == sta && last.info == bind {
				last.end = toEnd
			}
		} else {
			new := &bindingSegment{
				info: bind,
				sta:  sta,
				end:  toEnd,
			}
			newList = append(newList, new)
			toQueue = append(toQueue, new)
		}
		sta = toEnd
	}

	for _, it := range segs.list[idxSta:idxEnd] {
		if sta >= end {
			panic("BindSegment: resulted in invalid segment range")
		}
		if it.sta >= end || it.end <= sta {
			panic("BindSegment: existing segment is out of range")
		}

		if it.info == bind {
			it.sta = min(it.sta, sta)
			newList = append(newList, it)
			end = it.end
		} else if it.ShouldOverride(&bind) {
			if it.sta < sta && it.end > end {
				suffix := &bindingSegment{
					info: it.info,
					sta:  end,
					end:  it.end,
				}
				it.end = sta
				newList = append(newList, it)
				push(end)
				newList = append(newList, suffix)
			} else if it.sta < sta {
				it.end = sta
				newList = append(newList, it)
			} else if it.end > end {
				push(end)
				it.sta = end
				newList = append(newList, it)
			}
		} else {
			push(min(end, it.sta))
			newList = append(newList, it)
			sta = it.end
		}
	}
	push(end)

	pre := segs.list[:idxSta]
	pos := segs.list[idxEnd:]
	cnt := len(pre) + len(pos) + len(newList)
	if len := len(segs.list); cnt > len {
		segs.list = append(segs.list, make([]*bindingSegment, cnt-len)...)
	}

	for _, it := range newList {
		it.skip = false
	}

	copy(segs.list[len(newList)+idxEnd:], pos)
	copy(segs.list[idxSta:], newList)

	for _, it := range toQueue {
		parent.queue.Queue(it)
	}
}
