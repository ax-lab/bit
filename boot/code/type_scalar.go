package code

import "fmt"

type TypeScalarKind int

const (
	TypeScalarUnit TypeScalarKind = iota
	TypeScalarBool
	TypeScalarFloat
	TypeScalarInt
	TypeScalarNumber
	TypeScalarString
)

type TypeScalar struct {
	kind TypeScalarKind
}

func (scalar TypeScalar) Kind() TypeScalarKind {
	return scalar.kind
}

func (scalar TypeScalar) TypeDef() TypeDef { return scalar }

func (scalar TypeScalar) String() string {
	switch scalar.kind {
	case TypeScalarUnit:
		return "Unit"
	case TypeScalarBool:
		return "Bool"
	case TypeScalarFloat:
		return "Float"
	case TypeScalarInt:
		return "Int"
	case TypeScalarNumber:
		return "Number"
	case TypeScalarString:
		return "String"
	default:
		panic(fmt.Sprintf("invalid TypeScalar kind: %#v", scalar.kind))
	}
}
