package boot

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Type struct {
	inner *typeInner
}

func NewType() Type {
	inner := &typeInner{done: typeDone}
	out := Type{inner}
	return out
}

func TypeOf[T any]() Type {
	var zero T
	typ := reflect.TypeOf(zero)
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

func (typ Type) Name() string {
	return typ.inner.name
}

func (typ Type) String() string {
	if typ.inner == nil {
		return "<Type=nil>"
	}
	if name := typ.Name(); name != "" {
		return fmt.Sprintf("<%s>", name)
	}
	return fmt.Sprintf("<Type=%p>", typ.inner)
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
