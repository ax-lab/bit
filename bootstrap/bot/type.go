package bot

import "fmt"

type Type interface {
	Name() string
	Print(arg Value, out RuntimeWriter) error
}

func TypeBool() Type {
	return typeBool{}
}

func TypeInt() Type {
	return typeInt{}
}

func TypeStr() Type {
	return typeStr{}
}

type typeBool struct{}

func (typeBool) Name() string {
	return "bool"
}

func (typeBool) Print(arg Value, out RuntimeWriter) error {
	val := arg.(Bool)
	if val {
		return out.Write("true")
	} else {
		return out.Write("false")
	}
}

type typeInt struct{}

func (typeInt) Name() string {
	return "int"
}

func (typeInt) Print(arg Value, out RuntimeWriter) error {
	val := arg.(Int)
	return out.Write(fmt.Sprint(val))
}

type typeStr struct{}

func (typeStr) Name() string {
	return "str"
}

func (typeStr) Print(arg Value, out RuntimeWriter) error {
	val := arg.(Str)
	return out.Write(fmt.Sprint(val))
}
