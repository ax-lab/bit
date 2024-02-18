package main_test

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/bit_core"
	"axlab.dev/bit/tester"
)

var (
	flagCpp     bool
	flagGlob    string
	flagVerbose bool
)

func init() {
	flag.StringVar(&flagGlob, "bit.glob", "*.test.bit", "pass a custom glob for `*.bit` tests")
	flag.BoolVar(&flagCpp, "bit.cpp", false, "enable to run bit tests compiling to C")
	flag.BoolVar(&flagVerbose, "bit.v", false, "enable verbose output in bit tests")
}

func TestBits(t *testing.T) {
	if flagVerbose {
		fmt.Printf("\n>>> [BIT] running tests with (cpp: %v, glob: %s)\n\n", flagCpp, flagGlob)
	}
	test := tester.NewRunner(t, "../tests", BitRunner{})
	test.SetGlob(flagGlob)
	test.Run()
}

type BitRunner struct{}

func (BitRunner) Run(input tester.Input) (out tester.Output) {
	compiler := bit.NewCompiler(context.Background(), "../tests", "../build/tests/")
	compiler.SetCore(bit_core.InitCompiler)

	stdOut := strings.Builder{}
	stdErr := strings.Builder{}
	options := bit.RunOptions{
		StdOut: &stdOut,
		StdErr: &stdErr,
	}
	if flagCpp {
		options.Cpp = true
	}

	res := compiler.Run(input.Name(), options)

	out.Error = res.Err
	out.StdOut = stdOut.String()
	out.StdErr = stdErr.String()
	out.Data = res.Value

	if flagCpp {
		out.IgnoreData = true
	}

	if len(res.Log) > 0 {
		err := strings.Builder{}
		err.WriteString("Compilation errors:\n")
		for _, it := range res.Log {
			err.WriteString("\n")
			err.WriteString(it.Error())
			err.WriteString("\n")
		}
		out.StdErr = err.String() + out.StdErr
	}

	return
}
