package code

func TypeNumber() Type {
	return Type{typeNumber{}}
}

func TypeStr() Type {
	return Type{typeStr{}}
}

type Type struct {
	typeImpl
}

type typeImpl interface {
	IsType()
	String() string
}

func (typ Type) IsAssignable(other Type) bool {
	panic("TODO")
}
