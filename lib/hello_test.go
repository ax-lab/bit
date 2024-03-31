package lib_test

import (
	"testing"

	"axlab.dev/test/lib"
	"github.com/stretchr/testify/require"
)

func TestAnswer(t *testing.T) {
	test := require.New(t)
	test.Equal(42, lib.Answer())
}
