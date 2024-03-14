package core

import "sync/atomic"

var (
	idCounter atomic.Uint64
)

func NewData[T any]() Data[T] {
	return Data[T]{id: idCounter.Add(1)}
}

func NewList[T any]() List[T] {
	return List[T]{id: idCounter.Add(1)}
}

type Data[T any] struct {
	_  [0]*T
	id uint64
}

func (data Data[T]) IsZero() bool {
	return data.id == 0
}

func (data Data[T]) Get(v IsVersion) T {
	panic("TODO")
}

func (data Data[T]) Set(v NewVersion, value T) {
	panic("TODO")
}

type List[T any] struct {
	_  [0]*T
	id uint64
}

func (ls List[T]) IsZero() bool {
	return ls.id == 0
}

type IsVersion interface {
	hasVersion()
}

type Version struct{}

func (v Version) Write() NewVersion {
	panic("TODO")
}

func (v Version) Get() {}

type NewVersion struct{}

func (v NewVersion) Commit() Version {
	panic("TODO")
}
