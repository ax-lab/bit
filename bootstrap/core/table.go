package core

import (
	"sync/atomic"
)

const (
	enableTableChecks = 1
)

type Table[T any] struct {
	shared  *tableSharedState
	changes *tableChangeLog[T]
	buffer  tableBuffer[T]
}

func NewTable[T any]() *Table[T] {
	shared := &tableSharedState{}
	shared.tableId = tableIdCounter.Add(1)
	table := &Table[T]{shared: shared}
	return table
}

func (table *Table[T]) ToList() (out []T) {
	return table.buffer.ToList()
}

func (table *Table[T]) Write() *TableWriter[T] {
	newTable := &Table[T]{
		shared:  table.shared,
		changes: table.changes,
		buffer:  tableBuffer[T]{},
	}
	newTable.buffer.data.Store(table.buffer.data.Load())
	out := &TableWriter[T]{source: newTable}
	out.changes.Store(out.source.changes)
	return out
}

func (table *Table[T]) Get(id Id) (out T) {
	out, _ = table.TryGet(id)
	return
}

func (table *Table[T]) TryGet(id Id) (out T, ok bool) {
	if enableTableChecks > 0 && id.table[0] != table.shared.tableId {
		panic("Table: trying to get an invalid id for the table")
	}

	if v := table.buffer.Get(id.toIndex()); v != nil {
		out, ok = *v, true
	}
	return
}

var tableIdCounter atomic.Uint64

type Id struct {
	table [enableTableChecks]uint64
	value uint64
}

func (id Id) Valid() bool {
	if enableTableChecks > 0 && id.table[0] == 0 {
		return false
	}
	return id.value > 0
}

func (id Id) toIndex() uint64 {
	return id.value - 1
}

type tableSharedState struct {
	tableId uint64
	nextId  atomic.Uint64
}

type tableChangeLog[T any] struct {
	item TableChange[T]
	prev *tableChangeLog[T]
}

type TableChange[T any] interface {
	Apply(writer *TableWriter[T])
	Merge(other TableChange[T], combined *[]TableChange[T]) error
}
