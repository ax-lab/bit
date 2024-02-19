package code

func AddTypes(types ...Type) Type {
	if len(types) == 0 {
		return VoidType()
	}
	if len(types) == 1 {
		return types[0]
	}

	t0 := types[0]
	t1 := AddTypes(types[1:]...)
	if t0 == t1 {
		return t0
	} else {
		panic("TODO: sum type")
	}
}

func InvalidType() Type {
	panic("TODO: InvalidType")
}

func VoidType() Type {
	panic("TODO: VoidType")
}

type Type struct {
	*typeData
}

func (t Type) String() string {
	if t.typeData == nil {
		return "(nil)"
	}

	if t.Name != "" {
		return t.Name
	}
	return t.Impl.String()
}

func (t Type) CppType() string {
	if t.typeData == nil {
		return "!nil!"
	}
	return t.Impl.CppType()
}

func (t Type) CppPrint(ctx *CppContext, expr Expr) {
	t.Impl.CppPrint(ctx, expr)
}

func (t Type) CppPrintCondition(ctx *CppContext, expr Expr) string {
	return t.Impl.CppPrintCondition(ctx, expr)
}

// Implementation

type typeKind int

const (
	TypeNone typeKind = iota
	TypeInt
	TypeStr
	TypeBool
	TypeSum
	TypeNamed
	TypeError
)

type typeData struct {
	Kind typeKind
	Name string
	Impl typeImpl
}

type typeImpl interface {
	String() string
	CppType() string
	CppPrint(ctx *CppContext, expr Expr)
	CppPrintCondition(ctx *CppContext, expr Expr) string
}
