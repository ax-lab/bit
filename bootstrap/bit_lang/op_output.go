package bit_lang

import "axlab.dev/bit/bit"

type Output struct{}

func (op Output) IsSame(other bit.Binding) bool {
	if v, ok := other.(Output); ok {
		return v == op
	}
	return false
}

func (op Output) Precedence() bit.Precedence {
	return bit.PrecOutput
}

func (op Output) Process(args *bit.BindArgs) {
	// only flag nodes that can be output as-is as done
}

func (op Output) String() string {
	return "Output"
}
