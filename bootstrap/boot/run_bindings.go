package boot

import (
	"math"
	"slices"
	"sync"
)

type BindOrder int

type BindValue struct{}

type bindingMap struct {
	mutex   sync.Mutex
	sources bindingMapBySource
	globals map[bindingGlobal]bool
}

type bindingGlobal struct {
	Type Type
	Key  Key
	Val  BindValue
}

func (binds *bindingMap) BindSource(src *Source) {
	binds.mutex.Lock()
	defer binds.mutex.Unlock()
	if binds.sources == nil {
		binds.sources = make(bindingMapBySource)
	}

	if _, has := binds.sources[src]; !has {
		for global := range binds.globals {
			binds.doDefine(src, global.Type, global.Key, 0, math.MaxInt, global.Val)
		}
	}
}

func (binds *bindingMap) Define(typ Type, key Key, val BindValue) {
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

func (binds *bindingMap) DefineAt(span Span, typ Type, key Key, val BindValue) {
	binds.mutex.Lock()
	defer binds.mutex.Unlock()

	src, sta, end := span.Src(), span.Sta(), span.End()
	binds.doDefine(src, typ, key, sta, end, val)
}

func (binds *bindingMap) doDefine(src *Source, typ Type, key Key, sta, end int, val BindValue) {
	if src == nil || end < sta {
		panic("BindingMap: invalid range")
	} else if end == sta {
		return
	}
	bind := bindParams{Src: src, Type: typ, Key: key, Sta: sta, End: end, Val: val}
	if binds.sources == nil {
		binds.sources = make(bindingMapBySource)
	}
	binds.sources.BindSource(bind)
}

type bindParams struct {
	Src  *Source
	Type Type
	Key  Key
	Sta  int
	End  int
	Val  BindValue
}

type bindingMapBySource map[*Source]bindingMapByType

func (mSrc bindingMapBySource) BindSource(bind bindParams) {
	mTyp := mSrc[bind.Src]
	if mTyp == nil {
		mTyp = make(bindingMapByType)
		mSrc[bind.Src] = mTyp
	}
	mTyp.BindType(bind)
}

type bindingMapByType map[Type]bindingMapByKey

func (mTyp bindingMapByType) BindType(bind bindParams) {
	mKey := mTyp[bind.Type]
	if mKey == nil {
		mKey = make(bindingMapByKey)
		mTyp[bind.Type] = mKey
	}
	mKey.BindKey(bind)
}

type bindingMapByKey map[Key]*bindingMapSegments

func (mKey bindingMapByKey) BindKey(bind bindParams) {
	segs := mKey[bind.Key]
	if segs == nil {
		segs = &bindingMapSegments{}
		mKey[bind.Key] = segs
	}
	segs.BindSegment(bind)
}

type bindingMapSegments struct {
	list []*bindingSegment
}

type bindingSegment struct {
	params bindParams
	sta    int
	end    int
}

func (segs *bindingMapSegments) BindSegment(bind bindParams) {
	idxSta := SliceSkipParted(segs.list, func(it *bindingSegment) bool {
		return it.end <= bind.Sta
	})
	idxEnd := idxSta + SliceSkipParted(segs.list[idxSta:], func(it *bindingSegment) bool {
		return it.sta < bind.End
	})

	sta := bind.Sta
	end := bind.End

	if idxSta == idxEnd {
		segs.list = slices.Insert(segs.list, idxSta, &bindingSegment{
			params: bind,
			sta:    sta,
			end:    end,
		})
		return
	}

	panic("TODO")
}
