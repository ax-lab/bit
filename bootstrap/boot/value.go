package boot

type Value interface {
	Repr() string
	Type() Type
}
