package base_test

import (
	"testing"

	"axlab.dev/bit/base"
	"github.com/stretchr/testify/require"
)

func TestErrorSet(t *testing.T) {
	test := require.New(t)

	a := base.Error("error A")
	b := base.Error("error %s", "B")

	errA := base.Errors(a, nil)
	errB := base.Errors(nil, b)
	errC := base.Errors(a, b)

	test.Equal(a, errA)
	test.Equal(b, errB)
	test.ErrorIs(errC, a)
	test.ErrorIs(errC, b)

	msg := errC.Error()
	test.Contains(msg, "error A")
	test.Contains(msg, "error B")

	errD := base.Errors(base.Error("some error"), errC)
	test.ErrorIs(errD, a)
	test.ErrorIs(errD, b)
	test.Contains(errD.Error(), "some error")
}
