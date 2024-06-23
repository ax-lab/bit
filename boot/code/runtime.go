package code

import "io"

type Runtime struct {
	StdErr io.Writer
	StdOut io.Writer
}
