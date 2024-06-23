package core_test

import (
	"testing"

	"axlab.dev/bit/core"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	test := require.New(t)
	test.NotEmpty(core.Version())
}
