package boot

import (
	"cmp"
	"fmt"
)

type WithKey interface {
	Key() Key
}

type WithKeys interface {
	Keys() []Key
}

type KeyCompare interface {
	Value
	Cmp(other any) (bool, int)
}

type Key struct {
	val KeyCompare
	typ Type
}

func KeyNone() Key {
	return Key{nil, Type{}}
}

func KeyValue[T KeyCompare](val T) Key {
	return Key{val, TypeOf[T]()}
}

func KeyFrom[T cmp.Ordered](val T) Key {
	return Key{keyOrdered[T]{val}, TypeOf[T]()}
}

func (key Key) Type() Type {
	return key.typ
}

func (key Key) Value() Value {
	return key.val
}

func (key Key) Cmp(other Key) int {
	if key.val == nil {
		if other.val == nil {
			return 0
		} else {
			return -1
		}
	}
	if ok, res := key.val.Cmp(other.val); ok {
		return res
	}

	ta, tb := key.Type(), other.Type()
	if ta != tb {
		return ta.Cmp(tb)
	}

	panic(fmt.Sprintf("key failed to compare -- type %v and %v", ta, tb))
}

func (key Key) String() string {
	if key.val == nil {
		return "Key(nil)"
	}
	return fmt.Sprintf("Key(%v)", key.val.Repr())
}

type keyOrdered[T cmp.Ordered] struct {
	val T
}

func (key keyOrdered[T]) Cmp(other any) (bool, int) {
	if kv, ok := other.(keyOrdered[T]); ok {
		return true, cmp.Compare(key.val, kv.val)
	}
	return false, 0
}

func (key keyOrdered[T]) Repr() string {
	return fmt.Sprint(key.val)
}
