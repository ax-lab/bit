package code

import (
	"runtime"
	"sync/atomic"
)

const MaxLine int = 60

type Repr string

const (
	ReprDebug Repr = "debug"
	ReprLabel Repr = "label"
	ReprLine  Repr = "line"
)

var globalId atomic.Int64

type Id struct {
	val atomic.Int64
}

func (id *Id) GetId() int64 {
	for {
		if id := id.val.Load(); id > 0 {
			return id
		}

		if id.val.CompareAndSwap(0, -1) {
			newId := globalId.Add(1)
			id.val.Store(newId)
			return newId
		}

		runtime.Gosched()
	}
}

type Item interface {
	GetId() int64
	OutputCpp(ctx *CppContext)
	Repr(mode Repr) string
}

type Expr interface {
	Item
	Type() Type
	Eval(rt *Runtime) (Value, error)
}

type Stmt interface {
	Item
	Exec(rt *Runtime) error
}

func init() {
	assertExpr[*Int]()
	assertExpr[*Bool]()
	assertExpr[*Str]()
	assertExpr[*Variable]()

	assertStmt[*SetVar]()
	assertStmt[*Block]()
	assertStmt[*If]()
	assertStmt[*Print]()
}

func assertExpr[T Expr]() {}
func assertStmt[T Stmt]() {}
