package bit

type Type interface {
	CppType() string
}

type IntType struct{}

func (IntType) CppType() string {
	return "int"
}

type StrType struct{}

func (StrType) CppType() string {
	return "const char *"
}

type NoneType struct{}

func (NoneType) CppType() string {
	return "void"
}

type InvalidType struct{}

func (InvalidType) CppType() string {
	return "!INVALID!"
}
