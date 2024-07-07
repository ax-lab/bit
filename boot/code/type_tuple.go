package code

import "strings"

type TypeTuple struct {
	types []Type
}

func (tuple TypeTuple) TypeDef() TypeDef { return tuple }

func (tuple TypeTuple) Len() int {
	return len(tuple.types)
}

func (tuple TypeTuple) Get(nth int) Type {
	return tuple.types[nth]
}

func (tuple TypeTuple) String() string {
	out := strings.Builder{}
	out.WriteString("(")
	for n, it := range tuple.types {
		if n > 0 {
			out.WriteString(", ")
		}
		out.WriteString(it.String())
	}
	out.WriteString(")")
	return out.String()
}
