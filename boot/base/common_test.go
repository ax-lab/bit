package base_test

import (
	"testing"

	"axlab.dev/bit/base"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	test := require.New(t)
	test.NotEmpty(base.Version())
}
