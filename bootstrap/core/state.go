package core

import (
	"math"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"unsafe"
)

type State struct {
	inner *stateData
}

func NewState() State {
	return State{&stateData{}}
}

func (st State) Clone() State {
	ver := st.inner.Acquire()
	defer st.inner.Release(ver)

	st.inner.edit.Add(1) // force copy-on-write for shared state

	out := State{&stateData{}}
	out.inner.sets = slices.Clone(st.inner.sets)
	return out
}

func (st State) Merge(to State) (out State, ok bool) {
	v1 := st.inner.Acquire()
	v2 := to.inner.Acquire()
	defer st.inner.Release(v1)
	defer to.inner.Release(v2)
	panic("TODO")
}

type Value[T any] struct {
	_     [0]T
	idVal uint64
	idTyp uint32
}

func (v Value[T]) Valid() bool {
	return v.idVal > 0
}

func (v Value[T]) Get(st State) (out T) {
	if v.idVal == 0 || v.idTyp == 0 {
		panic("Value is invalid")
	}

	id := st.inner.Acquire()
	defer st.inner.Release(id)

	idxTyp := int(v.idTyp) - 1
	idxVal := int(v.idVal) - 1
	if idxTyp < len(st.inner.sets) {
		if set := (*stateSet[T])(st.inner.sets[idxTyp]); set != nil {
			index := idxVal / dataSetPageSize
			offset := idxVal % dataSetPageSize
			if page := set.pages[index]; page != nil {
				out = page.data[offset]
			}
		}
	}
	return out
}

func (v Value[T]) Set(st State, newValue T) {
	if v.idVal == 0 || v.idTyp == 0 {
		panic("Value is invalid")
	}

	id := st.inner.Acquire()
	defer st.inner.Release(id)

	inner := st.inner

	idxTyp := int(v.idTyp) - 1
	idxVal := int(v.idVal) - 1
	if len(inner.sets) <= idxTyp {
		size := idxTyp + 1
		inner.sets = slices.Grow(inner.sets, size-len(inner.sets))
		inner.sets = inner.sets[:cap(inner.sets)]
	}

	set := (*stateSet[T])(inner.sets[idxTyp])
	if set == nil || !set.CanWrite(inner) {
		newSet := &stateSet[T]{}
		inner.Own(&newSet.stateHandle)
		if set != nil {
			newSet.pages = append(newSet.pages, set.pages...)
		}
		inner.sets[idxTyp] = unsafe.Pointer(newSet)
		set = newSet
	}

	index := idxVal / dataSetPageSize
	offset := idxVal % dataSetPageSize
	if len(set.pages) <= index {
		size := index + 1
		set.pages = slices.Grow(set.pages, size-len(set.pages))
		set.pages = set.pages[:cap(set.pages)]
	}

	page := set.pages[index]
	if page == nil || !page.CanWrite(inner) {
		newPage := &stateSetPage[T]{}
		inner.Own(&newPage.stateHandle)
		if page != nil {
			newPage.data = page.data
		}
		set.pages[index] = newPage
		page = newPage
	}

	page.data[offset] = newValue
}

func New[T any]() Value[T] {
	typ := typeInfoOf[T]()
	val := typ.inc.Add(1)
	if val > valueMax || val > math.MaxInt {
		panic("Value: internal ID overflow")
	}
	return Value[T]{idVal: val, idTyp: typ.id}
}

var (
	typeIds  sync.Map
	typeCnt  atomic.Uint32
	typeLast atomic.Pointer[typeInfo]
)

const (
	typeIdBits uint32 = 16
	typeIdMax  uint32 = ^uint32(0) >> (32 - typeIdBits)
	valueMax   uint64 = ^uint64(0) >> typeIdBits
)

type typeInfo struct {
	m   sync.Mutex
	id  uint32
	inc atomic.Uint64
	typ reflect.Type
}

func TypeId[T any]() uint32 {
	return typeInfoOf[T]().id
}

func typeInfoOf[T any]() *typeInfo {
	var val T
	typ := reflect.TypeOf(val)

	if last := typeLast.Load(); last != nil && last.typ == typ {
		return last
	}

	var info *typeInfo
	if v, ok := typeIds.Load(typ); ok {
		info = v.(*typeInfo)
	} else {
		v, _ := typeIds.LoadOrStore(typ, &typeInfo{typ: typ})
		info = v.(*typeInfo)
	}

	if info.id == 0 {
		info.m.Lock()
		info.id = typeCnt.Add(1)
		if info.id > typeIdMax {
			panic("TypeId overflow")
		}
		info.m.Unlock()
	}

	typeLast.Store(info)

	return info
}

type stateData struct {
	lock atomic.Uint64
	edit atomic.Uint64
	sets []unsafe.Pointer
}

func (sd *stateData) Acquire() uint64 {
	if sd == nil {
		panic("State is uninitialized")
	}
	return sd.lock.Load()
}

func (sd *stateData) Release(v uint64) {
	if !sd.lock.CompareAndSwap(v, v+1) {
		panic("State: concurrent reading or writing detected")
	}
}

type stateHandle struct {
	data *stateData
	edit uint64
}

func (st *stateData) Own(own *stateHandle) {
	own.data = st
	own.edit = st.edit.Load()
}

func (own *stateHandle) CanWrite(data *stateData) bool {
	return own.data == data && own.edit == data.edit.Load()
}

const stateSetPageSize = 1024

type stateSet[T any] struct {
	stateHandle
	pages []*stateSetPage[T]
}

type stateSetPage[T any] struct {
	stateHandle
	data [stateSetPageSize]T
}
