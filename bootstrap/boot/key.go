package boot

import (
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

type KeyValue interface {
	comparable
	KeyCompare
}

type Key struct {
	val KeyCompare
}

func KeyNone() Key {
	return Key{keyNone{}}
}

func KeyNew[T KeyValue](val T) Key {
	return Key{val}
}

func (key Key) Type() Type {
	if key.val == nil {
		return Type{}
	}
	return key.val.Type()
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

	panic(fmt.Sprintf("key failed to compare -- type %v", ta))
}

type keyNone struct{}

func (me keyNone) Cmp(other any) (bool, int) {
	_, ok := other.(keyNone)
	return ok, 0
}

func (me keyNone) Repr() string {
	return "Key(none)"
}

func (me keyNone) Type() Type {
	return TypeOf[keyNone]()
}
