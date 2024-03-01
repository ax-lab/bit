package core_test

import (
	"fmt"
	"math/rand"
	"testing"

	"axlab.dev/bit/core"
	"github.com/stretchr/testify/require"
)

func TestBasicTable(t *testing.T) {
	test := require.New(t)

	t0 := core.NewTable[string]()

	w1 := t0.Write()
	xa := w1.Add("A0")
	t1 := w1.Finish()

	w2a := t1.Write()
	w2b := t1.Write()

	xb := w2a.Add("B0")
	xc := w2b.Add("C0")

	w2a.Set(xa, "A0-a")
	w2b.Set(xa, "A0-b")

	t2a := w2a.Finish()
	t2b := w2b.Finish()

	test.Equal(0, len(t0.ToList()))
	test.Equal([]string{"A0"}, t1.ToList())
	test.Equal([]string{"A0-a", "B0"}, t2a.ToList())
	test.Equal([]string{"A0-b", "", "C0"}, t2b.ToList())

	test.Equal("", t0.Get(xa))
	test.Equal("", t0.Get(xb))
	test.Equal("", t0.Get(xc))

	test.Equal("A0", t1.Get(xa))
	test.Equal("", t1.Get(xb))
	test.Equal("", t1.Get(xc))

	test.Equal("A0-a", t2a.Get(xa))
	test.Equal("B0", t2a.Get(xb))
	test.Equal("", t2a.Get(xc))

	test.Equal("A0-b", t2b.Get(xa))
	test.Equal("", t2b.Get(xb))
	test.Equal("C0", t2b.Get(xc))
}

func TestLargeTable(t *testing.T) {
	const (
		SizeInit = 7
		SizeGrow = 3
		SizeMod  = 0.3
		SizeMax  = 500_000
	)
	var (
		list []*core.Table[string]
		size []int
		name []core.Id
	)

	nextSize := 7

	tb := core.NewTable[string]()
	for nextSize < SizeMax {
		gen := len(size)
		size = append(size, nextSize)

		w := tb.Write()
		for i := len(name); i < nextSize; i++ {
			id := w.Add(fmt.Sprintf("item-%d", i))
			name = append(name, id)
		}

		chk := rand.New(rand.NewSource(int64(nextSize)))
		for i := 0; i < nextSize; i++ {
			if chk.Float32() < SizeMod {
				w.Set(name[i], fmt.Sprintf("item-%d-v%d", i, gen))
			}
		}

		tb = w.Finish()
		list = append(list, tb)
		nextSize *= 17
	}

	test := require.New(t)
	test.NotZero(len(list))
	test.Equal(len(list), len(size))
	test.NotZero(len(name))

	expected := make([]string, len(name))
	for i := range expected {
		expected[i] = fmt.Sprintf("item-%d", i)
	}

	for gen, tb := range list {
		all := tb.ToList()
		test.Equal(size[gen], len(all))

		chk := rand.New(rand.NewSource(int64(size[gen])))
		for i, it := range all {
			if chk.Float32() < SizeMod {
				expected[i] = fmt.Sprintf("item-%d-v%d", i, gen)
			}
			test.Equal(expected[i], it, "gen %d, item %d: expected `%s`, got `%s`", gen, i, expected[i], it)
		}
	}
}
