package core

import "sync/atomic"

type TableWriter[T any] struct {
	done    atomic.Bool
	source  *Table[T]
	changes atomic.Pointer[tableChangeLog[T]]
}

func (writer *TableWriter[T]) Source() *Table[T] {
	return writer.source
}

func (writer *TableWriter[T]) Add(value T) (out Id) {
	if writer.done.Load() {
		panic("Table: writing after commit")
	}

	out.value = writer.source.shared.nextId.Add(1)
	if enableTableChecks > 0 {
		out.table[:][0] = writer.source.shared.tableId
	}

	writer.source.buffer.Set(out.toIndex(), &value)
	return
}

func (writer *TableWriter[T]) Apply(change TableChange[T]) {
	log := &tableChangeLog[T]{
		item: change,
		prev: writer.changes.Load(),
	}

	change.Apply(writer)
	if !writer.changes.CompareAndSwap(log.prev, log) {
		panic("TableWriter: concurrent writes are not allowed")
	}
}

func (writer *TableWriter[T]) Set(id Id, value T) {
	if enableTableChecks > 0 && id.table[:][0] != writer.source.shared.tableId {
		panic("Table: trying to set an invalid id for the table")
	}
	if writer.done.Load() {
		panic("Table: writing after commit")
	}

	writer.source.buffer.Set(id.toIndex(), &value)
}

func (writer *TableWriter[T]) Finish() *Table[T] {
	if !writer.done.CompareAndSwap(false, true) {
		panic("Table: writer has already been committed")
	}
	table := writer.source
	writer.source = nil
	table.changes = writer.changes.Load()
	writer.changes.Store(nil)
	return table
}
