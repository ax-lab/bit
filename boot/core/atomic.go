package core

import "sync"

type Atomic[T comparable] struct {
	sync  sync.RWMutex
	value T
}

func AtomicNew[T comparable](value T) Atomic[T] {
	return Atomic[T]{value: value}
}

func (v *Atomic[T]) Store(value T) {
	v.sync.Lock()
	v.value = value
	v.sync.Unlock()
}

func (v *Atomic[T]) Load() (value T) {
	v.sync.RLock()
	value = v.value
	v.sync.RUnlock()
	return value
}

func (v *Atomic[T]) Swap(value T) (old T) {
	v.sync.Lock()
	old = v.value
	v.value = value
	v.sync.Unlock()
	return old
}

func (v *Atomic[T]) CompareAndSwap(old, new T) (swapped bool) {
	v.sync.Lock()
	if v.value == old {
		v.value = new
		swapped = true
	}
	v.sync.Unlock()
	return swapped
}
