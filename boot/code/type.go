package code

import (
	"sync"
	"sync/atomic"
)

type Type struct {
	data *typeData
}

func (typ Type) String() string {
	return typ.Def().String()
}

type typeData struct {
	set *TypeSet
	def TypeDef
}

func (typ Type) Def() TypeDef {
	return typ.data.def
}

type TypeDef interface {
	TypeDef() TypeDef
	String() string
}

type TypeSet struct {
	source atomic.Pointer[Program]

	keys typeKeyLookup

	scalarSync sync.Mutex
	scalarMap  map[TypeScalarKind]Type

	tupleSync sync.Mutex
	tupleMap  map[TypeKey]Type
}

func (set *TypeSet) Program() *Program {
	return set.source.Load()
}

func (set *TypeSet) GetKey(types ...Type) TypeKey {
	return set.keys.Get(types, 0)
}

func (set *TypeSet) Scalar(kind TypeScalarKind) Type {
	set.scalarSync.Lock()
	defer set.scalarSync.Unlock()

	typ, ok := set.scalarMap[kind]
	if !ok {
		typ = set.newType(TypeScalar{kind})
		if set.scalarMap == nil {
			set.scalarMap = make(map[TypeScalarKind]Type)
		}
		set.scalarMap[kind] = typ
	}

	return typ
}

func (set *TypeSet) Tuple(types ...Type) Type {
	key := set.GetKey(types...)

	set.tupleSync.Lock()
	defer set.tupleSync.Unlock()

	typ, ok := set.tupleMap[key]
	if !ok {
		typ = set.newType(TypeTuple{key.Types()})
		if set.tupleMap == nil {
			set.tupleMap = make(map[TypeKey]Type)
		}
		set.tupleMap[key] = typ
	}

	return typ
}

func (set *TypeSet) newType(def TypeDef) Type {
	data := &typeData{
		set: set,
		def: def,
	}
	return Type{data}
}
