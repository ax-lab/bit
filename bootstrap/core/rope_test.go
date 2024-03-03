package core_test

import (
	"testing"

	"axlab.dev/bit/core"
	"github.com/stretchr/testify/require"
)

func TestRope(t *testing.T) {
	test := require.New(t)

	ls := core.Rope[int]{}
	ls = ls.Insert(0, 1, 2, 3)
	ls = ls.Insert(3, 4, 5, 6)

	for i := 0; i < 6; i++ {
		test.Equal(i+1, ls.Get(i))
	}
}
