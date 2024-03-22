package bot

type Value interface {
	Type() Type
}

type Int int

func (val Int) Type() Type {
	return TypeInt()
}

type Str string

func (val Str) Type() Type {
	return TypeStr()
}

type Bool bool

func (val Bool) Type() Type {
	return TypeBool()
}
