package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelloWorld(t *testing.T) {
	test := require.New(t)

	run := Run("hello-world", test, `
		print 'hello world!!!'
	`)

	run.NoError()
	test.Equal("hello world!!!\n", run.StdOut())
}

func TestHelloWorldC(t *testing.T) {
	test := require.New(t)

	run := RunC("hello-world", test, `
		print 'hello world!!!'
	`)

	run.NoError()
	test.Equal("hello world!!!\n", run.StdOut())
}

func TestHelloWorldGo(t *testing.T) {
	test := require.New(t)

	run := RunGo("hello-world", test, `
		print 'hello world!!!'
	`)

	run.NoError()
	test.Equal("hello world!!!\n", run.StdOut())
}
