package boot

import (
	"cmp"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type Type struct {
	inner *typeInner
}

func TypeNew() Type {
	inner := &typeInner{done: typeDone}
	out := Type{inner}
	return out
}

func TypeFromValue(val any) Type {
	typ := reflect.TypeOf(val)
	if typ == nil {
		return Type{}
	}
	return TypeFrom(typ)
}

func TypeFrom(typ reflect.Type) Type {
	if val, ok := typeMap.Load(typ); ok {
		inner := val.(*typeInner)
		<-inner.done
		return Type{inner}
	}

	inner := &typeInner{done: make(chan struct{})}
	if val, loaded := typeMap.LoadOrStore(typ, inner); loaded {
		inner = val.(*typeInner)
		<-inner.done
	} else {
		defer close(inner.done)
		if name := typ.Name(); name != "" {
			pkg := strings.Replace(typ.PkgPath(), "axlab.dev/bit/", "", -1)
			if pkg != "" {
				inner.name = fmt.Sprintf("%s.%s", pkg, name)
			} else {
				inner.name = name
			}
		}
	}
	return Type{inner}
}

func TypeOf[T any]() Type {
	var zero T
	typ := reflect.TypeOf(zero)
	return TypeFrom(typ)
}

func (typ Type) Name() string {
	if typ.inner == nil {
		return ""
	}
	return typ.inner.name
}

func (typ Type) String() string {
	if typ.inner == nil {
		return "Type(nil)"
	}
	if name := typ.Name(); name != "" {
		return fmt.Sprintf("Type(%s)", name)
	}
	return fmt.Sprintf("Type(%p)", typ.inner)
}

func (typ Type) Cmp(other Type) int {
	if res := cmp.Compare(typ.Name(), other.Name()); res != 0 {
		return res
	}

	a := uintptr(unsafe.Pointer(typ.inner))
	b := uintptr(unsafe.Pointer(other.inner))
	return cmp.Compare(a, b)
}

var (
	typeMap  sync.Map
	typeDone chan struct{} = (func() chan struct{} {
		out := make(chan struct{})
		close(out)
		return out
	})()
)

type typeInner struct {
	done chan struct{}
	name string
}
