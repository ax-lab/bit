package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelloWorld(t *testing.T) {
	test := require.New(t)

	run := Run(test, `
		print 'hello world!!!'
	`)

	run.NoError()
	test.Equal("hello world!!!\n", run.StdOut)
}
