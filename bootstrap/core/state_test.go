package core_test

import (
	"fmt"
	"math/rand"
	"testing"

	"axlab.dev/bit/core"
	"github.com/stretchr/testify/require"
)

func TestTypeId(t *testing.T) {
	test := require.New(t)

	type X int
	type Y = int

	a := core.TypeId[int]()
	b := core.TypeId[int32]()
	c := core.TypeId[int64]()
	d := core.TypeId[X]()

	test.True(a != b)
	test.True(a != c)
	test.True(a != d)
	test.True(b != c)
	test.True(b != d)
	test.True(c != d)

	test.Equal(a, core.TypeId[int]())
	test.Equal(b, core.TypeId[int32]())
	test.Equal(c, core.TypeId[int64]())
	test.Equal(d, core.TypeId[X]())
	test.Equal(a, core.TypeId[Y]())
}

func TestBasicState(t *testing.T) {
	test := require.New(t)

	t0 := core.NewState()

	xa := core.New[string]()
	xb := core.New[string]()
	xc := core.New[string]()

	test.Equal("", xa.Get(t0))
	test.Equal("", xb.Get(t0))
	test.Equal("", xc.Get(t0))

	xa.Set(t0, "A0")
	test.Equal("A0", xa.Get(t0))

	t1 := t0.Clone()
	t2 := t1.Clone()

	xa.Set(t0, "A1")
	xb.Set(t0, "B0")
	xc.Set(t0, "C0")
	test.Equal("A1", xa.Get(t0))
	test.Equal("B0", xb.Get(t0))
	test.Equal("C0", xc.Get(t0))

	test.Equal("A0", xa.Get(t1))
	test.Equal("A0", xa.Get(t2))

	test.Equal("", xb.Get(t1))
	test.Equal("", xc.Get(t1))

	test.Equal("", xb.Get(t2))
	test.Equal("", xc.Get(t2))

	xb.Set(t1, "B1")
	xb.Set(t2, "B2")

	xc.Set(t1, "C1")
	xc.Set(t2, "C2")

	test.Equal("A1", xa.Get(t0))
	test.Equal("B0", xb.Get(t0))
	test.Equal("C0", xc.Get(t0))

	test.Equal("A0", xa.Get(t1))
	test.Equal("B1", xb.Get(t1))
	test.Equal("C1", xc.Get(t1))

	test.Equal("A0", xa.Get(t2))
	test.Equal("B2", xb.Get(t2))
	test.Equal("C2", xc.Get(t2))
}

func TestLargeState(t *testing.T) {
	const (
		SizeInit = 7
		SizeGrow = 3
		SizeMod  = 0.3
		SizeMax  = 500_000
	)
	var (
		list []core.State
		size []int

		strList []core.Value[string]
		intList []core.Value[int]
		numList []core.Value[float64]
	)

	nextSize := 7

	for nextSize < SizeMax {
		var state core.State
		if len(list) > 0 {
			state = list[len(list)-1].Clone()
		} else {
			state = core.NewState()
		}
		list = append(list, state)

		for i := len(strList); i < nextSize; i++ {
			valStr := core.New[string]()
			valInt := core.New[int]()
			valNum := core.New[float64]()

			strList = append(strList, valStr)
			intList = append(intList, valInt)
			numList = append(numList, valNum)

			valStr.Set(state, fmt.Sprintf("item-%d", i))
			valInt.Set(state, i)
			valNum.Set(state, float64(i)+0.5)
		}

		gen := len(size)
		chk := rand.New(rand.NewSource(int64(nextSize)))
		for i := 0; i < nextSize; i++ {
			if chk.Float32() < SizeMod {
				valStr := strList[i]
				valInt := intList[i]
				valNum := numList[i]

				valStr.Set(state, fmt.Sprintf("item-%d-v%d", i, gen))
				valInt.Set(state, i+gen*1_000_000)
				valNum.Set(state, float64(i)+0.5+float64(gen*1_000_000))
			}
		}

		size = append(size, nextSize)
		nextSize *= 17
	}

	test := require.New(t)
	test.NotZero(len(list))
	test.Equal(len(list), len(size))
	test.Equal(size[len(size)-1], len(strList))
	test.Equal(size[len(size)-1], len(numList))
	test.Equal(size[len(size)-1], len(intList))

	var (
		expectedStr []string
		expectedInt []int
		expectedNum []float64
	)
	for i := range strList {
		expectedStr = append(expectedStr, fmt.Sprintf("item-%d", i))
		expectedInt = append(expectedInt, i)
		expectedNum = append(expectedNum, float64(i)+0.5)
	}

	for cur, curState := range list {
		chk := rand.New(rand.NewSource(int64(size[cur])))
		for i := 0; i < size[cur]; i++ {
			valStr := strList[i].Get(curState)
			valInt := intList[i].Get(curState)
			valNum := numList[i].Get(curState)

			if chk.Float32() < SizeMod {
				expectedStr[i] = fmt.Sprintf("item-%d-v%d", i, cur)
				expectedInt[i] = i + cur*1_000_000
				expectedNum[i] = float64(i) + 0.5 + float64(cur*1_000_000)
			}

			expStr := expectedStr[i]
			expInt := expectedInt[i]
			expNum := expectedNum[i]

			test.Equal(expStr, valStr, "gen %d, item %d: expected `%s`, got `%s`", cur, i, expStr, valStr)
			test.Equal(expInt, valInt, "gen %d, item %d: expected `%s`, got `%s`", cur, i, expInt, valInt)
			test.Equal(expNum, valNum, "gen %d, item %d: expected `%s`, got `%s`", cur, i, expNum, valNum)
		}
	}
}
