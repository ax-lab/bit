package boot_test

import (
	"fmt"
	"strings"
	"testing"

	"axlab.dev/bit/boot"
	"github.com/stretchr/testify/require"
)

func TestBasicRun(t *testing.T) {
	test := require.New(t)
	state := boot.State{}

	typ := boot.TypeOf[int]()
	key := boot.KeyFrom(42)

	state.AddSource("src0", "123")
	state.AddSource("src1", "123456")

	var evalLog []string

	cnt := 0
	val := boot.BindFunc(boot.BindOrder(1), func(args boot.BindArgs) {
		if len(args.List) == 0 {
			panic("argument list is empty")
		}

		head := fmt.Sprintf("%d: %v -> %v", cnt, args.Type, args.Key)
		cnt++

		for _, it := range args.List {
			var pos []string
			for i := 0; i < it.Len(); i++ {
				span := it.Get(i)
				pos = append(pos, span.RangeString())
			}

			log := fmt.Sprintf("%s: %s %s", head, it.Src().Name(), strings.Join(pos, " "))
			evalLog = append(evalLog, log)
		}

	})

	state.AddSource("src2", "1234")

	state.Define(typ, key, val)
	state.Evaluate()

	test.Empty(state.Errors())
	test.Equal([]string{
		"0: Type(int) -> Key(42): src0 0…MAX",
		"0: Type(int) -> Key(42): src1 0…MAX",
		"0: Type(int) -> Key(42): src2 0…MAX",
	}, evalLog)
}
