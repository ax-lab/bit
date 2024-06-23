package code

type Type struct {
	typeImpl
}

func TypeNumber() Type {
	return Type{typeNumber{}}
}

func TypeStr() Type {
	return Type{typeStr{}}
}

type typeImpl interface {
	IsType()
	String() string
}
